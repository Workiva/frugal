package frugal

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

type mockFProcessorForHttp struct {
	err             error
	expectedPayload []byte
	response        []byte
}

func (m *mockFProcessorForHttp) Process(iprot, oprot *FProtocol) error {
	if m.expectedPayload != nil {
		actual := make([]byte, len(m.expectedPayload))
		if _, err := io.ReadFull(iprot.TProtocol.Transport(), actual); err != nil {
			return err
		}
		for i := range m.expectedPayload {
			if actual[i] != m.expectedPayload[i] {
				return errors.New("Payload doesn't match expected")
			}
		}
	}

	if m.err != nil {
		return m.err
	}

	oprot.TProtocol.Transport().Write(m.response)
	return nil
}

type mockWriteCloser struct {
	writeErr error
	closeErr error
}

func (m *mockWriteCloser) Write(p []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return len(p), nil
}

func (m *mockWriteCloser) Close() error {
	return m.closeErr
}

// Ensures that the payload size header has the wrong format an error is
// returned
func TestFrugalHandlerFuncHeaderError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	r, err := http.NewRequest("POST", "fooUrl", nil)
	r.Header.Add(payloadLimitHeader, "foo")
	assert.Nil(err)

	mockProcessor := &mockFProcessorForHttp{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		[]byte(fmt.Sprintf("%s header not an integer\n", payloadLimitHeader)),
		w.Body.Bytes(),
	)
}

// Ensures that if there is an error reading the frame size out of the request
// payload, an error is returned
func TestFrugalHandlerFuncFrameSizeError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	framedBody := []byte{0, 1, 2}
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	mockProcessor := &mockFProcessorForHttp{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		[]byte(fmt.Sprintf("Could not read the frugal frame bytes %s\n", io.ErrUnexpectedEOF)),
		w.Body.Bytes(),
	)
}

// Ensures that processor errors are handled and routed back in the http
// response
func TestFrugalHandlerFuncProcessorError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	processorErr := fmt.Errorf("processor error")
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, err: processorErr}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		[]byte(fmt.Sprintf("Frugal request failed %s\n", processorErr)),
		w.Body.Bytes(),
	)
}

// Ensures that if the response payload exceeds the request limit, a
// RequestEntityTooLarge error is returned.
func TestFrugalHandlerFuncTooLargeError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	r.Header.Add(payloadLimitHeader, "5")
	assert.Nil(err)

	response := make([]byte, 10)
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusRequestEntityTooLarge)
	assert.Equal(
		[]byte("Response size (10) larger than requested size (5)\n"),
		w.Body.Bytes(),
	)
}

// Ensures that base64 encoding errors are handled and routed back in the http
// response
func TestFrugalHandlerFuncBase64WriteError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	response := []byte("Hello")
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	oldNewEncoder := newEncoder
	writeErr := fmt.Errorf("write err")
	newEncoder = func(_ *bytes.Buffer) io.WriteCloser {
		return &mockWriteCloser{writeErr: writeErr}
	}

	handler(w, r)

	assert.Equal(w.Code, http.StatusInternalServerError)
	assert.Equal(
		[]byte(fmt.Sprintf("Problem encoding frugal bytes to base64 %s\n", writeErr)),
		w.Body.Bytes(),
	)

	newEncoder = oldNewEncoder
}

// Ensures that base64 encoding errors are handled and routed back in the http
// response
func TestFrugalHandlerFuncBase64CloseError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	response := []byte("Hello")
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	oldNewEncoder := newEncoder
	closeErr := fmt.Errorf("close err")
	newEncoder = func(_ *bytes.Buffer) io.WriteCloser {
		return &mockWriteCloser{closeErr: closeErr}
	}

	handler(w, r)

	assert.Equal(w.Code, http.StatusInternalServerError)
	assert.Equal(
		[]byte(fmt.Sprintf("Problem encoding frugal bytes to base64 %s\n", closeErr)),
		w.Body.Bytes(),
	)

	newEncoder = oldNewEncoder
}

// Ensures that the frugal payload is appropriately processed and returned
func TestFrugalHandlerFunc(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	response := []byte{9, 10, 11, 12}
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusOK)
	assert.Equal(w.Header().Get(contentTypeHeader), frugalContentType)
	assert.Equal(w.Header().Get(contentTransferEncodingHeader), base64Encoding)
	assert.Equal(
		[]byte(base64.StdEncoding.EncodeToString(append([]byte{0, 0, 0, 4}, response...))),
		w.Body.Bytes(),
	)

}

// Ensures the transport opens, writes, flushes, excecutes, and closes as
// expected
func TestHttpTransportLifecycle(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")
	responseBytes := []byte("I must've called a thousand times")
	f := make([]byte, 4)
	binary.BigEndian.PutUint32(f, uint32(len(responseBytes)))
	framedResponse := append(f, responseBytes...)

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get(contentTypeHeader), frugalContentType)
		assert.Equal(r.Header.Get(contentTransferEncodingHeader), base64Encoding)
		assert.Equal(r.Header.Get(acceptHeader), frugalContentType)

		respStr := base64.StdEncoding.EncodeToString(framedResponse)
		w.Write([]byte(respStr))
	}))

	// Instantiate http transport
	transport := NewHttpFTransportBuilder(&http.Client{}, ts.URL).Build()
	frameC := make(chan []byte, 1)
	flushErr := fmt.Errorf("foo")
	registry := &mockRegistry{
		frameC: frameC,
		err:    flushErr,
	}
	transport.SetRegistry(registry)

	// Open
	assert.Nil(transport.Open())

	// Flush before actually writing - make sure everything is fine
	assert.Nil(transport.Flush())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	assert.Equal(thrift.NewTTransportExceptionFromError(flushErr), transport.Flush())
	select {
	case actual := <-frameC:
		assert.Equal(responseBytes, actual)
	case <-time.After(time.Second):
		assert.True(false)
	}

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport handles one-way functions correctly
func TestHttpTransportOneway(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")
	framedResponse := make([]byte, 4)

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respStr := base64.StdEncoding.EncodeToString(framedResponse)
		w.Write([]byte(respStr))
	}))

	// Instantiate http transpor
	transport := NewHttpFTransportBuilder(&http.Client{}, ts.URL).Build()
	frameC := make(chan []byte, 1)
	flushErr := fmt.Errorf("foo")
	registry := &mockRegistry{
		frameC: frameC,
		err:    flushErr,
	}
	transport.SetRegistry(registry)

	// Open
	assert.Nil(transport.Open())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	assert.Nil(transport.Flush())

	// Make sure nothing is executed on the registry
	select {
	case <-frameC:
		assert.True(false)
	default:
	}

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error on a bad request
func TestHttpTransportBadRequest(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request bro"))
	}))

	// Instantiate and open http transport
	transport := NewHttpFTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	expectedErr := thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
		"response errored with code 400 and message bad request bro")
	actuaErr := transport.Flush()
	assert.Equal(actuaErr.(thrift.TTransportException).TypeId(), expectedErr.TypeId())
	assert.Equal(actuaErr.(thrift.TTransportException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error on a bad server data
func TestHttpTransportBadResponseData(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("`"))
	}))

	// Instantiate and open http transport
	transport := NewHttpFTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	expectedErr := thrift.NewTTransportExceptionFromError(errors.New("illegal base64 data at input byte 0"))
	actuaErr := transport.Flush()
	assert.Equal(actuaErr.(thrift.TTransportException).TypeId(), expectedErr.TypeId())
	assert.Equal(actuaErr.(thrift.TTransportException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an a Request Too Large error with
// request data exceeds limit
func TestHttpTransportRequestTooLarge(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.False(true)
	}))

	// Instantiate and open http transport
	transport := NewHttpFTransportBuilder(&http.Client{}, ts.URL).WithRequestSizeLimit(10).Build()
	assert.Nil(transport.Open())

	// Write request
	expectedErr := ErrTooLarge
	n, err := transport.Write(requestBytes)
	assert.Equal(0, n)

	assert.Equal(err.(thrift.TTransportException).TypeId(), expectedErr.TypeId())
	assert.Equal(err.(thrift.TTransportException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an Response Too Large error when
// requesting too much data from the server
func TestHttpTransportResponseTooLarge(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get(payloadLimitHeader), "10")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))

	// Instantiate and open http transport
	transport := NewHttpFTransportBuilder(&http.Client{}, ts.URL).WithResponseSizeLimit(10).Build()
	assert.Nil(transport.Open())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	expectedErr := thrift.NewTTransportException(RESPONSE_TOO_LARGE, "response was too large for the transport")
	actuaErr := transport.Flush()
	assert.Equal(actuaErr.(thrift.TTransportException).TypeId(), expectedErr.TypeId())
	assert.Equal(actuaErr.(thrift.TTransportException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error when server doesn't return
// enough data
func TestHttpTransportResponseNotEnoughData(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respStr := base64.StdEncoding.EncodeToString(make([]byte, 3))
		w.Write([]byte(respStr))
	}))

	// Instantiate and open http transport
	transport := NewHttpFTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	expectedErr := thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
		errors.New("frugal: invalid frame size"))
	actuaErr := transport.Flush()
	assert.Equal(actuaErr.(thrift.TProtocolException).TypeId(), expectedErr.TypeId())
	assert.Equal(actuaErr.(thrift.TProtocolException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error when server doesn't return
// an entire frame
func TestHttpTransportResponseMissingFrameData(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respStr := base64.StdEncoding.EncodeToString([]byte{0, 0, 0, 1})
		w.Write([]byte(respStr))
	}))

	// Instantiate and open http transport
	transport := NewHttpFTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	expectedErr := thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
		errors.New("frugal: missing data"))
	actuaErr := transport.Flush()
	assert.Equal(actuaErr.(thrift.TProtocolException).TypeId(), expectedErr.TypeId())
	assert.Equal(actuaErr.(thrift.TProtocolException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error when given a bad url
func TestHttpTransportBadURL(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Instantiate and open http transport
	transport := NewHttpFTransportBuilder(&http.Client{}, "nobody/home").Build()
	assert.Nil(transport.Open())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	expectedErr := thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
		"Post nobody/home: unsupported protocol scheme \"\"")
	actuaErr := transport.Flush()
	assert.Equal(actuaErr.(thrift.TTransportException).TypeId(), expectedErr.TypeId())
	assert.Equal(actuaErr.(thrift.TTransportException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures Read throws an error whenever called.
func TestHttpTransportRead(t *testing.T) {
	transport := NewHttpFTransportBuilder(&http.Client{}, "").Build()
	_, err := transport.Read(make([]byte, 0))
	assert.Error(t, err)
}

// TESTS FOR DEPRECATED HttpTTransport

// Ensures the transport opens, writes, reads, flushes, closes as expected
func TestHttpTTransportLifecycle(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")
	responseBytes := []byte("I must've called a thousand times")
	f := make([]byte, 4)
	binary.BigEndian.PutUint32(f, uint32(len(responseBytes)))
	framedResponse := append(f, responseBytes...)

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get(contentTypeHeader), frugalContentType)
		assert.Equal(r.Header.Get(contentTransferEncodingHeader), base64Encoding)
		assert.Equal(r.Header.Get(acceptHeader), frugalContentType)

		respStr := base64.StdEncoding.EncodeToString(framedResponse)
		w.Write([]byte(respStr))
	}))

	// Instantiate and open http transport
	transport := NewHttpTTransport(&http.Client{}, ts.URL)
	assert.Nil(transport.Open())

	// Flush before actually writing - make sure everything is fine
	assert.Nil(transport.Flush())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	assert.Nil(transport.Flush())

	// Read response
	actualResp := make([]byte, len(framedResponse))
	n, err = transport.Read(actualResp)
	assert.Equal(len(framedResponse), n)
	assert.Nil(err)
	assert.Equal(framedResponse, actualResp)

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport handles one-way functions correctly
func TestHttpTTransportOneway(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")
	framedResponse := make([]byte, 4)

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respStr := base64.StdEncoding.EncodeToString(framedResponse)
		w.Write([]byte(respStr))
	}))

	// Instantiate and open http transport
	transport := NewHttpTTransport(&http.Client{}, ts.URL)
	assert.Nil(transport.Open())

	// Write request
	n, err := transport.Write(requestBytes)
	assert.Equal(len(requestBytes), n)
	assert.Nil(err)

	// Flush
	assert.Nil(transport.Flush())

	// Trigger a delayed closed to unblock read
	go func() {
		time.Sleep(100 * time.Millisecond)
		// Close
		assert.Nil(transport.Close())
	}()

	// Read response
	n, err = transport.Read([]byte{})
	assert.Equal(0, n)
	assert.Equal(thrift.NewTTransportExceptionFromError(io.EOF), err)
}
