package frugal

import (
	"fmt"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRegistry struct {
	frameC chan ([]byte)
	err    error
}

func (m *mockRegistry) Register(ctx FContext, callback FAsyncCallback) error {
	return nil
}

func (m *mockRegistry) Unregister(ctx FContext) {
}

func (m *mockRegistry) Execute(frame []byte) error {
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
	assert.Equal(t, thrift.ALREADY_OPEN, trErr.TypeId())
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
	frameC := make(chan []byte)
	registry := &mockRegistry{
		frameC: frameC,
		err:    fmt.Errorf("foo"),
	}
	tr.registry = registry

	sizedFrame := prependFrameSize(frame)
	assert.Nil(t, conn.Publish(tr.inbox, sizedFrame))

	select {
	case actual := <-frameC:
		assert.Equal(t, frame, actual)
	case <-time.After(time.Second):
		assert.True(t, false)
	}
}

// Ensures Write buffers data. If the buffer exceeds 1MB, ErrTooLarge is
// returned.
func TestNatsTransportWrite(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	buff := make([]byte, 5)
	buff = make([]byte, 1024*1024+1)
	err := tr.Send(buff)
	assert.True(t, IsErrTooLarge(err))
	assert.Equal(t, TTRANSPORT_REQUEST_TOO_LARGE, err.(thrift.TTransportException).TypeId())
	assert.Equal(t, 0, tr.writeBuffer.Len())
}

// Ensures Flush returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestNatsTransportFlushNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tr := NewFNatsTransport(conn, "foo", "bar")

	err = tr.Send([]byte{})
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Flush returns a NOT_OPEN TTransportException if NATS is not
// connected.
func TestNatsTransportFlushNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newClientAndServer(t, false)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	conn.Close()

	err := tr.Send([]byte{})
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Flush doesn't send anything to NATS if no data is buffered.
func TestNatsTransportFlushNoData(t *testing.T) {
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
	assert.Nil(t, tr.Send([]byte{0, 0, 0, 0}))
	conn.Flush()
	_, err = sub.NextMsg(5 * time.Millisecond)
	assert.Equal(t, nats.ErrTimeout, err)
}

// Ensures Flush sends the frame to the correct NATS subject.
func TestNatsTransportFlush(t *testing.T) {
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
	assert.Nil(t, tr.Send(prependFrameSize(frame)))
	conn.Flush()
	msg, err := sub.NextMsg(5 * time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, prependFrameSize(frame), msg.Data)
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
		Build()
	mockTransport := new(mockFTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mock.AnythingOfType("*thrift.TMemoryBuffer")).Return(proto).Once()
	mockTProtocolFactory.On("GetProtocol", mock.AnythingOfType("*frugal.TMemoryOutputBuffer")).Return(proto).Once()
	fproto := &FProtocol{proto}
	mockProcessor.On("Process", fproto, fproto).Return(nil)

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
