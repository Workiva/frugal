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
	channels      map[uint64]chan []byte
}

func (m *mockFRegistry) AssignOpID(ctx FContext) error {
	return m.Called(ctx).Error(0)
}

func (m *mockFRegistry) Register(ctx FContext, resultC chan []byte) error {
	opID, err := getOpID(ctx)
	if err == nil {
		m.channels[opID] = resultC
	}

	return m.Called(ctx, resultC).Error(0)
}

func (m *mockFRegistry) Unregister(ctx FContext) {
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
	assert.Equal(TRANSPORT_EXCEPTION_ALREADY_OPEN, trErr.TypeId())
	mockTr.AssertExpectations(t)
}

// Ensures Open opens the underlying transport and starts the read goroutine.
func TestAdapterTransportOpenReadClose(t *testing.T) {
	assert := assert.New(t)
	mockTr := new(mockTTransport)
	mockTr.reads = make(chan []byte, 1)
	mockTr.reads <- frame
	close(mockTr.reads)
	tr := NewAdapterTransport(mockTr).(*fAdapterTransport)
	mockRegistry := new(mockFRegistry)
	executeCalled := make(chan struct{}, 1)
	mockRegistry.executeCalled = executeCalled
	mockRegistry.On("Execute", frame[4:]).Return(nil)
	tr.registry = mockRegistry
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
	case <-time.After(50 * time.Millisecond):
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
	tr := NewAdapterTransport(mockTr).(*fAdapterTransport)
	tr.registry = mockRegistry
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
	assert.Equal(TRANSPORT_EXCEPTION_NOT_OPEN, trErr.TypeId())
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
	time.Sleep(50 * time.Millisecond)

	mockTr.AssertExpectations(t)
	mockMonitor.AssertExpectations(t)
	mockMonitor2.AssertExpectations(t)
}
