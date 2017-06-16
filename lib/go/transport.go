package frugal

import (
	"bytes"
	"encoding/binary"
	//"errors"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// FPublisherTransportFactory produces FPublisherTransports and is typically
// used by an FScopeProvider.
type FPublisherTransportFactory interface {
	GetTransport() FPublisherTransport
}

// FPublisherTransport is used exclusively for pub/sub scopes. Publishers use it
// to publish messages to a topic.
type FPublisherTransport interface {
	// Open opens the transport.
	Open() error

	// Close closes the transport.
	Close() error

	// IsOpen returns true if the transport is open, false otherwise.
	IsOpen() bool

	// GetPublishSizeLimit returns the maximum allowable size of a payload
	// to be published. A non-positive number is returned to indicate an
	// unbounded allowable size.
	GetPublishSizeLimit() uint

	// Publish sends the given payload with the transport. Implementations
	// of publish should be threadsafe.
	Publish(string, []byte) error
}

// FSubscriberTransportFactory produces FSubscriberTransports and is typically
// used by an FScopeProvider.
type FSubscriberTransportFactory interface {
	GetTransport() FSubscriberTransport
}

// FSubscriberTransport is used exclusively for pub/sub scopes. Subscribers use
// it to subscribe to a pub/sub topic.
type FSubscriberTransport interface {
	// Subscribe opens the transport and sets the subscribe topic.
	Subscribe(string, FAsyncCallback) error

	// Unsubscribe unsubscribes from the topic and closes the transport.
	Unsubscribe() error

	// IsSubscribed returns true if the transport is subscribed to a topic,
	// false otherwise.
	IsSubscribed() bool
}

type FUnifiedPublisherTransportFactory interface {
	GetTransport() FUnifiedPublisherTransport
}

type FUnifiedPublisherTransport interface {
	Open() error

	Close() error

	IsOpen() bool

	GetPublishSizeLimit() uint

	PublishEphemeral(string, []byte) error

	PublishDurable(string, *string, []byte) error
}

type FUnifiedSubscriberTransportFactory interface {
	GetTransport() FUnifiedSubscriberTransport
}

type FUnifiedSubscriberTransport interface {
	Subscribe(string, FAsyncCallback) error

	Unsubscribe() error

	IsSubscribed() bool
}

// FDurablePublisherTransportFactory produces FDurablePublisherTransports and is
// typically used by an FDurableScopeProvider.
type FDurablePublisherTransportFactory interface {
	GetTransport() FDurablePublisherTransport
}

// FDurableSubscriberTransportFactory produces FDurableSubscriberTransports and
// is typically used by an FDurableScopeProvider.
type FDurableSubscriberTransportFactory interface {
	GetTransport() FDurableSubscriberTransport
}

// FDurableAsyncCallback is an internal callback which is constructed by
// generated code and invoked by an FDurableSubscriberTransport when a message
// is received.
type FDurableAsyncCallback func(thrift.TTransport, *string) error

// FDurablePublisherTransport is used exclusively for pub/sub scopes. Publishers
// use it to publish messages durably to a topic.
type FDurablePublisherTransport interface {
	// Open opens the transport.
	Open() error

	// Close closes the transport.
	Close() error

	// IsOpen returns true if the transport is open, false otherwise.
	IsOpen() bool

	// GetPublishSizeLimit returns the maximum allowable size of a payload
	// to be published. A non-positive number is returned to indicate an
	// unbounded allowable size.
	GetPublishSizeLimit() uint

	// Publish sends the given payload with the transport. Implementations
	// of publish should be threadsafe. The parameter groupID allows you
	// to assign arbitrary groupings to messages. A value of nil indicates
	// the message isn't part of any group.
	Publish(subejct string, groupID *string, payload []byte) error
}

// FDurableSubscriberTransport is used exclusively for pub/sub scopes.
// Subscribers use it to durably subscribe to a pub/sub topic.
type FDurableSubscriberTransport interface {
	// Subscribe opens the transport and sets the subscribe topic.
	Subscribe(string, FDurableAsyncCallback) error

	// Unsubscribe unsubscribes from the topic and closes the transport.
	Unsubscribe() error

	// IsSubscribed returns true if the transport is subscribed to a topic,
	// false otherwise.
	IsSubscribed() bool
}

// FTransport is Frugal's equivalent of Thrift's TTransport. FTransport is
// comparable to Thrift's TTransport in that it represents the transport layer
// for frugal clients. However, frugal is callback based and sends only framed
// data. Due to this, instead of read, write, and flush methods, FTransport has
// a send method that sends framed frugal messages. To handle callback data, an
// FTransport also has an FRegistry, so it provides methods for registering
// and unregistering an FAsyncCallback to an FContext.
type FTransport interface {
	// SetMonitor starts a monitor that can watch the health of, and reopen,
	// the transport.
	SetMonitor(FTransportMonitor)

	// Closed channel receives the cause of an FTransport close (nil if clean
	// close).
	Closed() <-chan error

	// Open prepares the transport to send data.
	Open() error

	// IsOpen returns true if the transport is open, false otherwise.
	IsOpen() bool

	// Close closes the transport.
	Close() error

	// Oneway transmits the given data and doesn't wait for a response.
	// Implementations of oneway should be threadsafe and respect the timeout
	// present on the context.
	Oneway(ctx FContext, payload []byte) error

	// Request transmits the given data and waits for a response.
	// Implementations of request should be threadsafe and respect the timeout
	// present on the context.
	Request(ctx FContext, payload []byte) (thrift.TTransport, error)

	// GetRequestSizeLimit returns the maximum number of bytes that can be
	// transmitted. Returns a non-positive number to indicate an unbounded
	// allowable size.
	GetRequestSizeLimit() uint
}

// FTransportFactory produces FTransports by wrapping a provided TTransport.
type FTransportFactory interface {
	GetTransport(tr thrift.TTransport) FTransport
}

type fBaseTransport struct {
	requestSizeLimit uint
	writeBuffer      bytes.Buffer
	registry         fRegistry
	closed           chan error
}

// Initialize a new fBaseTransport
func newFBaseTransport(requestSizeLimit uint) *fBaseTransport {
	return &fBaseTransport{
		requestSizeLimit: requestSizeLimit,
		registry:         newFRegistry(),
	}
}

// Intialize the close channels
func (f *fBaseTransport) Open() {
	f.closed = make(chan error, 1)
}

// Close the close channels
func (f *fBaseTransport) Close(cause error) {
	select {
	case f.closed <- cause:
	default:
		logger().Warnf("frugal: unable to put close error '%s' on fBaseTransport closed channel", cause)
	}
	close(f.closed)
}

// Execute a frugal frame (NOTE: this frame must include the frame size).
func (f *fBaseTransport) ExecuteFrame(frame []byte) error {
	return f.registry.Execute(frame[4:])
}

// Closed channel is closed when the FTransport is closed.
func (f *fBaseTransport) Closed() <-chan error {
	return f.closed
}

func prependFrameSize(buf []byte) []byte {
	frame := make([]byte, 4)
	binary.BigEndian.PutUint32(frame, uint32(len(buf)))
	return append(frame, buf...)
}
