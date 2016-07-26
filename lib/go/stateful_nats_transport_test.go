// TODO: Remove with 2.0
package frugal

import (
	"fmt"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Ensures Open returns an error if NATS is not connected.
func TestNatsServiceTTransportOpenNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	tr := NewNatsServiceTTransport(conn, "foo", 1, 3)

	assert.Error(t, tr.Open())
	assert.False(t, tr.IsOpen())
}

// Ensures Open returns an ALREADY_OPEN TTransportException if the transport
// is already open.
func TestNatsServiceTTransportOpenAlreadyOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	err := tr.Open()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.ALREADY_OPEN, trErr.TypeId())
}

// Ensures Open returns a TIMED_OUT TTransportException if the connect times
// out.
func TestNatsServiceTTransportOpenTimeout(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tr := NewNatsServiceTTransport(conn, "foo", 1, 3)

	err = tr.Open()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.TIMED_OUT, trErr.TypeId())
	assert.False(t, tr.IsOpen())
}

// Ensures Open subscribes to the right subject and buffers received frames
// which are returned on calls to Read.
func TestNatsServiceTTransportOpenRead(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	frame := []byte("helloworld")
	assert.Nil(t, conn.Publish(tr.listenTo, frame))

	buff := make([]byte, 5)
	n, err := tr.Read(buff)
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("hello"), buff)
	n, err = tr.Read(buff)
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("world"), buff)
}

// Ensures Open subscribes to the right subject and closes the transport when a
// disconnect message is received. Read returns an EOF when the transport has
// been closed.
func TestNatsServiceTTransportOpenDisconnectRead(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	assert.Nil(t, conn.PublishRequest(tr.listenTo, disconnect, nil))

	buff := make([]byte, 5)
	n, err := tr.Read(buff)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.END_OF_FILE, trErr.TypeId())
	assert.Equal(t, 0, n)
}

// Ensures Read returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestNatsServiceTTransportReadNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tr := NewNatsServiceTTransport(conn, "foo", 10*time.Millisecond, 3)

	n, err := tr.Read(make([]byte, 5))
	assert.Equal(t, 0, n)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Read returns a NOT_OPEN TTransportException if NATS is not connected.
func TestNatsServiceTTransportReadNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	conn.Close()

	buff := make([]byte, 5)
	n, err := tr.Read(buff)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
	assert.Equal(t, 0, n)
}

// Ensures Write returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestNatsServiceTTransportWriteNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	tr := NewNatsServiceTTransport(conn, "foo", 10*time.Millisecond, 3)

	n, err := tr.Write(make([]byte, 10))

	assert.Equal(t, 0, n)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Write returns a NOT_OPEN TTransportException if NATS is not
// connected.
func TestNatsServiceTTransportWriteNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	conn.Close()

	buff := make([]byte, 5)
	n, err := tr.Write(buff)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
	assert.Equal(t, 0, n)
}

// Ensures Write buffers data. If the buffer exceeds 1MB, ErrTooLarge is
// returned.
func TestNatsServiceTTransportWrite(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	buff := make([]byte, 5)
	n, err := tr.Write(buff)
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, tr.writeBuffer.Len())
	buff = make([]byte, 1024*1024)
	n, err = tr.Write(buff)
	assert.Equal(t, ErrTooLarge, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, tr.writeBuffer.Len())
}

// Ensures Flush returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestNatsServiceTTransportFlushNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tr := NewNatsServiceTTransport(conn, "foo", 10*time.Millisecond, 3)

	err = tr.Flush()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Flush returns a NOT_OPEN TTransportException if NATS is not
// connected.
func TestNatsServiceTTransportFlushNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	conn.Close()

	err := tr.Flush()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Flush doesn't send anything to NATS if no data is buffered.
func TestNatsServiceTTransportFlushNoData(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	sub, err := conn.SubscribeSync(tr.writeTo)
	assert.Nil(t, err)
	assert.Nil(t, tr.Flush())
	conn.Flush()
	_, err = sub.NextMsg(5 * time.Millisecond)
	assert.Equal(t, nats.ErrTimeout, err)
}

// Ensures Flush sends the frame to the correct NATS subject.
func TestNatsServiceTTransportFlush(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newNatsServiceClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	frame := []byte("helloworld")
	_, err := tr.Write(frame)
	assert.Nil(t, err)
	sub, err := conn.SubscribeSync(tr.writeTo)
	assert.Nil(t, err)
	assert.Nil(t, tr.Flush())
	conn.Flush()
	msg, err := sub.NextMsg(5 * time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, frame, msg.Data)
}

// Ensures RemainingBytes returns max uint64.
func TestNatsServiceTTransportRemainingBytes(t *testing.T) {
	tr := NewNatsServiceTTransport(nil, "foo", 10*time.Millisecond, 3)
	assert.Equal(t, ^uint64(0), tr.RemainingBytes())
}
func newNatsServiceClientAndServer(t *testing.T) (*natsServiceTTransport, *FNatsServer, *nats.Conn) {
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServer(conn, "foo", 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)
	mockTransport := new(mockFTransport)
	mockTransport.On("SetRegistry", mock.Anything).Return(nil)
	mockTransport.On("SetHighWatermark", defaultWatermark).Return(nil)
	mockTransport.On("Open").Return(nil)
	mockTransport.On("Closed").Return(toRecvChan(make(chan error)))
	mockTransport.On("Close").Return(nil)
	mockTransportFactory.On("GetTransport", mock.AnythingOfType("*frugal.natsServiceTTransport")).Return(mockTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mockTransport).Return(proto)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)
	tr := NewNatsServiceTTransport(conn, "foo", 10*time.Millisecond, 3)
	return tr.(*natsServiceTTransport), server.(*FNatsServer), conn
}
