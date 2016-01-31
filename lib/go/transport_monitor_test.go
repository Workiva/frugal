package frugal

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testTimeout = 10 * time.Millisecond

// Ensure that OnClosedUncleanly returns false if max attempts is 0.
func TestOnClosedUncleanlyMaxZero(t *testing.T) {
	m := &BaseFTransportMonitor{}
	retry, _ := m.OnClosedUncleanly(errors.New("respect my authoritah"))
	require.Equal(t, false, retry)
}

// Ensure that OnClosedUncleanly returns true if max attempts is > 0.
func TestOnClosedUncleanly(t *testing.T) {
	m := &BaseFTransportMonitor{MaxReopenAttempts: 1, InitialWait: time.Millisecond}
	retry, wait := m.OnClosedUncleanly(errors.New("but meuoom"))
	require.Equal(t, true, retry)
	require.Equal(t, time.Millisecond, wait)
}

// Ensure that OnReopenFailed returns false if max attempts reached.
func TestOnReopenFailedMaxAttempts(t *testing.T) {
	m := &BaseFTransportMonitor{}
	retry, _ := m.OnReopenFailed(5, time.Millisecond)
	require.Equal(t, false, retry)
}

// Ensure that OnReopenFailed returns true plus double the previous wait.
func TestOnReopenFailed(t *testing.T) {
	m := &BaseFTransportMonitor{MaxReopenAttempts: 6, MaxWait: time.Hour}
	retry, wait := m.OnReopenFailed(5, time.Millisecond)
	require.Equal(t, true, retry)
	require.Equal(t, 2*time.Millisecond, wait)
}

// Ensure that OnReopenFailed respects the max wait.
func TestOnReopenMaxWait(t *testing.T) {
	m := &BaseFTransportMonitor{MaxReopenAttempts: 6, MaxWait: time.Millisecond}
	retry, wait := m.OnReopenFailed(5, time.Millisecond)
	require.Equal(t, true, retry)
	require.Equal(t, time.Millisecond, wait)
}

// Ensure that run attempts to handle a clean close.
func TestRunCleanClose(t *testing.T) {
	ftm := &mockFTransportMonitor{}
	ftm.On("OnClosedCleanly").Return()
	closedChannel := make(chan error, 1)
	closedChannel <- nil
	r := monitorRunner{
		monitor:       ftm,
		closedChannel: closedChannel,
	}

	runExited := make(chan struct{})
	go func() {
		r.run()
		runExited <- struct{}{}
	}()

	select {
	case <-runExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	ftm.AssertExpectations(t)
}

// Ensure that run attempts to handle an unclean close.
func TestRunUncleanClose(t *testing.T) {
	cause := errors.New("pooped myself")
	ftm := &mockFTransportMonitor{}
	ftm.On("OnClosedUncleanly", cause).Return(false, time.Duration(0)).Once()
	closedChannel := make(chan error, 1)
	closedChannel <- cause
	r := monitorRunner{
		monitor:       ftm,
		closedChannel: closedChannel,
	}

	runExited := make(chan struct{})
	go func() {
		r.run()
		runExited <- struct{}{}
	}()

	select {
	case <-runExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	ftm.AssertExpectations(t)
}

// Ensure that handleCleanClose invokes OnClosedCleanly
func TestHandleCleanClose(t *testing.T) {
	ftm := &mockFTransportMonitor{}
	ftm.On("OnClosedCleanly").Return().Once()
	r := monitorRunner{monitor: ftm}

	handleCleanCloseExited := make(chan struct{})
	go func() {
		r.handleCleanClose()
		handleCleanCloseExited <- struct{}{}
	}()

	select {
	case <-handleCleanCloseExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	ftm.AssertExpectations(t)
}

// Ensure that handleUncleanClose returns false when OnClosedUncleanly instructs not to retry.
func TestHandleUncleanCloseNoRetry(t *testing.T) {
	cause := errors.New("Attacked by cosmic ray")
	ftm := &mockFTransportMonitor{}
	ftm.On("OnClosedUncleanly", cause).Return(false, time.Duration(0)).Once()
	r := monitorRunner{monitor: ftm}

	handleUncleanCloseExited := make(chan struct{})
	go func() {
		assert.False(t, r.handleUncleanClose(cause))
		handleUncleanCloseExited <- struct{}{}
	}()

	select {
	case <-handleUncleanCloseExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	ftm.AssertExpectations(t)
}

// Ensure that handleUncleanClose attempts re-open when OnClosedUncleanly instructs it to do so.
func TestHandleUncleanClose(t *testing.T) {
	cause := errors.New("I fart in your general direction")
	ftm := &mockFTransportMonitor{}
	ftm.On("OnClosedUncleanly", cause).Return(true, time.Duration(0)).Once()
	ftm.On("OnReopenSucceeded").Return().Once()
	mft := &mockFTransport{}
	mft.On("Open").Return(nil).Once()
	r := monitorRunner{
		monitor:   ftm,
		transport: mft,
	}

	handleUncleanCloseExited := make(chan struct{})
	go func() {
		assert.True(t, r.handleUncleanClose(cause))
		handleUncleanCloseExited <- struct{}{}
	}()

	select {
	case <-handleUncleanCloseExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	ftm.AssertExpectations(t)
	mft.AssertExpectations(t)
}

// Ensure that attemptReopen returns true when re-opening succeeds.
func TestAttemptReopenSuccess(t *testing.T) {
	ftm := &mockFTransportMonitor{}
	ftm.On("OnReopenSucceeded").Return().Once()
	mft := &mockFTransport{}
	mft.On("Open").Return(nil).Once()
	r := monitorRunner{
		monitor:   ftm,
		transport: mft,
	}

	attemptReopenExited := make(chan struct{})
	go func() {
		assert.True(t, r.attemptReopen(0))
		attemptReopenExited <- struct{}{}
	}()

	select {
	case <-attemptReopenExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	ftm.AssertExpectations(t)
	mft.AssertExpectations(t)
}

// Ensure that attemptReopen retries when OnReopenFailed instructs to do so.
func TestAttemptReopenFailRetrySucceed(t *testing.T) {
	ftm := &mockFTransportMonitor{}
	ftm.On("OnReopenFailed", uint(1), time.Nanosecond).Return(true, 2*time.Nanosecond).Once()
	ftm.On("OnReopenSucceeded").Return().Once()
	mft := &mockFTransport{}
	mft.On("Open").Return(errors.New("Tears of a thousand children")).Once()
	mft.On("Open").Return(nil).Once()
	r := monitorRunner{
		monitor:   ftm,
		transport: mft,
	}

	attemptReopenExited := make(chan struct{})
	go func() {
		assert.True(t, r.attemptReopen(time.Nanosecond))
		attemptReopenExited <- struct{}{}
	}()

	select {
	case <-attemptReopenExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	ftm.AssertExpectations(t)
	mft.AssertExpectations(t)
}

// Ensure that attemptReopen returns false and does not retry when OnReopenFailed instructs it not to do so.
func TestAttemptReopenFailNoRetry(t *testing.T) {
	ftm := &mockFTransportMonitor{}
	ftm.On("OnReopenFailed", uint(1), time.Nanosecond).Return(false, time.Duration(0)).Once()
	mft := &mockFTransport{}
	mft.On("Open").Return(errors.New("Shattered dreams of a PhD student")).Once()
	r := monitorRunner{
		monitor:   ftm,
		transport: mft,
	}

	attemptReopenExited := make(chan struct{})
	go func() {
		assert.False(t, r.attemptReopen(time.Nanosecond))
		attemptReopenExited <- struct{}{}
	}()

	select {
	case <-attemptReopenExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	ftm.AssertExpectations(t)
	mft.AssertExpectations(t)
}

type mockFTransportMonitor struct {
	mock.Mock
}

func (m *mockFTransportMonitor) OnClosedCleanly() {
	m.Called()
}

func (m *mockFTransportMonitor) OnClosedUncleanly(cause error) (bool, time.Duration) {
	args := m.Called(cause)
	return args.Get(0).(bool), args.Get(1).(time.Duration)
}

func (m *mockFTransportMonitor) OnReopenFailed(prevAttempts uint, prevWait time.Duration) (bool, time.Duration) {
	args := m.Called(prevAttempts, prevWait)
	return args.Get(0).(bool), args.Get(1).(time.Duration)
}

func (m *mockFTransportMonitor) OnReopenSucceeded() {
	m.Called()
}

type mockFTransport struct {
	mock.Mock
}

func (m *mockFTransport) Open() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockFTransport) IsOpen() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *mockFTransport) RemainingBytes() (num_bytes uint64) {
	args := m.Called()
	return args.Get(0).(uint64)
}

func (m *mockFTransport) Flush() (err error) {
	args := m.Called()
	return args.Error(0)
}

func (m *mockFTransport) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Get(0).(int), args.Error(1)
}

func (m *mockFTransport) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Get(0).(int), args.Error(1)
}

func (m *mockFTransport) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockFTransport) SetRegistry(fr FRegistry) {
	m.Called(fr)
}

func (m *mockFTransport) Register(fc *FContext, fac FAsyncCallback) error {
	args := m.Called(fc, fac)
	return args.Error(0)
}

func (m *mockFTransport) Unregister(fc *FContext) {
	m.Called(fc)
}

func (m *mockFTransport) Closed() <-chan error {
	args := m.Called()
	return args.Get(0).(<-chan error)
}

func (m *mockFTransport) SetMonitor(mon FTransportMonitor) {
	m.Called(mon)
}
