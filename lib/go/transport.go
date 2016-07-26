package frugal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
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
// allows a way for setting an FTransportMonitor and a high-water mark provided
// by an FServer.
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

	// SetHighWatermark sets the maximum amount of time a frame is allowed to
	// await processing before triggering transport overload logic.
	// DEPRECATED - This will be a constructor implementation detail for
	// transports which buffer response data.
	// TODO: Remove this with 2.0
	SetHighWatermark(watermark time.Duration)
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

	// TODO: Remove these with 2.0
	frameBuffer  chan []byte
	currentFrame []byte
	closeChan    chan struct{}
}

// Initialize a new fBaseTransport
func newFBaseTransport(requestSizeLimit uint) *fBaseTransport {
	return &fBaseTransport{requestSizeLimit: requestSizeLimit}
}

// Intialize a new fBaseTransprot for use with legacy TTransports
// TODO: Remove with 2.0
func newFBaseTransportForTTransport(requestSizeLimit, frameBufferSize uint) *fBaseTransport {
	return &fBaseTransport{
		requestSizeLimit: requestSizeLimit,
		frameBuffer:      make(chan []byte, frameBufferSize),
	}
}

// Intialize the close channels
func (f *fBaseTransport) Open() {
	f.closed = make(chan error, 1)

	// TODO: Remove this with 2.0
	f.closeChan = make(chan struct{})
}

// Close the close channels
func (f *fBaseTransport) Close(cause error) {

	select {
	case f.closed <- cause:
	default:
		log.Warnf("frugal: unable to put close error '%s' on fBaseTransport closed channel", cause)
	}
	close(f.closed)

	// TODO: Remove this with 2.0
	close(f.closeChan)
}

// Return the struct close channel
// TODO: Remove with 2.0
func (f *fBaseTransport) ClosedChannel() <-chan struct{} {
	return f.closeChan
}

// Read up to len(buf) bytes into buf.
// TODO: Remove all read logic with 2.0
func (f *fBaseTransport) Read(buf []byte) (int, error) {
	if len(f.currentFrame) == 0 {
		select {
		case frame := <-f.frameBuffer:
			f.currentFrame = frame
		case <-f.closeChan:
			return 0, thrift.NewTTransportExceptionFromError(io.EOF)
		}
	}
	num := copy(buf, f.currentFrame)
	f.currentFrame = f.currentFrame[num:]
	return num, nil
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

// This is a no-op for fBaseTransport
func (f *fBaseTransport) SetHighWatermark(watermark time.Duration) {
	return
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
