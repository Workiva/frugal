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
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRegistry struct {
	registered chan uint64
	frameC     chan []byte
	err        error
}

func (m *mockRegistry) AssignOpID(ctx FContext) error {
	return nil
}

func (m *mockRegistry) Register(ctx FContext, resultC chan []byte) error {
	opIdStr, _ := ctx.RequestHeader(opIDHeader)
	opId, _ := strconv.ParseUint(opIdStr, 10, 64)
	m.registered <- opId

	m.frameC = resultC
	return nil
}

func (m *mockRegistry) Unregister(ctx FContext) {
}

func (m *mockRegistry) Execute(frame []byte) error {
	m.frameC <- frame
	return m.err
}

func (m *mockRegistry) dispatch(opid uint64, frame []byte) error {
	m.frameC <- frame
	return m.err
}

// Ensures Open returns an error if NATS is not connected.
func TestNatsTransportOpenNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	tr := NewFNatsTransport(conn, "foo", "bar")

	assert.Error(t, tr.Open())
	assert.False(t, tr.IsOpen())
}

// Ensures Open returns an ALREADY_OPEN TTransportException if the transport
// is already open.
func TestNatsTransportOpenAlreadyOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	err := tr.Open()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, TRANSPORT_EXCEPTION_ALREADY_OPEN, trErr.TypeId())
}

// Ensures Open subscribes to the right subject and executes received frames.
func TestNatsTransportOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	frame := []byte("helloworld")
	frameC := make(chan []byte, 1)
	registry := &mockRegistry{
		err:        fmt.Errorf("foo"),
		registered: make(chan uint64, 1),
	}
	registry.Register(NewFContext(""), frameC)
	tr.registry = registry

	sizedFrame := prependFrameSize(frame)
	assert.Nil(t, conn.Publish(tr.inbox+".1", sizedFrame))

	select {
	case actual := <-frameC:
		assert.Equal(t, frame, actual)
	case <-time.After(time.Second):
		assert.True(t, false)
	}
}

// Ensures Request returns TTransportException with type
// TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE if too much data is provided.
func TestNatsTransportRequestTooLarge(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	buff := make([]byte, 1024*1024+1)
	_, err := tr.Request(NewFContext(""), buff)
	assert.True(t, IsErrTooLarge(err))
	assert.Equal(t, TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE, err.(thrift.TTransportException).TypeId())
	assert.Equal(t, 0, tr.writeBuffer.Len())
}

// Ensures Request returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestNatsTransportRequestNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tr := NewFNatsTransport(conn, "foo", "bar")

	_, err = tr.Request(nil, []byte{})
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, TRANSPORT_EXCEPTION_NOT_OPEN, trErr.TypeId())
}

// Ensures Request returns a NOT_OPEN TTransportException if NATS is not
// connected.
func TestNatsTransportRequestNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	conn.Close()

	_, err := tr.Request(nil, []byte{})
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, TRANSPORT_EXCEPTION_NOT_OPEN, trErr.TypeId())
}

// Ensures Request doesn't send anything to NATS if no data is buffered.
func TestNatsTransportRequesthNoData(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	sub, err := conn.SubscribeSync(tr.subject)
	assert.Nil(t, err)
	_, err = tr.Request(nil, []byte{0, 0, 0, 0})
	assert.Nil(t, err)
	conn.Flush()
	_, err = sub.NextMsg(5 * time.Millisecond)
	assert.Equal(t, nats.ErrTimeout, err)
}

// Ensures Request sends the frame to the correct NATS subject.
func TestNatsTransportRequest(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	frame := []byte("helloworld")
	sub, err := conn.SubscribeSync(tr.subject)
	assert.Nil(t, err)

	ctx := NewFContext("")
	ctx.SetTimeout(5 * time.Millisecond)
	_, err = tr.Request(ctx, prependFrameSize(frame))
	// expect a timeout error because nothing is answering
	assert.Equal(t, TRANSPORT_EXCEPTION_TIMED_OUT, err.(thrift.TTransportException).TypeId())
	conn.Flush()
	msg, err := sub.NextMsg(5 * time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, prependFrameSize(frame), msg.Data)
}

// Ensures Request returns an error if a duplicate opid is used.
func TestNatsTransportRequestSameOpid(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	frame := []byte("helloworld")
	ctx := NewFContext("")
	go func() {
		tr.Request(ctx, prependFrameSize(frame))
	}()
	time.Sleep(10 * time.Millisecond)
	_, err := tr.Request(ctx, prependFrameSize(frame))
	assert.Equal(t, TRANSPORT_EXCEPTION_UNKNOWN, err.(thrift.TTransportException).TypeId())
	opID, opErr := getOpID(ctx)
	assert.Nil(t, opErr)
	assert.Equal(t, fmt.Sprintf("frugal: context already registered, opid %d is in-flight for another request", opID), err.Error())
}

// Ensures empty status messages do not cause panic
func TestStatusMessage(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)

	defer server.Stop()
	defer conn.Close()
	defer tr.Close()

	msg := nats.Msg{
		Data:    make([]byte, 0),
		Subject: "subject.1",
		Header:  map[string][]string{"Status": {"503"}},
	}
	tr.handler(&msg)
}

func TestServiceUnavailable(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	frame := []byte("helloworld")
	_, err := conn.SubscribeSync(tr.subject)
	assert.Nil(t, err)

	ctx := NewFContext("")
	opId, ok := ctx.RequestHeader(opIDHeader)
	assert.True(t, ok)

	// Wait until the op id registered before calling the handler
	// the handler does not 
	go func() {
		opIdInt, _ := strconv.ParseInt(opId, 10, 64)
		var found bool
		for !found {
			tr.registry.(*fRegistryImpl).mu.RLock()
			_, found = tr.registry.(*fRegistryImpl).channels[uint64(opIdInt)]
			tr.registry.(*fRegistryImpl).mu.RUnlock()
			runtime.Gosched()
		}
		tr.handler(&nats.Msg{
			// Mimic a 503 being returned
			// Start the request asynchronously so we can mock a response
			Subject: tr.inbox + "." + opId,
			Header:  map[string][]string{"Status": {"503"}},
		})
	}()
	_, err = tr.Request(ctx, prependFrameSize(frame))
	assert.Equal(t, TRANSPORT_EXCEPTION_SERVICE_NOT_AVAILABLE, err.(thrift.TTransportException).TypeId())

	conn.Flush()
}

// HELPER METHODS

func newClientAndServer(t *testing.T, isTTransport bool) (*fNatsTransport, *fNatsServer, *nats.Conn) {
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	mockProcessor := new(mockFProcessor)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServerBuilder(conn, mockProcessor, protocolFactory, []string{"foo"}).
		WithQueueGroup("queue").
		WithWorkerCount(1).
		WithRequestReceivedEventHandler(func(map[interface{}]interface{}) {}).
		WithRequestStartedEventHandler(func(map[interface{}]interface{}) {}).
		Build()
	mockTransport := new(mockFTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mock.AnythingOfType("*thrift.TMemoryBuffer")).Return(proto).Once()
	mockTProtocolFactory.On("GetProtocol", mock.AnythingOfType("*frugal.TMemoryOutputBuffer")).Return(proto).Once()
	inputFproto := &FProtocol{TProtocol: proto, ephemeralProperties: make(map[interface{}]interface{})}
	outputPproto := &FProtocol{TProtocol: proto, ephemeralProperties: make(map[interface{}]interface{})}
	mockProcessor.On("Process", inputFproto, outputPproto).Return(nil)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)
	var tr *fNatsTransport
	if isTTransport {
		tr = NewFNatsTransport(conn, "foo", "bar").(*fNatsTransport)
	} else {
		tr = NewFNatsTransport(conn, "foo", "bar").(*fNatsTransport)
	}
	return tr, server.(*fNatsServer), conn
}
