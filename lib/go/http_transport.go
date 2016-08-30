package frugal

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

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
func NewFrugalHandlerFunc(processor FProcessor, inPfactory, outPfactory *FProtocolFactory) http.HandlerFunc {

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

		// Create a decoder based on the payload
		decoder := base64.NewDecoder(base64.StdEncoding, r.Body)

		// Read out the frame size
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
		if err := processor.Process(inPfactory.GetProtocol(input), outPfactory.GetProtocol(output)); err != nil {
			http.Error(w,
				fmt.Sprintf("Frugal request failed %s", err),
				http.StatusBadRequest,
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
		encoded := new(bytes.Buffer)
		encoder := newEncoder(encoded)
		var err error
		binary.BigEndian.PutUint32(frameSize, uint32(outBuf.Len()))
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

// FHttpTransportBuilder configures and builds HTTP FTransport instances.
type FHttpTransportBuilder struct {
	client            *http.Client
	url               string
	requestSizeLimit  uint
	responseSizeLimit uint
}

// NewFHttpTransportBuilder creates a builder which configures and builds HTTP
// FTransport instances.
func NewFHttpTransportBuilder(client *http.Client, url string) *FHttpTransportBuilder {
	return &FHttpTransportBuilder{
		client: client,
		url:    url,
	}
}

// WithRequestSizeLimit adds a request size limit. If set to 0 (the default),
// there is no size limit on requests.
func (h *FHttpTransportBuilder) WithRequestSizeLimit(requestSizeLimit uint) *FHttpTransportBuilder {
	h.requestSizeLimit = requestSizeLimit
	return h
}

// WithResponseSizeLimit adds a response size limit. If set to 0 (the default),
// there is no size limit on responses.
func (h *FHttpTransportBuilder) WithResponseSizeLimit(responseSizeLimit uint) *FHttpTransportBuilder {
	h.responseSizeLimit = responseSizeLimit
	return h
}

// Build a new configured HTTP FTransport.
func (h *FHttpTransportBuilder) Build() FTransport {
	return &fHttpTransport{
		fBaseTransport:    newFBaseTransport(h.requestSizeLimit),
		client:            h.client,
		url:               h.url,
		responseSizeLimit: h.responseSizeLimit,
	}
}

// fHttpTransport implements FTransport. This is a "stateless"
// transport in the sense that this transport is not persistently connected to
// a single server. A request is simply an http request and a response is an
// http response. This assumes requests/responses fit within a single http
// request.
type fHttpTransport struct {
	*fBaseTransport
	client            *http.Client
	url               string
	responseSizeLimit uint
	isOpen            bool
}

// Open initializes the transport for use.
func (h *fHttpTransport) Open() error {
	// no-op
	return nil
}

// IsOpen returns true if the transport is open for use.
func (h *fHttpTransport) IsOpen() bool {
	// it's always open
	return true
}

// Close closes the transport.
func (h *fHttpTransport) Close() error {
	// no-op
	return nil
}

// Read should not be called, it will return an error
func (h *fHttpTransport) Read(buf []byte) (int, error) {
	return 0, errors.New("Cannot read on FTransport")
}

// Flush sends the buffered bytes over HTTP.
func (h *fHttpTransport) Flush() error {
	if !h.IsOpen() {
		return h.getClosedConditionError("flush:")
	}
	data := h.GetWriteBytes()
	if len(data) == 0 {
		return nil
	}

	h.ResetWriteBuffer()
	data = prependFrameSize(data)

	// Make the HTTP request
	response, err := h.makeRequest(data)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}

	// All responses should be framed with 4 bytes (uint32)
	if len(response) < 4 {
		return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
			errors.New("frugal: invalid frame size"))
	}

	// If there are only 4 bytes, this needs to be a one-way
	// (i.e. frame size 0)
	if len(response) == 4 {
		if binary.BigEndian.Uint32(response) != 0 {
			return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
				errors.New("frugal: missing data"))
		}
		// it's a one-way, drop it
		return nil
	}

	return thrift.NewTTransportExceptionFromError(h.fBaseTransport.ExecuteFrame(response))
}

// This is a no-op for fHttpTransport
func (h *fHttpTransport) SetMonitor(monitor FTransportMonitor) {
}

func (h *fHttpTransport) makeRequest(requestPayload []byte) ([]byte, error) {
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
	request, err := http.NewRequest("POST", h.url, encoded)
	if err != nil {
		return nil, err
	}

	// Add request headers
	request.Header.Add(contentTypeHeader, frugalContentType)
	request.Header.Add(acceptHeader, frugalContentType)
	request.Header.Add(contentTransferEncodingHeader, base64Encoding)
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
		return nil, thrift.NewTTransportException(RESPONSE_TOO_LARGE,
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
		return nil, fmt.Errorf("response errored with code %d and message %s",
			response.StatusCode, body)
	}

	// Decode and return response body
	bts, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}
	return bts, nil

}

func (h *fHttpTransport) getClosedConditionError(prefix string) error {
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s HTTP TTransport not open", prefix))
}
