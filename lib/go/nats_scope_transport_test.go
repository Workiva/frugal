package frugal

import (
	"fmt"
	"net"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
)

var defaultOptions = server.Options{
	Host:   "localhost",
	Port:   11222,
	NoLog:  true,
	NoSigs: true,
}

func runServer(opts *server.Options) *server.Server {
	if opts == nil {
		opts = &defaultOptions
	}
	s := server.New(opts)
	if s == nil {
		panic("No NATS Server object returned.")
	}

	// Run server in Go routine.
	go s.Start()

	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		addr := s.GetListenEndpoint()
		if addr == "" {
			time.Sleep(10 * time.Millisecond)
			// Retry. We might take a little while to open a connection.
			continue
		}
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			// Retry after 50ms
			time.Sleep(50 * time.Millisecond)
			continue
		}
		conn.Close()
		// Wait a bit to give a chance to the server to remove this
		// "client" from its state, which may otherwise interfere with
		// some tests.
		time.Sleep(25 * time.Millisecond)

		return s
	}
	panic("Unable to start NATS Server in Go Routine")

}

// Ensures LockTopic returns nil when a publisher successfully locks a topic.
// Subsequent calls will wait on the mutex. Unlock releases the topic.
func TestNatsScopeLockUnlockTopic(t *testing.T) {
	tr := NewNatsFScopeTransport(nil)
	assert.Nil(t, tr.LockTopic("foo"))
	acquired := make(chan bool)
	go func() {
		assert.Nil(t, tr.LockTopic("bar"))
		assert.Equal(t, "bar", tr.(*fNatsScopeTransport).subject)
		acquired <- true
	}()
	assert.Equal(t, "foo", tr.(*fNatsScopeTransport).subject)
	tr.UnlockTopic()
	<-acquired

	tr.UnlockTopic()
	assert.Equal(t, "", tr.(*fNatsScopeTransport).subject)
}

// Ensures LockTopic returns an error if the transport is a subscriber.
func TestNatsScopeLockTopicSubscriberError(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)

	tr.Subscribe("foo")

	assert.Error(t, tr.LockTopic("blah"))
}

// Ensures UnlockTopic returns an error if the transport is a subscriber.
func TestNatsScopeUnlockTopicSubscriberError(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)

	tr.Subscribe("foo")

	assert.Error(t, tr.UnlockTopic())
}

// Ensures Subscribe subscribes to the topic on NATS and puts received frames
// on the read buffer and received in Read calls. All subscribers receive the
// message when they aren't subscribed to a queue.
func TestNatsScopeSubscribeRead(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr1 := NewNatsFScopeTransport(conn)
	tr2 := NewNatsFScopeTransport(conn)

	assert.Nil(t, tr1.Subscribe("foo"))
	assert.Nil(t, tr2.Subscribe("foo"))

	frame := make([]byte, 50)
	assert.Nil(t, conn.Publish("frugal.foo", append(make([]byte, 4), frame...))) // Add 4 bytes for frame size

	// Both transports should receive the frame.
	frameBuff := []byte{}
	buff := make([]byte, 10)
	for i := 0; i < 5; i++ {
		n, err := tr1.Read(buff)
		assert.Nil(t, err)
		assert.Equal(t, 10, n)
		frameBuff = append(frameBuff, buff...)
	}
	assert.Equal(t, frame, frameBuff)

	frameBuff = []byte{}
	buff = make([]byte, 10)
	for i := 0; i < 5; i++ {
		n, err := tr2.Read(buff)
		assert.Nil(t, err)
		assert.Equal(t, 10, n)
		frameBuff = append(frameBuff, buff...)
	}
	assert.Equal(t, frame, frameBuff)
}

// Ensures Subscribe subscribes to the topic on NATS and puts received frames
// on the read buffer. If the transport specifies a queue, only one member of
// the queue group receives the message.
func TestNatsScopeSubscribeQueue(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr1 := NewNatsFScopeTransportWithQueue(conn, "fooqueue")
	tr2 := NewNatsFScopeTransportWithQueue(conn, "fooqueue")

	assert.Nil(t, tr1.Subscribe("foo"))
	assert.Nil(t, tr2.Subscribe("foo"))

	frame := make([]byte, 50)
	assert.Nil(t, conn.Publish("frugal.foo", append(make([]byte, 4), frame...))) // Add 4 bytes for frame size
	conn.Flush()
	time.Sleep(10 * time.Millisecond)

	// Only one of the two transports should have received the frame.
	received := false
	select {
	case fr := <-tr1.(*fNatsScopeTransport).frameBuffer:
		assert.Equal(t, frame, fr)
		received = true
	default:
	}
	select {
	case fr := <-tr2.(*fNatsScopeTransport).frameBuffer:
		assert.False(t, received)
		assert.Equal(t, frame, fr)
		received = true
	default:
	}

	assert.True(t, received)
}

// Ensures Read returns an EOF if the transport is not open.
func TestNatsScopeReadNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)

	n, err := tr.Read(make([]byte, 5))
	assert.Equal(t, 0, n)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.END_OF_FILE, trErr.TypeId())
}

// Ensures Subscribe returns an error if the NATS connection is not open.
func TestNatsScopeSubscribeConnectionNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	tr := NewNatsFScopeTransport(conn)
	conn.Close()

	assert.Error(t, tr.Subscribe("foo"))
}

// Ensures Open returns nil on success and writes work.
func TestNatsScopeOpenPublisherWriteFlush(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)

	assert.Nil(t, tr.Open())
	assert.True(t, tr.IsOpen())
	assert.Nil(t, tr.LockTopic("foo"))
	frame := make([]byte, 10)
	n, err := tr.Write(frame)
	assert.Nil(t, err)
	assert.Equal(t, 10, n)
	received := make(chan bool)
	_, err = conn.Subscribe("frugal.foo", func(msg *nats.Msg) {
		assert.Equal(t, append([]byte{0, 0, 0, 10}, frame...), msg.Data)
		received <- true
	})
	assert.Nil(t, err)
	assert.Nil(t, tr.Flush())
	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("expected to receive frame")
	}
}

// Ensures Open returns an ALREADY_OPEN TTransportException if the transport is
// already open.
func TestNatsScopeOpenAlreadyOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)

	assert.Nil(t, tr.Open())

	err = tr.Open()

	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.ALREADY_OPEN, trErr.TypeId())
}

// Ensures Open returns an error for subscribers with no subject set.
func TestNatsScopeOpenSubscriberNoSubject(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)
	tr.(*fNatsScopeTransport).pull = true

	assert.Error(t, tr.Open())
}

// Ensures subscribers discard invalid frames (size < 4).
func TestNatsScopeDiscardInvalidFrame(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)
	assert.Nil(t, tr.Subscribe("blah"))

	closed := make(chan bool)
	go func() {
		buff := make([]byte, 3)
		n, err := tr.Read(buff)
		assert.Equal(t, 0, n)
		trErr := err.(thrift.TTransportException)
		assert.Equal(t, thrift.END_OF_FILE, trErr.TypeId())
		closed <- true
	}()

	assert.Nil(t, conn.Publish("frugal.blah", make([]byte, 2)))
	assert.Nil(t, conn.Flush())
	time.Sleep(10 * time.Millisecond)
	assert.Nil(t, tr.Close())
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("expected transport to close")
	}
}

// Ensures Close returns nil if the transport is not open.
func TestNatsScopeCloseNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)
	assert.Nil(t, tr.Close())
	assert.False(t, tr.IsOpen())
}

// Ensures Close closes the publisher transport and returns nil.
func TestNatsScopeClosePublisher(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)
	assert.Nil(t, tr.LockTopic("foo"))
	assert.Nil(t, tr.Open())
	assert.True(t, tr.IsOpen())
	assert.Nil(t, tr.Close())
	assert.False(t, tr.IsOpen())
}

// Ensures Close returns an error if the unsubscribe fails.
func TestNatsScopeCloseSubscriberError(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	tr := NewNatsFScopeTransport(conn)
	assert.Nil(t, tr.Subscribe("foo"))
	conn.Close()
	assert.Error(t, tr.Close())
}

// Ensures Write returns an ErrTooLarge if the written frame exceeds 1MB.
func TestNatsScopeWriteTooLarge(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)
	assert.Nil(t, tr.Open())

	n, err := tr.Write(make([]byte, 5))
	assert.Equal(t, 5, n)
	assert.Nil(t, err)
	n, err = tr.Write(make([]byte, 1024*1024+10))
	assert.Equal(t, 0, n)
	assert.Equal(t, ErrTooLarge, err)
	assert.Equal(t, 0, tr.(*fNatsScopeTransport).writeBuffer.Len())
}

// Ensures Flush returns an error if the transport is not open.
func TestNatsScopeFlushNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)

	err = tr.Flush()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Flush returns nil and nothing is sent to NATS when there is no data
// to flush.
func TestNatsScopeFlushNoData(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	tr := NewNatsFScopeTransport(conn)
	assert.Nil(t, tr.Open())
	assert.Nil(t, tr.LockTopic("foo"))
	_, err = conn.Subscribe("frugal.foo", func(msg *nats.Msg) {
		t.Fatal("did not expect to receive message")
	})
	assert.Nil(t, err)
	assert.Nil(t, tr.Flush())
	assert.Nil(t, conn.Flush())
	time.Sleep(10 * time.Millisecond)
}

// Ensures RemainingBytes returns max uint64.
func TestNatsScopeRemainingBytes(t *testing.T) {
	tr := NewNatsFScopeTransport(nil)
	assert.Equal(t, ^uint64(0), tr.RemainingBytes())
}
