package frugal

import (
	"bytes"
	"encoding/binary"
	"errors"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const (
	// REQUEST_TOO_LARGE is a TTransportException error type indicating the
	// request exceeded the size limit.
	REQUEST_TOO_LARGE = 100

	// RESPONSE_TOO_LARGE is a TTransportException error type indicating the
	// response exceeded the size limit.
	RESPONSE_TOO_LARGE = 101
)

// ErrTransportClosed is returned by service calls when the transport is
// unexpectedly closed, perhaps as a result of the transport entering an
// invalid state. If this is returned, the transport should be reinitialized.
var ErrTransportClosed = errors.New("frugal: transport was unexpectedly closed")

// ErrTooLarge is returned when attempting to write a message which exceeds the
// transport's message size limit.
var ErrTooLarge = thrift.NewTTransportException(REQUEST_TOO_LARGE,
	"request was too large for the transport")

// IsErrTooLarge indicates if the given error is an ErrTooLarge.
func IsErrTooLarge(err error) bool {
	if err == ErrTooLarge {
		return true
	}
	if e, ok := err.(thrift.TTransportException); ok {
		return e.TypeId() == REQUEST_TOO_LARGE || e.TypeId() == RESPONSE_TOO_LARGE
	}
	return false
}

// FScopeTransportFactory produces FScopeTransports and is typically used by an
// FScopeProvider.
type FScopeTransportFactory interface {
	GetTransport() FScopeTransport
}

// FScopeTransport extends Thrift's TTransport and is used exclusively for
// pub/sub scopes. Subscribers use an FScopeTransport to subscribe to a pub/sub
// topic. Publishers use it to publish to a topic.
type FScopeTransport interface {
	thrift.TTransport

	// LockTopic sets the publish topic and locks the transport for exclusive
	// access.
	LockTopic(string) error

	// UnlockTopic unsets the publish topic and unlocks the transport.
	UnlockTopic() error

	// Subscribe sets the subscribe topic and opens the transport.
	Subscribe(string) error

	// DiscardFrame discards the current message frame the transport is
	// reading, if any. After calling this, a subsequent call to Read will read
	// from the next frame. This must be called from the same goroutine as the
	// goroutine calling Read.
	DiscardFrame()
}

// FTransport is Frugal's equivalent of Thrift's TTransport. FTransport extends
// TTransport and exposes some additional methods. An FTransport typically has
// an FRegistry, so it provides methods for setting the FRegistry and
// registering and unregistering an FAsyncCallback to an FContext. It also
// allows a way for setting an FTransportMonitor.
//
// FTransport wraps a TTransport, meaning all existing TTransport
// implementations will work in Frugal. However, all FTransports must used a
// framed protocol, typically implemented by wrapping a TFramedTransport.
//
// Most Frugal language libraries include an FMuxTransport implementation,
// which uses a worker pool to handle messages in parallel.
type FTransport interface {
	thrift.TTransport

	// SetRegistry sets the Registry on the FTransport.
	SetRegistry(FRegistry)

	// Register a callback for the given Context.
	Register(*FContext, FAsyncCallback) error

	// Unregister a callback for the given Context.
	Unregister(*FContext)

	// SetMonitor starts a monitor that can watch the health of, and reopen,
	// the transport.
	SetMonitor(FTransportMonitor)

	// Closed channel receives the cause of an FTransport close (nil if clean
	// close).
	Closed() <-chan error
}

// FTransportFactory produces FTransports by wrapping a provided TTransport.
type FTransportFactory interface {
	GetTransport(tr thrift.TTransport) FTransport
}

type fBaseTransport struct {
	requestSizeLimit uint
	writeBuffer      bytes.Buffer
	registry         FRegistry
	closed           chan error
}

// Initialize a new fBaseTransport
func newFBaseTransport(requestSizeLimit uint) *fBaseTransport {
	return &fBaseTransport{requestSizeLimit: requestSizeLimit}
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

// Read should not be called, it will return an error
func (f *fBaseTransport) Read(buf []byte) (int, error) {
	return 0, errors.New("don't call read")
}

// Write the bytes to a buffer. Returns ErrTooLarge if the buffer exceeds the
// request size limit.
func (f *fBaseTransport) Write(buf []byte) (int, error) {
	if f.requestSizeLimit > 0 && len(buf)+f.writeBuffer.Len() > int(f.requestSizeLimit) {
		f.writeBuffer.Reset()
		return 0, ErrTooLarge
	}
	num, err := f.writeBuffer.Write(buf)
	return num, thrift.NewTTransportExceptionFromError(err)
}

func (f *fBaseTransport) RemainingBytes() uint64 {
	return ^uint64(0)
}

// Get the write bytes
func (f *fBaseTransport) GetWriteBytes() []byte {
	return f.writeBuffer.Bytes()
}

// Reset the write buffer
func (f *fBaseTransport) ResetWriteBuffer() {
	f.writeBuffer.Reset()
}

// Execute a frugal frame (NOTE: this frame must include the frame size).
func (f *fBaseTransport) ExecuteFrame(frame []byte) error {
	return f.registry.Execute(frame[4:])
}

// SetRegistry sets the Registry on the FTransport.
func (f *fBaseTransport) SetRegistry(registry FRegistry) {
	if registry == nil {
		panic("frugal: registry cannot be nil")
	}
	if f.registry != nil {
		return
	}
	f.registry = registry
}

// Register a callback for the given Context. Only called by generated code.
func (f *fBaseTransport) Register(ctx *FContext, callback FAsyncCallback) error {
	return f.registry.Register(ctx, callback)
}

// Unregister a callback for the given Context. Only called by generated code.
func (f *fBaseTransport) Unregister(ctx *FContext) {
	f.registry.Unregister(ctx)
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
