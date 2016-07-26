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
	"sync"

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

// NewFrugalHandlerFunc is a function that create a ready to use Frugal handler
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

// HttpFTransportBuilder configures and builds HTTP FTransport instances.
type HttpFTransportBuilder struct {
	client            *http.Client
	url               string
	requestSizeLimit  uint
	responseSizeLimit uint
}

// NewHttpFTransportBuilder creates a builder which configures and builds HTTP
// FTransport instances.
func NewHttpFTransportBuilder(client *http.Client, url string) *HttpFTransportBuilder {
	return &HttpFTransportBuilder{
		client: client,
		url:    url,
	}
}

// WithRequestSizeLimit adds a request size limit. If set to 0 (the default),
// there is no size limit on requests.
func (h *HttpFTransportBuilder) WithRequestSizeLimit(requestSizeLimit uint) *HttpFTransportBuilder {
	h.requestSizeLimit = requestSizeLimit
	return h
}

// WithResponseSizeLimit adds a response size limit. If set to 0 (the default),
// there is no size limit on responses.
func (h *HttpFTransportBuilder) WithResponseSizeLimit(responseSizeLimit uint) *HttpFTransportBuilder {
	h.responseSizeLimit = responseSizeLimit
	return h
}

// Build a new configured HTTP FTransport.
func (h *HttpFTransportBuilder) Build() FTransport {
	return &httpFTransport{
		fBaseTransport:    newFBaseTransport(h.requestSizeLimit),
		client:            h.client,
		url:               h.url,
		responseSizeLimit: h.responseSizeLimit,
	}
}

// NewHttpTTransport returns a new Thrift TTransport which uses the
// HTTP as the underlying transport. This TTransport is stateless in that
// there is no connection maintained between the client and server. A request
// is simply an http request and a response is an http response. This assumes
// requests/responses fit within a single http request.
// DEPRECATED - Use HttpFTransportBuilder to create an FTransport directly.
// TODO: Remove this with 2.0
func NewHttpTTransport(client *http.Client, url string) thrift.TTransport {
	return NewHttpTTransportWithLimits(client, url, 0, 0)
}

// NewHttpTTransportWithLimits returns a new Thrift TTransport which uses the
// HTTP as the underlying transport. This TTransport is stateless in that
// there is no connection maintained between the client and server. A request
// is simply an http request and a response is an http response. This assumes
// requests/responses fit within a single http request. The size limits for
// request/response data may be set with requestSizeLimit and
// responseSizeLimit, respectively. Setting to 0 implies no limit.
// DEPRECATED - Use HttpFTransportBuilder to create an FTransport directly.
// TODO: Remove this with 2.0
func NewHttpTTransportWithLimits(client *http.Client, url string,
	requestSizeLimit uint, responseSizeLimit uint) thrift.TTransport {
	return &httpFTransport{
		fBaseTransport:    newFBaseTransportForTTransport(requestSizeLimit, frameBufferSize),
		client:            client,
		url:               url,
		responseSizeLimit: responseSizeLimit,
		isTTransport:      true,
	}
}

// httpFTransport implements thrift.TTransport. This is a "stateless"
// transport in the sense that this transport is not persistently connected to
// a single server. A request is simply an http request and a response is an
// http response. This assumes requests/responses fit within a single http
// request.
type httpFTransport struct {
	*fBaseTransport
	client            *http.Client
	url               string
	responseSizeLimit uint
	isOpen            bool

	// TODO: Remove with 2.0
	isTTransport bool
	mu           sync.RWMutex
}

// Open initializes the close channel and sets the open flag to true.
func (h *httpFTransport) Open() error {
	// TODO: Open can be a no-op with 2.0
	h.mu.Lock()
	h.isOpen = true
	h.fBaseTransport.Open()
	h.mu.Unlock()
	return nil
}

func (h *httpFTransport) IsOpen() bool {
	// TODO: Remove locking with 2.0
	h.mu.RLock()
	defer h.mu.RUnlock()

	// TODO: This should always return true with 2.0
	return h.isOpen
}

// Close closes the close channel and sets the open flag to false.
func (h *httpFTransport) Close() error {
	// TODO: This should be a no-op with 2.0
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isOpen {
		return nil
	}
	h.isOpen = false
	h.fBaseTransport.Close(nil)
	return nil
}

// Read up to len(buf) bytes into buf.
// TODO: This should just return an error with 2.0
func (h *httpFTransport) Read(buf []byte) (int, error) {
	if !h.isTTransport {
		return 0, errors.New("Cannot read on FTransport")
	}

	// TODO: Remove all read logic with 2.0
	if !h.IsOpen() {
		return 0, h.getClosedConditionError("read:")
	}
	return h.fBaseTransport.Read(buf)
}

// Write the bytes to a buffer. Returns ErrTooLarge if the buffer exceeds the
// client specified request size limit.
func (h *httpFTransport) Write(buf []byte) (int, error) {
	if !h.IsOpen() {
		return 0, h.getClosedConditionError("write:")
	}
	return h.fBaseTransport.Write(buf)
}

// Flush sends the buffered bytes over HTTP.
func (h *httpFTransport) Flush() error {
	if !h.IsOpen() {
		return h.getClosedConditionError("flush:")
	}
	data := h.GetWriteBytes()
	if len(data) == 0 {
		return nil
	}

	h.ResetWriteBuffer()
	// TODO: Remove this check in 2.0
	if !h.isTTransport {
		data = prependFrameSize(data)
	}

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

	// TODO: Remove this with 2.0
	if h.isTTransport {
		select {
		case h.frameBuffer <- response:
		case <-h.ClosedChannel():
		}
		return nil
	}

	return thrift.NewTTransportExceptionFromError(h.fBaseTransport.ExecuteFrame(response))
}

// This is a no-op for httpFTransport
func (h *httpFTransport) SetMonitor(monitor FTransportMonitor) {
}

func (h *httpFTransport) makeRequest(requestPayload []byte) ([]byte, error) {
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

func (h *httpFTransport) getClosedConditionError(prefix string) error {
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s HTTP TTransport not open", prefix))
}
