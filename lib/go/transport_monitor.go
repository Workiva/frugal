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

import "time"

// FTransportMonitor watches and heals an FTransport. It exposes a number of
// hooks which can be used to add logic around FTransport events, such as
// unexpected disconnects, expected disconnects, failed reconnects, and
// successful reconnects.
//
// Most Frugal implementations include a base FTransportMonitor which
// implements basic reconnect logic with backoffs and max attempts. This can be
// extended or reimplemented to provide custom logic.
type FTransportMonitor interface {
	// OnClosedCleanly is called when the transport is closed cleanly by a call
	// to Close()
	OnClosedCleanly()

	// OnClosedUncleanly is called when the transport is closed for a reason
	// *other* than a call to Close(). Returns whether to try reopening the
	// transport and, if so, how long to wait before making the attempt.
	OnClosedUncleanly(cause error) (reopen bool, wait time.Duration)

	// OnReopenFailed is called when an attempt to reopen the transport fails.
	// Given the number of previous attempts to re-open the transport and the
	// length of the previous wait. Returns whether to attempt to re-open the
	// transport, and how long to wait before making the attempt.
	OnReopenFailed(prevAttempts uint, prevWait time.Duration) (reopen bool, wait time.Duration)

	// OnReopenSucceeded is called after the transport has been successfully
	// re-opened.
	OnReopenSucceeded()
}

// BaseFTransportMonitor is a default monitor implementation that attempts to
// re-open a closed transport with exponential backoff behavior and a capped
// number of retries. Its behavior can be customized by embedding this struct
// type in a new struct which "overrides" desired callbacks.
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

// OnClosedUncleanly is called when the transport is closed for a reason
// *other* than a call to Close(). Returns whether to try reopening the
// transport and, if so, how long to wait before making the attempt.
func (m *BaseFTransportMonitor) OnClosedUncleanly(cause error) (bool, time.Duration) {
	return m.MaxReopenAttempts > 0, m.InitialWait
}

// OnReopenFailed is called when an attempt to reopen the transport fails.
// Given the number of previous attempts to re-open the transport and the
// length of the previous wait. Returns whether to attempt to re-open the
// transport, and how long to wait before making the attempt.
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

// OnClosedCleanly is called when the transport is closed cleanly by a call
// to Close()
func (m *BaseFTransportMonitor) OnClosedCleanly() {}

// OnReopenSucceeded is called after the transport has been successfully
// re-opened.
func (m *BaseFTransportMonitor) OnReopenSucceeded() {}

type monitorRunner struct {
	monitor       FTransportMonitor
	transport     FTransport
	closedChannel <-chan error
}

// Starts a runner to monitor the transport.
func (r *monitorRunner) run() {
	logger().Info("frugal: FTransportMonitor beginning to monitor transport...")
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
	logger().Info("frugal: FTransportMonitor signaled FTransport was closed cleanly. Terminating...")
	r.monitor.OnClosedCleanly()
}

// Handle an unclean close of the transport.
func (r *monitorRunner) handleUncleanClose(cause error) bool {
	logger().Warnf("frugal: FTransportMonitor signaled FTransport was closed uncleanly because: %v", cause)

	reopen, InitialWait := r.monitor.OnClosedUncleanly(cause)
	if !reopen {
		logger().Warn("frugal: FTransportMonitor instructed not to reopen. Terminating...")
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
		logger().Infof("frugal: FTransportMonitor attempting to reopen after %v", wait)
		time.Sleep(wait)

		if err := r.transport.Open(); err != nil {
			logger().Errorf("frugal: FTransportMonitor failed to re-open transport due to: %v", err)
			prevAttempts++

			reopen, wait = r.monitor.OnReopenFailed(prevAttempts, wait)
			continue
		}
		logger().Info("frugal: FTransportMonitor successfully re-opened!")
		// Do a sanity check. TODO: Remove this once the "transport not open"
		// bug is fixed.
		if !r.transport.IsOpen() {
			logger().Error("frugal: FTransportMonitor sanity check failed - transport is not open!")
		}
		r.monitor.OnReopenSucceeded()
		return true
	}

	logger().Warn("frugal: FTransportMonitor ReopenFailed callback instructed not to reopen. Terminating...")
	return false
}
