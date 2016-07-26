package frugal

import (
	"errors"
	"fmt"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
)

// NewFNatsTransport returns a new FTransport which uses the NATS messaging
// system as the underlying transport. This FTransport is stateless in that
// there is no connection maintained between the client and server. A request
// is simply published to a subject and responses are received on another
// subject. This requires requests and responses to fit within a single NATS
// message.
func NewFNatsTransport(conn *nats.Conn, subject, inbox string) FTransport {
	if inbox == "" {
		inbox = nats.NewInbox()
	}
	return &fNatsTransport{
		// FTransports manually frame messages.
		// Leave enough room for frame size.
		fBaseTransport: newFBaseTransport(natsMaxMessageSize - 4),
		conn:           conn,
		subject:        subject,
		inbox:          inbox,
	}
}

// NewStatelessNatsTTransport returns a new Thrift TTransport which uses the
// NATS messaging system as the underlying transport. Unlike the TTransport
// created by NewNatsServiceTTransport, this TTransport is stateless in that
// there is no connection maintained between the client and server. A request
// is simply published to a subject and responses are received on another
// subject. This requires requests and responses to fit within a single NATS
// message.
// DEPRECATED - Use NewNatsFTransport to create an FTransport directly.
// TODO: Remove this with 2.0
func NewStatelessNatsTTransport(conn *nats.Conn, subject, inbox string) thrift.TTransport {
	if inbox == "" {
		inbox = nats.NewInbox()
	}
	return &fNatsTransport{
		fBaseTransport: newFBaseTransportForTTransport(natsMaxMessageSize, frameBufferSize),
		conn:           conn,
		subject:        subject,
		inbox:          inbox,
		isTTransport:   true,
	}
}

// fNatsTransport implements FTransport and, until the next major release,
// thrift.TTransport that may be wrapped with fMuxTransport (DEPRECATED). This
// is a "stateless" transport in the sense that there is no connection with a
// server. A request is simply published to a subject and responses are
// received on another subject. This assumes requests/responses fit within a
// single NATS message.
type fNatsTransport struct {
	*fBaseTransport
	conn    *nats.Conn
	subject string
	inbox   string
	sub     *nats.Subscription

	// TODO: Remove these with 2.0
	isTTransport bool
	mu           sync.RWMutex
}

// Open subscribes to the configured inbox subject.
func (f *fNatsTransport) Open() error {
	// TODO: Remove locking with 2.0
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: NATS not connected, has status %d", f.conn.Status()))
	}
	if f.sub != nil {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS transport already open")
	}

	handler := f.handler
	// TODO: Remove this with 2.0
	if f.isTTransport {
		handler = f.tTransportHandler
	}

	sub, err := f.conn.Subscribe(f.inbox, handler)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = sub

	// TODO: Remove this with 2.0
	f.fBaseTransport.Open()
	return nil
}

// handler receives a NATS message and executes the frame
func (f *fNatsTransport) handler(msg *nats.Msg) {
	if err := f.fBaseTransport.ExecuteFrame(msg.Data); err != nil {
		log.Warn("Could not execute frame", err)
	}
}

// tTransportHandler receives a NATS message and places it on the frame buffer
// for reading.
// TODO: Remove this with 2.0
func (f *fNatsTransport) tTransportHandler(msg *nats.Msg) {
	select {
	case f.frameBuffer <- msg.Data:
	case <-f.fBaseTransport.ClosedChannel():
	}
}

// Returns true if the transport is open
func (f *fNatsTransport) IsOpen() bool {
	// TODO: Remove locking with 2.0
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.sub != nil && f.conn.Status() == nats.CONNECTED
}

// Close unsubscribes from the inbox subject.
func (f *fNatsTransport) Close() error {
	// TODO: Remove locking with 2.0
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.sub == nil {
		return nil
	}
	if err := f.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = nil

	f.fBaseTransport.Close(nil)

	return nil
}

// Read up to len(buf) bytes into buf.
// TODO: This should just return an error with 2.0
func (f *fNatsTransport) Read(buf []byte) (int, error) {
	if !f.isTTransport {
		return 0, errors.New("Cannot read on FTransport")
	}

	// TODO: Remove all read logic with 2.0
	if !f.IsOpen() {
		return 0, f.getClosedConditionError("read:")
	}
	return f.fBaseTransport.Read(buf)
}

// Write the bytes to a buffer. Returns ErrTooLarge if the buffer exceeds 1MB.
func (f *fNatsTransport) Write(buf []byte) (int, error) {
	if !f.IsOpen() {
		return 0, f.getClosedConditionError("write:")
	}
	return f.fBaseTransport.Write(buf)
}

// Flush sends the buffered bytes over NATS.
func (f *fNatsTransport) Flush() error {
	if !f.IsOpen() {
		return f.getClosedConditionError("flush:")
	}
	data := f.GetWriteBytes()
	if len(data) == 0 {
		return nil
	}

	f.ResetWriteBuffer()
	// TODO: Remove this check in 2.0
	if !f.isTTransport {
		data = prependFrameSize(data)
	}

	err := f.conn.PublishRequest(f.subject, f.inbox, data)
	return thrift.NewTTransportExceptionFromError(err)
}

// This is a no-op for fNatsTransport
func (f *fNatsTransport) SetMonitor(monitor FTransportMonitor) {
}

func (f *fNatsTransport) getClosedConditionError(prefix string) error {
	if f.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.NOT_OPEN,
			fmt.Sprintf("%s stateless NATS client not connected (has status code %d)", prefix, f.conn.Status()))
	}
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s stateless NATS service TTransport not open", prefix))
}
