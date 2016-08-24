package frugal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

// frameBufferSize is the number of message frames to buffer on the subscriber.
const frameBufferSize = 5

// FNatsScopeTransportFactory creates FNatsScopeTransports.
type FNatsScopeTransportFactory struct {
	conn  *nats.Conn
	queue string
}

// NewFNatsScopeTransportFactory creates an FNatsScopeTransportFactory using
// the provided NATS connection. Subscribers using this transport will not use
// a queue.
func NewFNatsScopeTransportFactory(conn *nats.Conn) *FNatsScopeTransportFactory {
	return &FNatsScopeTransportFactory{conn: conn}
}

// NewFNatsScopeTransportFactoryWithQueue creates an FNatsScopeTransportFactory
// using the provided NATS connection. Subscribers using this transport will
// subscribe to the provided queue, forming a queue group. When a queue group
// is formed, only one member receives the message.
func NewFNatsScopeTransportFactoryWithQueue(conn *nats.Conn, queue string) *FNatsScopeTransportFactory {
	return &FNatsScopeTransportFactory{conn: conn, queue: queue}
}

// GetTransport creates a new NATS FScopeTransport.
func (n *FNatsScopeTransportFactory) GetTransport() FScopeTransport {
	return NewNatsFScopeTransportWithQueue(n.conn, n.queue)
}

// fNatsScopeTransport implements FScopeTransport.
type fNatsScopeTransport struct {
	conn         *nats.Conn
	subject      string
	queue        string
	frameBuffer  chan []byte
	closeChan    chan struct{}
	currentFrame []byte
	writeBuffer  *bytes.Buffer
	sub          *nats.Subscription
	pull         bool
	topicMu      sync.Mutex
	openMu       sync.RWMutex
	isOpen       bool
	sizeBuffer   []byte
}

// NewNatsFScopeTransport creates a new FScopeTransport which is used for
// pub/sub. Subscribers using this transport will not use a queue.
func NewNatsFScopeTransport(conn *nats.Conn) FScopeTransport {
	return &fNatsScopeTransport{conn: conn}
}

// NewNatsFScopeTransportWithQueue creates a new FScopeTransport which is used
// for pub/sub. Subscribers using this transport will subscribe to the provided
// queue, forming a queue group. When a queue group is formed, only one member
// receives the message.
func NewNatsFScopeTransportWithQueue(conn *nats.Conn, queue string) FScopeTransport {
	return &fNatsScopeTransport{conn: conn, queue: queue}
}

// LockTopic sets the publish topic and locks the transport for exclusive
// access.
func (n *fNatsScopeTransport) LockTopic(topic string) error {
	if n.pull {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"subscriber cannot lock topic")
	}
	n.topicMu.Lock()
	n.subject = topic
	return nil
}

// UnlockTopic unsets the publish topic and unlocks the transport.
func (n *fNatsScopeTransport) UnlockTopic() error {
	if n.pull {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"subscriber cannot unlock topic")
	}
	n.subject = ""
	n.topicMu.Unlock()
	return nil
}

// Subscribe sets the subscribe topic and opens the transport.
func (n *fNatsScopeTransport) Subscribe(topic string) error {
	n.pull = true
	n.subject = topic
	return n.Open()
}

// Open initializes the transport based on whether it's a publisher or
// subscriber. If Open is called before Subscribe, the transport is assumed to
// be a publisher.
func (n *fNatsScopeTransport) Open() error {
	n.openMu.Lock()
	defer n.openMu.Unlock()
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: NATS not connected, has status %d", n.conn.Status()))
	}

	if n.isOpen {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS transport already open")
	}

	if !n.pull {
		n.writeBuffer = bytes.NewBuffer(make([]byte, 0, natsMaxMessageSize))
		n.sizeBuffer = make([]byte, 4)
		n.isOpen = true
		return nil
	}

	if n.subject == "" {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"cannot subscribe to empty subject")
	}

	n.closeChan = make(chan struct{})
	n.frameBuffer = make(chan []byte, frameBufferSize)

	sub, err := n.conn.QueueSubscribe(n.formattedSubject(), n.queue, n.handleMessage)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	n.sub = sub
	n.isOpen = true
	return nil
}

func (n *fNatsScopeTransport) handleMessage(msg *nats.Msg) {
	if len(msg.Data) < 4 {
		logger().Warn("frugal: Discarding invalid scope message frame")
		return
	}
	// Discard frame size.
	select {
	case n.frameBuffer <- msg.Data[4:]:
	case <-n.closeChan:
	}
}

func (n *fNatsScopeTransport) IsOpen() bool {
	n.openMu.RLock()
	defer n.openMu.RUnlock()
	return n.conn.Status() == nats.CONNECTED && n.isOpen
}

func (n *fNatsScopeTransport) getClosedConditionError(prefix string) error {
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.NOT_OPEN,
			fmt.Sprintf("%s NATS client not connected (has status code %d)", prefix, n.conn.Status()))
	}
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s NATS FScopeTransport not open", prefix))
}

// Close unsubscribes in the case of a subscriber and clears the buffer in the
// case of a publisher.
func (n *fNatsScopeTransport) Close() error {
	n.openMu.Lock()
	defer n.openMu.Unlock()
	if !n.isOpen {
		return nil
	}

	if !n.pull {
		n.isOpen = false
		return nil
	}

	if err := n.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	n.sub = nil
	close(n.closeChan)
	n.isOpen = false
	return nil
}

func (n *fNatsScopeTransport) Read(p []byte) (int, error) {
	if !n.IsOpen() {
		return 0, thrift.NewTTransportExceptionFromError(io.EOF)
	}
	if len(n.currentFrame) == 0 {
		select {
		case frame := <-n.frameBuffer:
			n.currentFrame = frame
		case <-n.closeChan:
			return 0, thrift.NewTTransportExceptionFromError(io.EOF)
		}
	}
	num := copy(p, n.currentFrame)
	// TODO: We could be more efficient here. If the provided buffer isn't
	// full, we could attempt to get the next frame.

	n.currentFrame = n.currentFrame[num:]
	return num, nil
}

// DiscardFrame discards the current message frame the transport is reading, if
// any. After calling this, a subsequent call to Read will read from the next
// frame. This must be called from the same goroutine as the goroutine calling
// Read.
func (n *fNatsScopeTransport) DiscardFrame() {
	n.currentFrame = nil
}

// Write bytes to publish. If buffered bytes exceeds 1MB, ErrTooLarge is
// returned.
func (n *fNatsScopeTransport) Write(p []byte) (int, error) {
	// Include 4 bytes for frame size.
	if len(p)+n.writeBuffer.Len()+4 > natsMaxMessageSize {
		n.writeBuffer.Reset() // Clear any existing bytes.
		return 0, ErrTooLarge
	}

	num, err := n.writeBuffer.Write(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

// Flush publishes the buffered message.
func (n *fNatsScopeTransport) Flush() error {
	if !n.IsOpen() {
		return n.getClosedConditionError("flush:")
	}
	defer n.writeBuffer.Reset()
	data := n.writeBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}
	binary.BigEndian.PutUint32(n.sizeBuffer, uint32(len(data)))
	err := n.conn.Publish(n.formattedSubject(), append(n.sizeBuffer, data...))
	return thrift.NewTTransportExceptionFromError(err)
}

func (n *fNatsScopeTransport) RemainingBytes() uint64 {
	return ^uint64(0) // We don't know unless framed is used.
}

func (n *fNatsScopeTransport) formattedSubject() string {
	return fmt.Sprintf("%s%s", frugalPrefix, n.subject)
}
