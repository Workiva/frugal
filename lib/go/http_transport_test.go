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
	"encoding/base64"
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

type mockFProcessorForHTTP struct {
	err             error
	expectedPayload []byte
	response        []byte
}

func (m *mockFProcessorForHTTP) Process(iprot, oprot *FProtocol) error {
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

func (m *mockFProcessorForHTTP) AddMiddleware(middleware ServiceMiddleware) {}

func (m *mockFProcessorForHTTP) Annotations() map[string]map[string]string {
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

func getRequestHeaders(fctx FContext) map[string]string {
	return map[string]string{
		"first-header":              fctx.CorrelationID(),
		"second-header":             "yup",
		"content-type":              "these headers",
		"accept":                    "should be",
		"content-transfer-encoding": "overwritten",
	}
}

func getRequestHeadersEmpty(fctx FContext) map[string]string {
	return map[string]string{}
}

// Ensures that the payload size header has the wrong format an error is
// returned
func TestFrugalHandlerFuncHeaderError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	r, err := http.NewRequest("POST", "fooUrl", nil)
	r.Header.Add(payloadLimitHeader, "foo")
	assert.Nil(err)

	mockProcessor := &mockFProcessorForHTTP{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		fmt.Sprintf("%s header not an integer\n", payloadLimitHeader),
		string(w.Body.Bytes()),
	)
}

// Ensures that a 400 error is returned if the request body has less than 4
// bytes.
func TestFrugalHandlerFuncBadFrameError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	r, err := http.NewRequest("POST", "fooUrl", nil)
	assert.Nil(err)

	mockProcessor := &mockFProcessorForHTTP{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal("Invalid request size 0\n", string(w.Body.Bytes()))
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

	mockProcessor := &mockFProcessorForHTTP{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		fmt.Sprintf("Could not read the frugal frame bytes %s\n", io.ErrUnexpectedEOF),
		string(w.Body.Bytes()),
	)
}

// Ensures that processor errors are handled and routed back in the http
// response as a 500 error.
func TestFrugalHandlerFuncProcessorError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	processorErr := fmt.Errorf("processor error")
	mockProcessor := &mockFProcessorForHTTP{expectedPayload: expectedBody, err: processorErr}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusInternalServerError)
	assert.Equal(
		fmt.Sprintf("Error processing request: %s\n", processorErr),
		string(w.Body.Bytes()),
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
	mockProcessor := &mockFProcessorForHTTP{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusRequestEntityTooLarge)
	assert.Equal(
		"Response size (10) larger than requested size (5)\n",
		string(w.Body.Bytes()),
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
	mockProcessor := &mockFProcessorForHTTP{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory)

	oldNewEncoder := newEncoder
	writeErr := fmt.Errorf("write err")
	newEncoder = func(_ *bytes.Buffer) io.WriteCloser {
		return &mockWriteCloser{writeErr: writeErr}
	}

	handler(w, r)

	assert.Equal(w.Code, http.StatusInternalServerError)
	assert.Equal(
		fmt.Sprintf("Problem encoding frugal bytes to base64 %s\n", writeErr),
		string(w.Body.Bytes()),
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
	mockProcessor := &mockFProcessorForHTTP{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory)

	oldNewEncoder := newEncoder
	closeErr := fmt.Errorf("close err")
	newEncoder = func(_ *bytes.Buffer) io.WriteCloser {
		return &mockWriteCloser{closeErr: closeErr}
	}

	handler(w, r)

	assert.Equal(w.Code, http.StatusInternalServerError)
	assert.Equal(
		fmt.Sprintf("Problem encoding frugal bytes to base64 %s\n", closeErr),
		string(w.Body.Bytes()),
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
	mockProcessor := &mockFProcessorForHTTP{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusOK)
	assert.Equal(w.Header().Get(contentTypeHeader), frugalContentType)
	assert.Equal(w.Header().Get(contentTransferEncodingHeader), base64Encoding)
	assert.Equal(
		base64.StdEncoding.EncodeToString(append([]byte{0, 0, 0, 4}, response...)),
		string(w.Body.Bytes()),
	)

}

// Ensures the transport opens, writes, flushes, excecutes, and closes as
// expected
func TestHTTPTransportLifecycle(t *testing.T) {
	assert := assert.New(t)
	// Setup test data
	requestBytes := []byte("Hello from the other side")
	framedRequestBytes := prependFrameSize(requestBytes)
	responseBytes := []byte("I must've called a thousand times")
	f := make([]byte, 4)
	bigEndianPutUint32(f, uint32(len(responseBytes)))
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
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).Build().(*fHTTPTransport)

	// Open
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush before actually writing - make sure everything is fine
	_, err := transport.Request(ctx, []byte{0, 0, 0, 0})
	assert.Nil(err)

	// Flush
	result, err := transport.Request(ctx, framedRequestBytes)
	assert.Nil(err)
	assert.Equal(responseBytes, result.(*thrift.TMemoryBuffer).Bytes())

	// Close
	assert.Nil(transport.Close())
}

// Ensures custom headers can be written to the request
func TestHTTPRequestHeaders(t *testing.T) {
	assert := assert.New(t)
	// Setup test data
	requestBytes := []byte("Hello from the other side")
	framedRequestBytes := prependFrameSize(requestBytes)
	responseBytes := []byte("I must've called a thousand times")
	f := make([]byte, 4)
	bigEndianPutUint32(f, uint32(len(responseBytes)))
	framedResponse := append(f, responseBytes...)

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get(contentTypeHeader), frugalContentType)
		assert.Equal(r.Header.Get(contentTransferEncodingHeader), base64Encoding)
		assert.Equal(r.Header.Get(acceptHeader), frugalContentType)
		assert.Equal(r.Header.Get("foo"), "bar")

		respStr := base64.StdEncoding.EncodeToString(framedResponse)
		w.Write([]byte(respStr))
	}))

	customRequestHeaders := make(map[string]string, 1)
	customRequestHeaders["foo"] = "bar"
	// Instantiate http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).WithRequestHeaders(customRequestHeaders).Build().(*fHTTPTransport)

	// Open
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush before actually writing - make sure everything is fine
	_, err := transport.Request(ctx, []byte{0, 0, 0, 0})
	assert.Nil(err)

	// Flush
	result, err := transport.Request(ctx, framedRequestBytes)
	assert.Nil(err)
	assert.Equal(responseBytes, result.(*thrift.TMemoryBuffer).Bytes())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport handles one-way functions correctly
func TestHTTPTransportOneway(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")
	framedResponse := make([]byte, 4)

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respStr := base64.StdEncoding.EncodeToString(framedResponse)
		w.Write([]byte(respStr))
	}))

	// Instantiate http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).Build().(*fHTTPTransport)
	frameC := make(chan []byte, 1)

	// Open
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush
	err := transport.Oneway(ctx, requestBytes)
	assert.Nil(err)

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
func TestHTTPTransportBadRequest(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request bro"))
	}))

	// Instantiate and open http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush
	expectedErr := thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
		"response errored with code 400 and message bad request bro")
	_, actualErr := transport.Request(ctx, requestBytes)
	assert.Equal(actualErr.(thrift.TTransportException).TypeId(), expectedErr.TypeId())
	assert.Equal(actualErr.(thrift.TTransportException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error on a bad server data
func TestHTTPTransportBadResponseData(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("`"))
	}))

	// Instantiate and open http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush
	expectedErr := thrift.NewTTransportExceptionFromError(errors.New("illegal base64 data at input byte 0"))
	_, actualErr := transport.Request(ctx, requestBytes)
	assert.Equal(actualErr.(thrift.TTransportException).TypeId(), expectedErr.TypeId())
	assert.Equal(actualErr.(thrift.TTransportException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an a Request Too Large error with
// request data exceeds limit
func TestHTTPTransportRequestTooLarge(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.False(true)
	}))

	// Instantiate and open http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).WithRequestSizeLimit(10).Build()
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Write request
	_, err := transport.Request(ctx, requestBytes)

	assert.Equal(err.(thrift.TTransportException).TypeId(), TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE)

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an Response Too Large error when
// requesting too much data from the server
func TestHTTPTransportResponseTooLarge(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get(payloadLimitHeader), "10")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))

	// Instantiate and open http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).WithResponseSizeLimit(10).Build()
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush
	_, actualErr := transport.Request(ctx, requestBytes)
	assert.Equal(actualErr.(thrift.TTransportException).TypeId(), TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE)

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error when server doesn't return
// enough data
func TestHTTPTransportResponseNotEnoughData(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respStr := base64.StdEncoding.EncodeToString(make([]byte, 3))
		w.Write([]byte(respStr))
	}))

	// Instantiate and open http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush
	expectedErr := thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
		errors.New("frugal: invalid frame size"))
	_, actualErr := transport.Request(ctx, requestBytes)
	assert.Equal(actualErr.(thrift.TProtocolException).TypeId(), expectedErr.TypeId())
	assert.Equal(actualErr.(thrift.TProtocolException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error when server doesn't return
// an entire frame
func TestHTTPTransportResponseMissingFrameData(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respStr := base64.StdEncoding.EncodeToString([]byte{0, 0, 0, 1})
		w.Write([]byte(respStr))
	}))

	// Instantiate and open http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush
	expectedErr := thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
		errors.New("frugal: missing data"))
	_, actualErr := transport.Request(ctx, requestBytes)
	assert.Equal(actualErr.(thrift.TProtocolException).TypeId(), expectedErr.TypeId())
	assert.Equal(actualErr.(thrift.TProtocolException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures a timeout error is returned when the server doesn't respond
func TestHTTPTransportResponseTimeout(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer ts.Close()

	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).Build()
	assert.Nil(transport.Open())

	ctx := NewFContext("")
	ctx.SetTimeout(20 * time.Millisecond)
	_, actualErr := transport.Request(ctx, []byte{})
	assert.Equal(actualErr.(thrift.TTransportException).TypeId(), TRANSPORT_EXCEPTION_TIMED_OUT)

	assert.Nil(transport.Close())
}

// Ensures the transport flush returns an error when given a bad url
func TestHTTPTransportBadURL(t *testing.T) {
	assert := assert.New(t)

	// Setup test data
	requestBytes := []byte("Hello from the other side")

	// Instantiate and open http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, "nobody/home").Build()
	assert.Nil(transport.Open())

	// Create a context to use
	ctx := NewFContext("")

	// Flush
	expectedErr := thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
		"Post nobody/home: unsupported protocol scheme \"\"")
	_, actualErr := transport.Request(ctx, requestBytes)
	assert.Equal(actualErr.(thrift.TTransportException).TypeId(), expectedErr.TypeId())
	assert.Equal(actualErr.(thrift.TTransportException).Error(), expectedErr.Error())

	// Close
	assert.Nil(transport.Close())
}

// Ensures custom headers can be written to the request
func TestHTTPRequestHeadersWithContext(t *testing.T) {
	assert := assert.New(t)
	// Setup test data
	requestBytes := []byte("Hello from the other side")
	framedRequestBytes := prependFrameSize(requestBytes)
	responseBytes := []byte("I must've called a thousand times")
	f := make([]byte, 4)
	bigEndianPutUint32(f, uint32(len(responseBytes)))
	framedResponse := append(f, responseBytes...)

	// Setup test server
	// Create a context to use
	ctx := NewFContext("")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get(contentTypeHeader), frugalContentType)
		assert.Equal(r.Header.Get(contentTransferEncodingHeader), base64Encoding)
		assert.Equal(r.Header.Get(acceptHeader), frugalContentType)
		assert.Equal(r.Header.Get("foo"), "bar")
		assert.Equal(r.Header.Get("first-header"), ctx.CorrelationID())
		assert.Equal(r.Header.Get("second-header"), "yup")

		respStr := base64.StdEncoding.EncodeToString(framedResponse)
		w.Write([]byte(respStr))
	}))

	customRequestHeaders := make(map[string]string, 1)
	customRequestHeaders["foo"] = "bar"
	// Instantiate http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).
		WithRequestHeaders(customRequestHeaders).
		WithRequestHeadersFromFContext(getRequestHeaders).
		Build().(*fHTTPTransport)

	// Open
	assert.Nil(transport.Open())

	// Flush before actually writing - make sure everything is fine
	_, err := transport.Request(ctx, []byte{0, 0, 0, 0})
	assert.Nil(err)

	// Flush
	result, err := transport.Request(ctx, framedRequestBytes)
	assert.Nil(err)
	assert.Equal(responseBytes, result.(*thrift.TMemoryBuffer).Bytes())

	// Close
	assert.Nil(transport.Close())
}

// Ensures custom headers can be written to the request
func TestHTTPRequestHeadersWithContextReturnsEmptyMap(t *testing.T) {
	assert := assert.New(t)
	// Setup test data
	requestBytes := []byte("Hello from the other side")
	framedRequestBytes := prependFrameSize(requestBytes)
	responseBytes := []byte("I must've called a thousand times")
	f := make([]byte, 4)
	bigEndianPutUint32(f, uint32(len(responseBytes)))
	framedResponse := append(f, responseBytes...)

	// Setup test server
	// Create a context to use
	ctx := NewFContext("")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get(contentTypeHeader), frugalContentType)
		assert.Equal(r.Header.Get(contentTransferEncodingHeader), base64Encoding)
		assert.Equal(r.Header.Get(acceptHeader), frugalContentType)
		assert.Equal(r.Header.Get("foo"), "bar")

		respStr := base64.StdEncoding.EncodeToString(framedResponse)
		w.Write([]byte(respStr))
	}))

	customRequestHeaders := make(map[string]string, 1)
	customRequestHeaders["foo"] = "bar"
	// Instantiate http transport
	transport := NewFHTTPTransportBuilder(&http.Client{}, ts.URL).
		WithRequestHeaders(customRequestHeaders).
		WithRequestHeadersFromFContext(getRequestHeadersEmpty).
		Build().(*fHTTPTransport)

	// Open
	assert.Nil(transport.Open())

	// Flush before actually writing - make sure everything is fine
	_, err := transport.Request(ctx, []byte{0, 0, 0, 0})
	assert.Nil(err)

	// Flush
	result, err := transport.Request(ctx, framedRequestBytes)
	assert.Nil(err)
	assert.Equal(responseBytes, result.(*thrift.TMemoryBuffer).Bytes())

	// Close
	assert.Nil(transport.Close())
}
