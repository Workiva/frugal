package frugal

import (
	"errors"
	"io"
	"log"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const (
	REQUEST_TOO_LARGE  = 100
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

// FScopeTransportFactory produces FScopeTransports which are used by pub/sub
// scopes.
type FScopeTransportFactory interface {
	GetTransport() FScopeTransport
}

// FScopeTransport is a TTransport extension for pub/sub scopes.
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

// FTransport is a TTransport for services.
type FTransport interface {
	thrift.TTransport

	// SetRegistry sets the Registry on the FTransport.
	SetRegistry(FRegistry)

	// Register a callback for the given Context.
	Register(*FContext, FAsyncCallback) error

	// Unregister a callback for the given Context.
	Unregister(*FContext)

	// SetMonitor starts a monitor that can watch the health of, and reopen, the transport.
	SetMonitor(FTransportMonitor)

	// Closed channel receives the cause of an FTransport close (nil if clean close).
	Closed() <-chan error
}

// FTransportFactory produces FTransports which are used by services.
type FTransportFactory interface {
	GetTransport(tr thrift.TTransport) FTransport
}

type fMuxTransportFactory struct {
	numWorkers uint
}

// NewFMuxTransportFactory creates a new FTransportFactory which produces
// multiplexed FTransports. The numWorkers argument specifies the number of
// goroutines to use to process requests concurrently.
func NewFMuxTransportFactory(numWorkers uint) FTransportFactory {
	return &fMuxTransportFactory{numWorkers: numWorkers}
}

func (f *fMuxTransportFactory) GetTransport(tr thrift.TTransport) FTransport {
	return NewFMuxTransport(tr, f.numWorkers)
}

type fMuxTransport struct {
	*TFramedTransport
	registry            FRegistry
	numWorkers          uint
	workC               chan []byte
	open                bool
	mu                  sync.Mutex
	closed              chan error
	monitorClosedSignal chan<- error
}

// NewFMuxTransport wraps the given TTransport in a multiplexed FTransport. The
// numWorkers argument specifies the number of goroutines processing
// requests concurrently.
func NewFMuxTransport(tr thrift.TTransport, numWorkers uint) FTransport {
	if numWorkers == 0 {
		numWorkers = 1
	}
	return &fMuxTransport{
		TFramedTransport: NewTFramedTransport(tr),
		numWorkers:       numWorkers,
		workC:            make(chan []byte, numWorkers),
	}
}

func (f *fMuxTransport) SetMonitor(monitor FTransportMonitor) {
	// Stop the previous monitor, if any
	select {
	case f.monitorClosedSignal <- nil:
	default:
	}

	// Start the new monitor
	monitorClosedSignal := make(chan error, 1)
	runner := &monitorRunner{
		monitor:       monitor,
		transport:     f,
		closedChannel: monitorClosedSignal,
	}
	f.monitorClosedSignal = monitorClosedSignal
	go runner.run()
}

// SetRegistry sets the Registry on the FTransport.
func (f *fMuxTransport) SetRegistry(registry FRegistry) {
	if registry == nil {
		panic("frugal: registry cannot be nil")
	}
	f.mu.Lock()
	if f.registry != nil {
		f.mu.Unlock()
		return
	}
	f.registry = registry
	f.mu.Unlock()
}

// Register a callback for the given Context. Only called by generated code.
func (f *fMuxTransport) Register(ctx *FContext, callback FAsyncCallback) error {
	return f.registry.Register(ctx, callback)
}

// Unregister a callback for the given Context. Only called by generated code.
func (f *fMuxTransport) Unregister(ctx *FContext) {
	f.registry.Unregister(ctx)
}

// Open will open the underlying TTransport and start a goroutine which reads
// from the transport and places the read frames into a work channel.
func (f *fMuxTransport) Open() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.open {
		return errors.New("frugal: transport already open")
	}

	f.closed = make(chan error, 1)

	if err := f.TFramedTransport.Open(); err != nil {
		// It's OK if the underlying transport is already open.
		if e, ok := err.(thrift.TTransportException); !(ok && e.TypeId() == thrift.ALREADY_OPEN) {
			return err
		}
	}

	go func() {
		for {
			frame, err := f.readFrame()
			if err != nil {
				defer f.close(err)
				if err, ok := err.(thrift.TTransportException); ok && err.TypeId() == thrift.END_OF_FILE {
					// EOF indicates remote peer disconnected.
					return
				}
				if !f.IsOpen() {
					// Indicates the transport was closed.
					return
				}
				log.Println("frugal: error reading protocol frame, closing transport:", err)
				return
			}

			select {
			case f.workC <- frame:
			case <-f.closedChan():
				return
			}
		}
	}()

	f.startWorkers()

	f.open = true
	return nil
}

// Close will close the underlying TTransport and stops all goroutines.
func (f *fMuxTransport) Close() error {
	return f.close(nil)
}

func (f *fMuxTransport) close(cause error) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.open {
		return errors.New("frugal: transport not open")
	}

	if err := f.TFramedTransport.Close(); err != nil {
		return err
	}

	f.open = false
	select {
	case f.closed <- cause:
	default:
	}
	close(f.closed)

	// Signal transport monitor of close.
	select {
	case f.monitorClosedSignal <- cause:
	default:
	}

	return nil
}

// Closed channel is closed when the FTransport is closed.
func (f *fMuxTransport) Closed() <-chan error {
	return f.closedChan()
}

func (f *fMuxTransport) readFrame() ([]byte, error) {
	_, err := f.Read([]byte{})
	if err != nil {
		return nil, err
	}
	buff := make([]byte, f.RemainingBytes())
	_, err = io.ReadFull(f, buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (f *fMuxTransport) startWorkers() {
	for i := uint(0); i < f.numWorkers; i++ {
		go func() {
			for {
				select {
				case <-f.closedChan():
					return
				case frame := <-f.workC:
					if err := f.registry.Execute(frame); err != nil {
						// An error here indicates an unrecoverable error, teardown transport.
						log.Println("frugal: closing transport due to unrecoverable error processing frame:", err)
						f.close(err)
						return
					}
				}
			}
		}()
	}
}

func (f *fMuxTransport) closedChan() <-chan error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}
