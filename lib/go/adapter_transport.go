package frugal

import (
	"errors"
	"io"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
)

type fAdapterTransportFactory struct{}

// NewAdapterTransportFactory creates a new FTransportFactory which produces an
// FTransport implementation that acts as an adapter for thrift.TTransport.
// This allows TTransports which support blocking reads to work with Frugal by
// starting a goroutine that reads from the underlying transport and calling
// the registry on received frames.
func NewAdapterTransportFactory() FTransportFactory {
	return &fAdapterTransportFactory{}
}

// GetTransport returns a new adapter FTransport.
func (f *fAdapterTransportFactory) GetTransport(tr thrift.TTransport) FTransport {
	return NewAdapterTransport(tr)
}

type fAdapterTransport struct {
	*TFramedTransport
	isOpen             bool
	mu                 sync.RWMutex
	closeSignal        chan struct{}
	closeChan          chan error
	monitorCloseSignal chan<- error
	registry           FRegistry
}

// NewAdapterTransport returns an FTransport which uses the given TTransport
// for read/write operations in a way that is compatible with Frugal. This
// allows TTransports which support blocking reads to work with Frugal by
// starting a goroutine that reads from the underlying transport and calling
// the registry on received frames.
func NewAdapterTransport(tr thrift.TTransport) FTransport {
	return &fAdapterTransport{
		TFramedTransport: NewTFramedTransport(tr),
		closeSignal:      make(chan struct{}, 1),
	}
}

func (f *fAdapterTransport) Open() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.isOpen {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN,
			"frugal: transport already open")
	}

	if err := f.TFramedTransport.Open(); err != nil {
		// It's OK if the underlying transport is already open.
		if e, ok := err.(thrift.TTransportException); !(ok && e.TypeId() == thrift.ALREADY_OPEN) {
			return err
		}
	}

	go f.readLoop()
	f.isOpen = true
	f.closeChan = make(chan error, 1)
	return nil
}

func (f *fAdapterTransport) readLoop() {
	for {
		frame, err := f.readFrame()
		if err != nil {
			// First check if the transport was closed.
			select {
			case <-f.closeSignal:
				// Transport was closed.
				return
			default:
			}

			if err, ok := err.(thrift.TTransportException); ok && err.TypeId() == thrift.END_OF_FILE {
				// EOF indicates remote peer disconnected.
				f.Close()
				return
			}

			log.Error("frugal: error reading protocol frame, closing transport: ", err)
			f.close(err)
			return
		}

		if err := f.registry.Execute(frame); err != nil {
			// An error here indicates an unrecoverable error, teardown transport.
			log.Error("frugal: closing transport due to unrecoverable error processing frame: ", err)
			f.close(err)
			return
		}
	}
}

func (f *fAdapterTransport) readFrame() ([]byte, error) {
	_, err := f.TFramedTransport.Read([]byte{})
	if err != nil {
		return nil, err
	}
	buff := make([]byte, f.RemainingBytes())
	_, err = io.ReadFull(f.TFramedTransport, buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (f *fAdapterTransport) IsOpen() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.isOpen && f.TFramedTransport.IsOpen()
}

func (f *fAdapterTransport) Close() error {
	return f.close(nil)
}

func (f *fAdapterTransport) close(cause error) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isOpen {
		return thrift.NewTTransportException(thrift.NOT_OPEN, "Transport not open")
	}

	f.closeSignal <- struct{}{}
	if err := f.TFramedTransport.Close(); err != nil {
		// Close failed, drain close signal.
		select {
		case <-f.closeSignal:
		default:
		}
		return err
	}

	select {
	case f.closeChan <- cause:
	default:
	}
	close(f.closeChan)

	if cause == nil {
		log.Debug("frugal: transport closed")
	} else {
		log.Debugf("frugal: transport closed with cause: %s", cause)
	}

	// Signal transport monitor of close.
	select {
	case f.monitorCloseSignal <- cause:
	default:
	}

	f.isOpen = false
	return nil
}

// Read returns an error as it should not be called directly. The adapter
// transport handles reading from the underlying transport and is responsible
// for invoking callbacks on received frames.
func (f *fAdapterTransport) Read(buf []byte) (int, error) {
	return 0, errors.New("Do not call Read directly on FTransport")
}

func (f *fAdapterTransport) SetRegistry(registry FRegistry) {
	if registry == nil {
		panic("frugal: registry cannot be nil")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.registry != nil {
		panic("frugal: registry already set")
	}
	f.registry = registry
}

func (f *fAdapterTransport) Register(ctx *FContext, cb FAsyncCallback) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.registry.Register(ctx, cb)
}

func (f *fAdapterTransport) Unregister(ctx *FContext) {
	f.mu.RLock()
	f.registry.Unregister(ctx)
	f.mu.RUnlock()
}

func (f *fAdapterTransport) SetMonitor(monitor FTransportMonitor) {
	// Stop the previous monitor, if any.
	select {
	case f.monitorCloseSignal <- nil:
	default:
	}

	// Start the new monitor.
	monitorClosedSignal := make(chan error, 1)
	runner := &monitorRunner{
		monitor:       monitor,
		transport:     f,
		closedChannel: monitorClosedSignal,
	}
	f.monitorCloseSignal = monitorClosedSignal
	go runner.run()
}

func (f *fAdapterTransport) Closed() <-chan error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.closeChan
}

func (f *fAdapterTransport) SetHighWatermark(watermark time.Duration) {
	// No-op
}
