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
	"fmt"
	"strconv"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// FAsyncCallback is an internal callback which is constructed by generated
// code and invoked by an FRegistry when a RPC response is received. In other
// words, it's used to complete RPCs. The operation ID on FContext is used to
// look up the appropriate callback. FAsyncCallback is passed an in-memory
// TTransport which wraps the complete message. The callback returns an error
// or throws an exception if an unrecoverable error occurs and the transport
// needs to be shutdown.
type FAsyncCallback func(thrift.TTransport) error

// FRegistry is responsible for multiplexing and handling received messages.
// Typically there is a client implementation and a server implementation. An
// FRegistry is used by an FTransport.
//
// The client implementation is used on the client side, which is making RPCs.
// When a request is made, an FAsyncCallback is registered to an FContext. When
// a response for the FContext is received, the FAsyncCallback is looked up,
// executed, and unregistered.
//
// The server implementation is used on the server side, which is handling
// RPCs. It does not actually register FAsyncCallbacks but rather has an
// FProcessor registered with it. When a message is received, it's buffered and
// passed to the FProcessor to be handled.
type fRegistry interface {
	// Register a channel for the given Context.
	Register(ctx FContext, resultC chan []byte) error
	// Unregister a callback for the given Context.
	Unregister(FContext)
	// Execute dispatches a single Thrift message frame.
	Execute([]byte) error
}

type fRegistryImpl struct {
	mu       sync.RWMutex
	channels map[uint64]chan []byte
}

// NewFRegistry creates a Registry intended for use by Frugal clients.
// This is only to be called by generated code.
func newFRegistry() fRegistry {
	return &fRegistryImpl{channels: make(map[uint64]chan []byte)}
}

// Register a channel for the given Context.
func (c *fRegistryImpl) Register(ctx FContext, resultC chan []byte) error {
	// An FContext can be reused for multiple requests. Because of this,
	// FContext's have a monotonically increasing atomic uint64. We check
	// the channels map to ensure that request is not still in-flight.
	opID, err := getOpID(ctx)

	c.mu.Lock()
	defer c.mu.Unlock()
	if err == nil {
		_, ok := c.channels[opID]
		if ok {
			return fmt.Errorf("frugal: context already registered, opid %d is in-flight for another request", opID)
		}
	}
	c.channels[opID] = resultC
	return nil
}

// Unregister a callback for the given Context.
func (c *fRegistryImpl) Unregister(ctx FContext) {
	opID, err := getOpID(ctx)
	if err != nil {
		logger().Warnf("Attempted to unregister an FContext with a malformed opid: %s", err)
		return
	}
	c.mu.Lock()
	delete(c.channels, opID)
	c.mu.Unlock()
}

// Execute dispatches a single Thrift message frame.
func (c *fRegistryImpl) Execute(frame []byte) error {
	headers, err := getHeadersFromFrame(frame)
	if err != nil {
		logger().Warn("frugal: invalid protocol frame headers:", err)
		return err
	}

	opid, err := strconv.ParseUint(headers[opIDHeader], 10, 64)
	if err != nil {
		logger().Warn("frugal: invalid protocol frame, op id not a uint64:", err)
		return err
	}

	c.mu.RLock()
	resultC, ok := c.channels[opid]
	if !ok {
		logger().Warn("frugal: unregistered context")
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	resultC <- frame
	return nil
}
