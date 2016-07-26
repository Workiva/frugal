package frugal

import (
	"errors"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFRegistry struct {
	mock.Mock
	executeCalled chan struct{}
}

func (m *mockFRegistry) Register(ctx *FContext, cb FAsyncCallback) error {
	return m.Called(ctx, cb).Error(0)
}

func (m *mockFRegistry) Unregister(ctx *FContext) {
	m.Called(ctx)
}

func (m *mockFRegistry) Execute(frame []byte) error {
	select {
	case m.executeCalled <- struct{}{}:
	default:
	}
	return m.Called(frame).Error(0)
}

// Ensures Open returns an error if the wrapped transport fails to open.
func TestAdapterTransportOpenError(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewAdapterTransport(mockTr)
	err := errors.New("error")
	mockTr.On("Open").Return(err)
	assert.Equal(t, err, tr.Open())
	mockTr.AssertExpectations(t)
}

// Ensures Open returns an ALREADY_OPEN TTransportException if the transport is
// already open.
func TestAdapterTransportAlreadyOpen(t *testing.T) {
	assert := assert.New(t)
	mockTr := new(mockTTransport)
	tr := NewAdapterTransport(mockTr)
	mockTr.On("Open").Return(nil)
	assert.Nil(tr.Open())

	err := tr.Open()

	assert.Error(err)
	trErr, ok := err.(thrift.TTransportException)
	assert.True(ok)
	assert.Equal(thrift.ALREADY_OPEN, trErr.TypeId())
	mockTr.AssertExpectations(t)
}

// Ensures Open opens the underlying transport and starts the read goroutine.
func TestAdapterTransportOpenReadClose(t *testing.T) {
	assert := assert.New(t)
	mockTr := new(mockTTransport)
	mockTr.reads = make(chan []byte, 1)
	mockTr.reads <- frame
	close(mockTr.reads)
	tr := NewAdapterTransport(mockTr)
	mockRegistry := new(mockFRegistry)
	executeCalled := make(chan struct{}, 1)
	mockRegistry.executeCalled = executeCalled
	mockRegistry.On("Execute", frame[4:]).Return(nil)
	tr.SetRegistry(mockRegistry)
	mockTr.On("Open").Return(nil)
	mockTr.On("Close").Return(nil)
	assert.Nil(tr.Open())

	assert.Nil(tr.Close())

	select {
	case err := <-tr.Closed():
		assert.Nil(err)
	default:
		t.Fatal("Expected transport to close")
	}

	assert.False(tr.IsOpen())

	select {
	case <-executeCalled:
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Expected Execute to be called")
	}

	mockTr.AssertExpectations(t)
}

// Ensures the read loop closes the transport when it encounters an error.
func TestAdapterTransportReadError(t *testing.T) {
	assert := assert.New(t)
	mockTr := new(mockTTransport)
	err := errors.New("error")
	mockTr.readError = err
	mockTr.On("Open").Return(nil)
	mockTr.On("Close").Return(nil)
	tr := NewAdapterTransport(mockTr)
	assert.Nil(tr.Open())

	select {
	case reason := <-tr.Closed():
		assert.Equal(err, reason)
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Expected transport to close")
	}

	assert.False(tr.IsOpen())

	mockTr.AssertExpectations(t)
}

// Ensures the read loop closes the transport when registry execute encounters
// an error.
func TestAdapterTransportExecuteError(t *testing.T) {
	assert := assert.New(t)
	mockTr := new(mockTTransport)
	mockTr.reads = make(chan []byte, 1)
	mockTr.reads <- frame
	mockTr.On("Open").Return(nil)
	mockTr.On("Close").Return(nil)
	mockRegistry := new(mockFRegistry)
	executeCalled := make(chan struct{}, 1)
	mockRegistry.executeCalled = executeCalled
	err := errors.New("error")
	mockRegistry.On("Execute", frame[4:]).Return(err)
	tr := NewAdapterTransport(mockTr)
	tr.SetRegistry(mockRegistry)
	assert.Nil(tr.Open())

	select {
	case reason := <-tr.Closed():
		assert.Equal(err, reason)
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Expected transport to close")
	}

	assert.False(tr.IsOpen())

	mockTr.AssertExpectations(t)
	mockRegistry.AssertExpectations(t)
}

// Ensures Close returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestAdapterTransportCloseNotOpen(t *testing.T) {
	assert := assert.New(t)
	mockTr := new(mockTTransport)
	tr := NewAdapterTransport(mockTr)

	err := tr.Close()

	assert.Error(err)
	trErr, ok := err.(thrift.TTransportException)
	assert.True(ok)
	assert.Equal(thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Close returns an error if the underlying transport fails to close and
// the close channel is not signaled.
func TestAdapterTransportCloseError(t *testing.T) {
	assert := assert.New(t)
	mockTr := new(mockTTransport)
	err := errors.New("error")
	tr := NewAdapterTransport(mockTr)
	mockTr.On("Close").Return(err)
	mockTr.On("Open").Return(nil)
	assert.Nil(tr.Open())

	assert.Equal(err, tr.Close())
	select {
	case <-tr.Closed():
		t.Fatal("Transport should not have closed")
	default:
	}

	mockTr.AssertExpectations(t)
}

// Ensures SetMonitor starts the FTransportMonitor and setting another monitor
// triggers the previous monitor's clean close. Closing the transport triggers
// a clean close on the active monitor.
func TestAdapterTransportSetMonitor(t *testing.T) {
	mockTr := new(mockTTransport)
	mockMonitor := new(mockFTransportMonitor)
	tr := NewAdapterTransport(mockTr)
	mockTr.On("Open").Return(nil)
	mockTr.On("Close").Return(nil)
	mockTr.reads = make(chan []byte)
	mockMonitor.On("OnClosedCleanly").Return(nil)

	tr.SetMonitor(mockMonitor)
	assert.Nil(t, tr.Open())
	mockMonitor2 := new(mockFTransportMonitor)
	mockMonitor2.On("OnClosedCleanly").Return(nil)

	// Setting a new monitor should trigger clean close on the previous one.
	tr.SetMonitor(mockMonitor2)
	assert.Nil(t, tr.Close())
	time.Sleep(5 * time.Millisecond)

	mockTr.AssertExpectations(t)
	mockMonitor.AssertExpectations(t)
	mockMonitor2.AssertExpectations(t)
}

// Ensures SetRegistry panics when the registry is nil.
func TestAdapterTransportSetRegistryNilPanic(t *testing.T) {
	tr := NewAdapterTransport(nil)
	defer func() {
		assert.NotNil(t, recover())
	}()
	tr.SetRegistry(nil)
}

// Ensures SetRegistry does nothing if the registry is already set.
func TestAdapterTransportSetRegistryAlreadySet(t *testing.T) {
	registry := NewFClientRegistry()
	tr := NewAdapterTransport(nil)
	tr.SetRegistry(registry)
	assert.Equal(t, registry, tr.(*fAdapterTransport).registry)
	defer func() {
		assert.NotNil(t, recover())
	}()
	tr.SetRegistry(NewServerRegistry(nil, nil, nil))
}

// Ensures a direct Read returns an error.
func TestAdapterTransportDirectReadError(t *testing.T) {
	tr := NewAdapterTransport(nil)
	_, err := tr.Read(make([]byte, 5))
	assert.Error(t, err)
}

// Ensures Register calls through to the registry to register a callback.
func TestAdapterTransportRegister(t *testing.T) {
	tr := NewAdapterTransport(nil)
	mockRegistry := new(mockFRegistry)
	tr.SetRegistry(mockRegistry)
	ctx := NewFContext("")
	cb := func(thrift.TTransport) error {
		return nil
	}
	mockRegistry.On("Register", ctx, mock.AnythingOfType("FAsyncCallback")).Return(nil)

	assert.Nil(t, tr.Register(ctx, cb))

	mockRegistry.AssertExpectations(t)
}

// Ensures Unregister calls through to the registry to unregister a callback.
func TestAdapterTransportUnregister(t *testing.T) {
	tr := NewAdapterTransport(nil)
	mockRegistry := new(mockFRegistry)
	tr.SetRegistry(mockRegistry)
	ctx := NewFContext("")
	mockRegistry.On("Unregister", ctx).Return(nil)

	tr.Unregister(ctx)

	mockRegistry.AssertExpectations(t)
}
