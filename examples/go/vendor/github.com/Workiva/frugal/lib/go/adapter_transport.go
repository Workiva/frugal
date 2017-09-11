/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package frugal

import (
	"bytes"
	"io"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
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
	transport          thrift.TTransport
	isOpen             bool
	mu                 sync.RWMutex
	closeSignal        chan struct{}
	closeChan          chan error
	monitorCloseSignal chan<- error
	registry           fRegistry
}

// NewAdapterTransport returns an FTransport which uses the given TTransport
// for read/write operations in a way that is compatible with Frugal. This
// allows TTransports which support blocking reads to work with Frugal by
// starting a goroutine that reads from the underlying transport and calling
// the registry on received frames.
func NewAdapterTransport(tr thrift.TTransport) FTransport {
	return &fAdapterTransport{
		registry:    newFRegistry(),
		transport:   tr,
		closeSignal: make(chan struct{}, 1),
	}
}

// Open prepares the transport to send data.
func (f *fAdapterTransport) Open() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.isOpen {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_ALREADY_OPEN,
			"frugal: transport already open")
	}

	if err := f.transport.Open(); err != nil {
		// It's OK if the underlying transport is already open.
		if e, ok := err.(thrift.TTransportException); !(ok && e.TypeId() == TRANSPORT_EXCEPTION_ALREADY_OPEN) {
			return err
		}
	}

	go f.readLoop()
	f.isOpen = true
	f.closeChan = make(chan error, 1)
	return nil
}

func (f *fAdapterTransport) readLoop() {
	framedTransport := NewTFramedTransport(f.transport)
	for {
		frame, err := f.readFrame(framedTransport)
		if err != nil {
			// First check if the transport was closed.
			select {
			case <-f.closeSignal:
				// Transport was closed.
				return
			default:
			}

			if err, ok := err.(thrift.TTransportException); ok && err.TypeId() == TRANSPORT_EXCEPTION_END_OF_FILE {
				// EOF indicates remote peer disconnected.
				f.Close()
				return
			}

			logger().Error("frugal: error reading protocol frame, closing transport: ", err)
			f.close(err)
			return
		}

		if err := f.registry.Execute(frame); err != nil {
			// An error here indicates an unrecoverable error, teardown transport.
			logger().Error("frugal: closing transport due to unrecoverable error processing frame: ", err)
			f.close(err)
			return
		}
	}
}

func (f *fAdapterTransport) readFrame(framedTransport *TFramedTransport) ([]byte, error) {
	_, err := framedTransport.Read([]byte{})
	if err != nil {
		return nil, err
	}
	buff := make([]byte, framedTransport.RemainingBytes())
	_, err = io.ReadFull(framedTransport, buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

// IsOpen returns true if the transport is open, false otherwise.
func (f *fAdapterTransport) IsOpen() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.isOpen && f.transport.IsOpen()
}

// Close closes the transport.
func (f *fAdapterTransport) Close() error {
	return f.close(nil)
}

func (f *fAdapterTransport) close(cause error) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isOpen {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_NOT_OPEN, "Transport not open")
	}

	f.closeSignal <- struct{}{}
	if err := f.transport.Close(); err != nil {
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
		logger().Debug("frugal: transport closed")
	} else {
		logger().Debugf("frugal: transport closed with cause: %s", cause)
	}

	// Signal transport monitor of close.
	select {
	case f.monitorCloseSignal <- cause:
	default:
	}

	f.isOpen = false
	return nil
}

// Oneway transmits the given data and doesn't wait for a response.
// Implementations of oneway should be threadsafe and respect the timeout
// present on the context.
func (f *fAdapterTransport) Oneway(ctx FContext, payload []byte) error {
	errorC := make(chan error, 1)
	go f.send(payload, errorC, true)

	select {
	case err := <-errorC:
		return err
	case <-time.After(ctx.Timeout()):
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_TIMED_OUT, "frugal: request timed out")
	}
}

// Request transmits the given data and waits for a response.
// Implementations of request should be threadsafe and respect the timeout
// present on the context.
func (f *fAdapterTransport) Request(ctx FContext, payload []byte) (thrift.TTransport, error) {
	resultC := make(chan []byte, 1)
	errorC := make(chan error, 1)

	f.registry.Register(ctx, resultC)
	defer f.registry.Unregister(ctx)

	go f.send(payload, errorC, false)

	select {
	case result := <-resultC:
		return &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(result)}, nil
	case err := <-errorC:
		return nil, err
	case <-time.After(ctx.Timeout()):
		return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_TIMED_OUT, "frugal: request timed out")
	}
}

func (f *fAdapterTransport) send(payload []byte, errorC chan error, oneway bool) {
	// TODO: does this need to be called in a goroutine?
	// i.e. can Write() and Flush() block?
	if _, err := f.transport.Write(payload); err != nil {
		errorC <- err
		return
	}
	if err := f.transport.Flush(); err != nil {
		errorC <- err
		return
	}

	if oneway {
		// If it's a oneway, no result will be sent back from the server
		// so let the goroutine know everything succeeded
		errorC <- nil
	}
}

// GetRequestSizeLimit returns the maximum number of bytes that can be
// transmitted. Returns a non-positive number to indicate an unbounded
// allowable size.
func (f *fAdapterTransport) GetRequestSizeLimit() uint {
	return 0
}

// SetMonitor starts a monitor that can watch the health of, and reopen,
// the transport.
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

// Closed channel receives the cause of an FTransport close (nil if clean
// close).
func (f *fAdapterTransport) Closed() <-chan error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.closeChan
}
