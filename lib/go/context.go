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
	"sync/atomic"
	"time"

	"github.com/nats-io/nuid"
)

const (
	// Header containing correlation id
	cidHeader = "_cid"

	// Header containing op id (uint64 as string)
	opIDHeader = "_opid"

	// Header containing request timeout (milliseconds as string)
	timeoutHeader = "_timeout"

	// Default request timeout
	defaultTimeout = 5 * time.Second
)

// FContext is the context for a Frugal message. Every RPC has an FContext,
// which can be used to set request headers, response headers, and the request
// timeout. The default timeout is five seconds. An FContext is also sent with
// every publish message which is then received by subscribers.
//
// In addition to headers, the FContext also contains a correlation ID which
// can be used for distributed tracing purposes. A random correlation ID is
// generated for each FContext if one is not provided.
//
// FContext also plays a key role in Frugal's multiplexing support. A unique,
// per-request operation ID is set on every FContext before a request is made.
// This operation ID is sent in the request and included in the response, which
// is then used to correlate a response to a request. The operation ID is an
// internal implementation detail and is not exposed to the user.
//
// An FContext should belong to a single request for the lifetime of that
// request. It can be reused once the request has completed, though they should
// generally not be reused.
//
// Implementations of FContext must adhere to the following:
//		1)	The CorrelationID should be stored as a request header with the
//			header name "_cid"
//		2)	Threadsafe
type FContext interface {
	// CorrelationID returns the correlation id for the context.
	CorrelationID() string

	// AddRequestHeader adds a request header to the context for the given
	// name. The headers _cid and _opid are reserved. Returns the same FContext
	// to allow for chaining calls.
	AddRequestHeader(name, value string) FContext

	// RequestHeader gets the named request header.
	RequestHeader(name string) (string, bool)

	// RequestHeaders returns the request headers map.
	RequestHeaders() map[string]string

	// AddResponseHeader adds a response header to the context for the given
	// name. The _opid header is reserved. Returns the same FContext to allow
	// for chaining calls.
	AddResponseHeader(name, value string) FContext

	// ResponseHeader gets the named response header.
	ResponseHeader(name string) (string, bool)

	// ResponseHeaders returns the response headers map.
	ResponseHeaders() map[string]string

	// SetTimeout sets the request timeout. Default is 5 seconds. Returns the
	// same FContext to allow for chaining calls.
	SetTimeout(timeout time.Duration) FContext

	// Timeout returns the request timeout.
	Timeout() time.Duration
}

// Clone performs a deep copy of an FContext while handling opids correctly.
// TODO 3.0 consider adding this to the FContext interface.
func Clone(ctx FContext) FContext {
	clone := &FContextImpl{
		requestHeaders:  ctx.RequestHeaders(),
		responseHeaders: ctx.ResponseHeaders(),
	}
	clone.requestHeaders[opIDHeader] = getNextOpID()
	return clone
}

var nextOpID uint64

func getNextOpID() string {
	return strconv.FormatUint(atomic.AddUint64(&nextOpID, 1), 10)
}

// FContextImpl is an implementation of FContext.
type FContextImpl struct {
	requestHeaders  map[string]string
	responseHeaders map[string]string
	mu              sync.RWMutex
}

// NewFContext returns a Context for the given correlation id. If an empty
// correlation id is given, one will be generated. A Context should belong to a
// single request for the lifetime of the request. It can be reused once its
// request has completed, though they should generally not be reused.
func NewFContext(correlationID string) FContext {
	if correlationID == "" {
		correlationID = generateCorrelationID()
	}
	ctx := &FContextImpl{
		requestHeaders: map[string]string{
			cidHeader:     correlationID,
			opIDHeader:    getNextOpID(),
			timeoutHeader: strconv.FormatInt(int64(defaultTimeout/time.Millisecond), 10),
		},
		responseHeaders: make(map[string]string),
	}

	return ctx
}

// CorrelationID returns the correlation id for the context.
func (c *FContextImpl) CorrelationID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requestHeaders[cidHeader]
}

// AddRequestHeader adds a request header to the context for the given name.
// The headers _cid and _opid are reserved. Returns the same FContext to allow
// for chaining calls.
func (c *FContextImpl) AddRequestHeader(name, value string) FContext {
	c.mu.Lock()
	c.requestHeaders[name] = value
	c.mu.Unlock()
	return c
}

// RequestHeader gets the named request header.
func (c *FContextImpl) RequestHeader(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.requestHeaders[name]
	return val, ok
}

// RequestHeaders returns the request headers map.
func (c *FContextImpl) RequestHeaders() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	headers := make(map[string]string, len(c.requestHeaders))
	for name, value := range c.requestHeaders {
		headers[name] = value
	}
	return headers
}

// AddResponseHeader adds a response header to the context for the given name.
// The _opid header is reserved. Returns the same FContext to allow for
// chaining calls.
func (c *FContextImpl) AddResponseHeader(name, value string) FContext {
	c.mu.Lock()
	c.responseHeaders[name] = value
	c.mu.Unlock()
	return c
}

// ResponseHeader gets the named response header.
func (c *FContextImpl) ResponseHeader(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.responseHeaders[name]
	return val, ok
}

// ResponseHeaders returns the response headers map.
func (c *FContextImpl) ResponseHeaders() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	headers := make(map[string]string, len(c.responseHeaders))
	for name, value := range c.responseHeaders {
		headers[name] = value
	}
	return headers
}

// SetTimeout sets the request timeout. Default is 5 seconds. Returns the same
// FContext to allow for chaining calls.
func (c *FContextImpl) SetTimeout(timeout time.Duration) FContext {
	c.mu.Lock()
	c.requestHeaders[timeoutHeader] = strconv.FormatInt(int64(timeout/time.Millisecond), 10)
	c.mu.Unlock()
	return c
}

// Timeout returns the request timeout.
func (c *FContextImpl) Timeout() time.Duration {
	c.mu.RLock()
	timeoutMillisStr := c.requestHeaders[timeoutHeader]
	c.mu.RUnlock()
	timeoutMillis, err := strconv.ParseInt(timeoutMillisStr, 10, 64)
	if err != nil {
		return defaultTimeout
	}
	return time.Millisecond * time.Duration(timeoutMillis)
}

// setRequestOpID sets the request operation id for context.
func setRequestOpID(ctx FContext, id uint64) {
	opIDStr := strconv.FormatUint(id, 10)
	ctx.AddRequestHeader(opIDHeader, opIDStr)
}

// opID returns the request operation id for the given context.
func getOpID(ctx FContext) (uint64, error) {
	opIDStr, ok := ctx.RequestHeader(opIDHeader)
	if !ok {
		// Should not happen unless a client/server sent a bogus context.
		return 0, fmt.Errorf("FContext does not have the required %s request header", opIDHeader)
	}
	id, err := strconv.ParseUint(opIDStr, 10, 64)
	if err != nil {
		// Should not happen unless a client/server sent a bogus context.
		return 0, fmt.Errorf("FContext has an opid that is not a non-negative integer: %s", opIDStr)

	}
	return id, nil
}

// setResponseOpID sets the response operation id for context.
func setResponseOpID(ctx FContext, id string) {
	ctx.AddResponseHeader(opIDHeader, id)
}

// generateCorrelationID returns a random string id. It's assigned to a var for
// testability purposes.
var generateCorrelationID = func() string {
	return nuid.Next()
}
