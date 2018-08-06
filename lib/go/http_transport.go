/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package frugal

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const (
	payloadLimitHeader            = "x-frugal-payload-limit"
	acceptHeader                  = "accept"
	contentTypeHeader             = "content-type"
	contentTransferEncodingHeader = "content-transfer-encoding"

	frugalContentType = "application/x-frugal"
	base64Encoding    = "base64"
)

var newEncoder = func(buf *bytes.Buffer) io.WriteCloser {
	return base64.NewEncoder(base64.StdEncoding, buf)
}

// NewFrugalHandlerFunc is a function that creates a ready to use Frugal handler
// function.
func NewFrugalHandlerFunc(processor FProcessor, protocolFactory *FProtocolFactory) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(contentTypeHeader, frugalContentType)

		// Check for size limitation
		limitStr := r.Header.Get(payloadLimitHeader)
		var limit int64
		if limitStr != "" {
			var err error
			limit, err = strconv.ParseInt(limitStr, 10, 64)
			if err != nil {
				http.Error(w,
					fmt.Sprintf("%s header not an integer", payloadLimitHeader),
					http.StatusBadRequest,
				)
				return
			}
		}

		// Need 4 bytes for the frame size, at a minimum.
		if r.ContentLength < 4 {
			http.Error(w, fmt.Sprintf("Invalid request size %d", r.ContentLength), http.StatusBadRequest)
			return
		}

		// Create a decoder based on the payload
		decoder := base64.NewDecoder(base64.StdEncoding, r.Body)

		// Read out the frame size
		// TODO: should we do something with the frame size?
		frameSize := make([]byte, 4)
		if _, err := io.ReadFull(decoder, frameSize); err != nil {
			http.Error(w,
				fmt.Sprintf("Could not read the frugal frame bytes %s", err),
				http.StatusBadRequest,
			)
			return
		}

		// Read and process frame
		input := thrift.NewStreamTransportR(decoder)
		outBuf := new(bytes.Buffer)
		output := &thrift.TMemoryBuffer{Buffer: outBuf}
		iprot := protocolFactory.GetProtocol(input)
		oprot := protocolFactory.GetProtocol(output)
		if err := processor.Process(iprot, oprot); err != nil {
			http.Error(w,
				fmt.Sprintf("Error processing request: %s", err),
				http.StatusInternalServerError,
			)
			return
		}

		// If client requested a limit, check the buffer size
		if limit > 0 && outBuf.Len() > int(limit) {
			http.Error(w,
				fmt.Sprintf("Response size (%d) larger than requested size (%d)", outBuf.Len(), limit),
				http.StatusRequestEntityTooLarge,
			)
			return
		}

		// Encode response
		var (
			encoded = new(bytes.Buffer)
			encoder = newEncoder(encoded)
			err     error
		)
		bigEndianPutUint32(frameSize, uint32(outBuf.Len()))
		if _, e := encoder.Write(frameSize); e != nil {
			err = e
		}
		if _, e := encoder.Write(outBuf.Bytes()); e != nil {
			err = e
		}
		if e := encoder.Close(); e != nil {
			err = e
		}

		// Check for encoding errors
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Problem encoding frugal bytes to base64 %s", err),
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Add(contentTransferEncodingHeader, base64Encoding)
		w.Write(encoded.Bytes())
	}
}

type GetHeadersWithContext func(FContext) map[string]string

// FHTTPTransportBuilder configures and builds HTTP FTransport instances.
type FHTTPTransportBuilder struct {
	client            *http.Client
	url               string
	requestSizeLimit  uint
	responseSizeLimit uint
	requestHeaders    map[string]string
	getRequestHeaders GetHeadersWithContext
}

// NewFHTTPTransportBuilder creates a builder which configures and builds HTTP
// FTransport instances.
func NewFHTTPTransportBuilder(client *http.Client, url string) *FHTTPTransportBuilder {
	return &FHTTPTransportBuilder{
		client: client,
		url:    url,
	}
}

// WithRequestSizeLimit adds a request size limit. If set to 0 (the default),
// there is no size limit on requests.
func (h *FHTTPTransportBuilder) WithRequestSizeLimit(requestSizeLimit uint) *FHTTPTransportBuilder {
	h.requestSizeLimit = requestSizeLimit
	return h
}

// WithResponseSizeLimit adds a response size limit. If set to 0 (the default),
// there is no size limit on responses.
func (h *FHTTPTransportBuilder) WithResponseSizeLimit(responseSizeLimit uint) *FHTTPTransportBuilder {
	h.responseSizeLimit = responseSizeLimit
	return h
}

// withRequestHeaders adds custom request headers. If set to nil (the default),
// there is no size limit on responses.
func (h *FHTTPTransportBuilder) WithRequestHeaders(requestHeaders map[string]string) *FHTTPTransportBuilder {
	h.requestHeaders = requestHeaders
	return h
}

// withRequestHeadersFromFContext adds custom request headers to each request
// with a provided function that accepts an FContext and returns map of
// string key-value pairs
func (h *FHTTPTransportBuilder) WithRequestHeadersFromFContext(getRequestHeaders GetHeadersWithContext) *FHTTPTransportBuilder {
	h.getRequestHeaders = getRequestHeaders
	return h
}

// Build a new configured HTTP FTransport.
func (h *FHTTPTransportBuilder) Build() FTransport {
	return &fHTTPTransport{
		fBaseTransport:    newFBaseTransport(h.requestSizeLimit),
		client:            h.client,
		url:               h.url,
		responseSizeLimit: h.responseSizeLimit,
		requestHeaders:    h.requestHeaders,
		getRequestHeaders: h.getRequestHeaders,
	}
}

// fHTTPTransport implements FTransport. This is a "stateless"
// transport in the sense that this transport is not persistently connected to
// a single server. A request is simply an http request and a response is an
// http response. This assumes requests/responses fit within a single http
// request.
type fHTTPTransport struct {
	*fBaseTransport
	client            *http.Client
	url               string
	responseSizeLimit uint
	isOpen            bool
	requestHeaders    map[string]string
	getRequestHeaders GetHeadersWithContext
}

// Open initializes the transport for use.
func (h *fHTTPTransport) Open() error {
	// no-op
	return nil
}

// IsOpen returns true if the transport is open for use.
func (h *fHTTPTransport) IsOpen() bool {
	// it's always open
	return true
}

// Close closes the transport.
func (h *fHTTPTransport) Close() error {
	// no-op
	return nil
}

// Oneway transmits the given data and doesn't wait for a response.
// Implementations of oneway should be threadsafe and respect the timeout
// present on the context.
func (h *fHTTPTransport) Oneway(ctx FContext, data []byte) error {
	_, err := h.Request(ctx, data)
	return err
}

// Request transmits the given data and waits for a response.
// Implementations of request should be threadsafe and respect the timeout
// present the on context. The data is expected to already be framed.
func (h *fHTTPTransport) Request(ctx FContext, data []byte) (thrift.TTransport, error) {
	if !h.IsOpen() {
		return nil, h.getClosedConditionError("request:")
	}

	if len(data) == 4 {
		return nil, nil
	}

	if h.requestSizeLimit > 0 && len(data) > int(h.requestSizeLimit) {
		return nil, thrift.NewTTransportException(
			TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE,
			fmt.Sprintf("Message exceeds %d bytes, was %d bytes", h.requestSizeLimit, len(data)))
	}

	// Make the HTTP request
	response, err := h.makeRequest(ctx, data)
	if err != nil {
		if strings.HasSuffix(err.Error(), "net/http: request canceled") ||
			strings.HasSuffix(err.Error(), "net/http: timeout awaiting response headers") ||
			strings.HasSuffix(err.Error(), "net/http: request canceled while waiting for connection") {
			return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_TIMED_OUT, "frugal: http request timed out")
		}
		return nil, thrift.NewTTransportExceptionFromError(err)
	}

	// All responses should be framed with 4 bytes (uint32)
	if len(response) < 4 {
		return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
			errors.New("frugal: invalid frame size"))
	}

	// If there are only 4 bytes, this needs to be a one-way
	// (i.e. frame size 0)
	if len(response) == 4 {
		if bigEndianUint32(response) != 0 {
			return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
				errors.New("frugal: missing data"))
		}
		// it's a one-way, drop it
		return nil, nil
	}

	return &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(response[4:])}, nil
}

// GetRequestSizeLimit returns the maximum number of bytes that can be
// transmitted. Returns a non-positive number to indicate an unbounded
// allowable size.
func (h *fHTTPTransport) GetRequestSizeLimit() uint {
	return h.requestSizeLimit
}

// This is a no-op for fHTTPTransport
func (h *fHTTPTransport) SetMonitor(monitor FTransportMonitor) {
}

func (h *fHTTPTransport) makeRequest(fCtx FContext, requestPayload []byte) ([]byte, error) {
	// Encode request payload
	encoded := new(bytes.Buffer)
	encoder := newEncoder(encoded)
	if _, err := encoder.Write(requestPayload); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	// Initialize request
	ctx, cancel := context.WithTimeout(context.Background(), fCtx.Timeout())
	defer cancel()
	request, err := http.NewRequest("POST", h.url, encoded)
	if err != nil {
		return nil, err
	}
	request = request.WithContext(ctx)

	// add user supplied headers first, to avoid monkeying
	// with the size limits headers below.
	// add dynamic headers from fcontext first
	if h.getRequestHeaders != nil {
		for key, value := range h.getRequestHeaders(fCtx) {
			request.Header.Set(key, value)
		}
	}
	// now add manually passed in request headers
	if h.requestHeaders != nil {
		for key, value := range h.requestHeaders {
			request.Header.Set(key, value)
		}
	}

	// Add request headers
	request.Header.Set(contentTypeHeader, frugalContentType)
	request.Header.Set(acceptHeader, frugalContentType)
	request.Header.Set(contentTransferEncodingHeader, base64Encoding)
	if h.responseSizeLimit > 0 {
		request.Header.Add(payloadLimitHeader, strconv.FormatUint(uint64(h.responseSizeLimit), 10))
	}

	// Make request
	response, err := h.client.Do(request)
	if err != nil {
		return nil, err
	}

	// Response too large
	if response.StatusCode == http.StatusRequestEntityTooLarge {
		return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE,
			"response was too large for the transport")
	}

	// Decode body
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(response.Body); err != nil {
		return nil, err
	}
	if err := response.Body.Close(); err != nil {
		return nil, err
	}
	body := string(buf.Bytes())

	// Check bad status code
	if response.StatusCode >= 300 {
		return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
			fmt.Sprintf("response errored with code %d and message %s",
				response.StatusCode, body))
	}

	// Decode and return response body
	bts, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}
	return bts, nil

}

func (h *fHTTPTransport) getClosedConditionError(prefix string) error {
	return thrift.NewTTransportException(TRANSPORT_EXCEPTION_NOT_OPEN,
		fmt.Sprintf("%s HTTP TTransport not open", prefix))
}
