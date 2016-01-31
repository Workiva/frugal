package frugal

import (
	"fmt"
	"time"
)

// FTransportMonitor watches and heals an FTransport.
type FTransportMonitor interface {
	// OnClosedCleanly is called when the transport is closed cleanly by a call to Close()
	OnClosedCleanly()

	// OnClosedUncleanly is called when the transport is closed for a reason *other* than a call to Close().
	// Returns whether to try reopening the transport and, if so, how long to wait before making the attempt.
	OnClosedUncleanly(cause error) (reopen bool, wait time.Duration)

	// OnReopenFailed is called when an attempt to reopen the transport fails.
	// Given the number of previous attempts to re-open the transport and the length of the previous wait,
	// Returns whether to attempt to re-open the transport, and how long to wait before making the attempt.
	OnReopenFailed(prevAttempts uint, prevWait time.Duration) (reopen bool, wait time.Duration)

	// OnReopenSucceeded is called after the transport has been successfully re-opened.
	OnReopenSucceeded()
}

// BaseFTransportMonitor is a default monitor implementation that attempts to re-open a closed transport
// with exponential backoff behavior and a capped number of retries. Its behavior can be customized
// by embedding this struct type in a new struct which "overrides" desired callbacks.
type BaseFTransportMonitor struct {
	MaxReopenAttempts uint
	InitialWait       time.Duration
	MaxWait           time.Duration
}

func (m *BaseFTransportMonitor) OnClosedUncleanly(cause error) (bool, time.Duration) {
	return m.MaxReopenAttempts > 0, m.InitialWait
}

func (m *BaseFTransportMonitor) OnReopenFailed(prevAttempts uint, prevWait time.Duration) (bool, time.Duration) {
	if prevAttempts >= m.MaxReopenAttempts {
		return false, 0
	}

	nextWait := prevWait * 2
	if nextWait > m.MaxWait {
		nextWait = m.MaxWait
	}
	return true, nextWait
}

func (m *BaseFTransportMonitor) OnClosedCleanly() {}

func (m *BaseFTransportMonitor) OnReopenSucceeded() {}

type monitorRunner struct {
	monitor       FTransportMonitor
	transport     FTransport
	closedChannel <-chan error
}

// Starts a runner to monitor the transport.
func (r *monitorRunner) run() {
	fmt.Println("FTransport Monitor: Beginning to monitor transport...")
	for {
		if cause := <-r.closedChannel; cause != nil {
			if shouldContinue := r.handleUncleanClose(cause); !shouldContinue {
				return
			}
		} else {
			r.handleCleanClose()
			return
		}
	}
}

// Handle a clean close of the transport.
func (r *monitorRunner) handleCleanClose() {
	fmt.Println("FTransport Monitor: FTransport was closed cleanly. Terminating...")
	r.monitor.OnClosedCleanly()
}

// Handle an unclean close of the transport.
func (r *monitorRunner) handleUncleanClose(cause error) bool {
	fmt.Printf("FTransport Monitor: FTransport was closed uncleanly because: %v\n", cause)

	reopen, InitialWait := r.monitor.OnClosedUncleanly(cause)
	if !reopen {
		fmt.Println("FTransport Monitor: Instructed not to reopen. Terminating...")
		return false
	}

	return r.attemptReopen(InitialWait)
}

// Attempt to reopen the uncleanly closed transport.
func (r *monitorRunner) attemptReopen(InitialWait time.Duration) bool {
	wait := InitialWait
	reopen := true
	prevAttempts := uint(0)

	for reopen {
		fmt.Printf("FTransport Monitor: Attempting to reopen after %v\n", wait)
		time.Sleep(wait)

		if err := r.transport.Open(); err != nil {
			fmt.Printf("FTransport Monitor: Failed to re-open transport due to: %v\n", err)
			prevAttempts++

			reopen, wait = r.monitor.OnReopenFailed(prevAttempts, wait)
			continue
		}
		fmt.Printf("FTransport Monitor: Successfully re-opened!")
		r.monitor.OnReopenSucceeded()
		return true
	}

	fmt.Println("FTransport Monitor: ReopenFailed callback instructed not to reopen. Terminating...")
	return false
}
