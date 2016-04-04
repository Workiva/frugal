package frugal

import (
	"time"

	log "github.com/Sirupsen/logrus"
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

// NewDefaultFTransportMonitor creates a new FTransportMonitor with default
// reconnect options (attempts to reconnect 60 times with 2 seconds between
// each attempt).
func NewDefaultFTransportMonitor() FTransportMonitor {
	return &BaseFTransportMonitor{
		MaxReopenAttempts: 60,
		InitialWait:       2 * time.Second,
		MaxWait:           2 * time.Second,
	}
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
	log.Info("frugal: FTransportMonitor beginning to monitor transport...")
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
	log.Info("frugal: FTransportMonitor signaled FTransport was closed cleanly. Terminating...")
	r.monitor.OnClosedCleanly()
}

// Handle an unclean close of the transport.
func (r *monitorRunner) handleUncleanClose(cause error) bool {
	log.Warnf("frugal: FTransportMonitor signaled FTransport was closed uncleanly because: %v\n", cause)

	reopen, InitialWait := r.monitor.OnClosedUncleanly(cause)
	if !reopen {
		log.Warn("frugal: FTransportMonitor instructed not to reopen. Terminating...")
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
		log.Infof("frugal: FTransportMonitor attempting to reopen after %v\n", wait)
		time.Sleep(wait)

		if err := r.transport.Open(); err != nil {
			log.Errorf("frugal: FTransportMonitor failed to re-open transport due to: %v\n", err)
			prevAttempts++

			reopen, wait = r.monitor.OnReopenFailed(prevAttempts, wait)
			continue
		}
		log.Info("frugal: FTransportMonitor successfully re-opened!")
		// Do a sanity check. TODO: Remove this once the "transport not open"
		// bug is fixed.
		if !r.transport.IsOpen() {
			log.Error("frugal: FTransportMonitor sanity check failed - transport is not open!")
		}
		r.monitor.OnReopenSucceeded()
		return true
	}

	log.Warn("frugal: FTransportMonitor ReopenFailed callback instructed not to reopen. Terminating...")
	return false
}
