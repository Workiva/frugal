package frugal

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattrobenolt/gocql/uuid"
)

// ErrTimeout is returned when a request timed out.
var ErrTimeout = errors.New("frugal: request timed out")

const (
	cid            = "_cid"
	opID           = "_opid"
	defaultTimeout = time.Minute
)

// FContext is the context for a Frugal message. Every RPC has an FContext,
// which can be used to set request headers, response headers, and the request
// timeout. The default timeout is one minute. An FContext is also sent with
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
type FContext struct {
	requestHeaders  map[string]string
	responseHeaders map[string]string
	mu              sync.RWMutex
	timeout         time.Duration
}

// NewFContext returns a Context for the given correlation id. If an empty
// correlation id is given, one will be generated. A Context should belong to a
// single request for the lifetime of the request. It can be reused once its
// request has completed, though they should generally not be reused.
func NewFContext(correlationID string) *FContext {
	if correlationID == "" {
		correlationID = generateCorrelationID()
	}
	ctx := &FContext{
		requestHeaders: map[string]string{
			cid:  correlationID,
			opID: "0",
		},
		responseHeaders: make(map[string]string),
		timeout:         defaultTimeout,
	}
	return ctx
}

// CorrelationID returns the correlation id for the context
func (c *FContext) CorrelationID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requestHeaders[cid]
}

// setOpID returns the operation id for the context
func (c *FContext) setOpID(id uint64) {
	opIDStr := strconv.FormatUint(id, 10)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestHeaders[opID] = opIDStr
}

// opID returns the operation id for the context
func (c *FContext) opID() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	opIDStr := c.requestHeaders[opID]
	id, err := strconv.ParseUint(opIDStr, 10, 64)
	if err != nil {
		// Should not happen.
		panic(err)
	}
	return id
}

// AddRequestHeader adds a request header to the context for the given name.
// The headers _cid and _opid are reserved.
func (c *FContext) AddRequestHeader(name, value string) {
	if name == cid || name == opID {
		return
	}
	c.mu.Lock()
	c.requestHeaders[name] = value
	c.mu.Unlock()
}

// RequestHeader gets the named request header
func (c *FContext) RequestHeader(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.requestHeaders[name]
	return val, ok
}

// RequestHeaders returns the request headers map
func (c *FContext) RequestHeaders() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	headers := make(map[string]string, len(c.requestHeaders))
	for name, value := range c.requestHeaders {
		headers[name] = value
	}
	return headers
}

// AddResponseHeader adds a response header to the context for the given name.
// The _opid header is reserved.
func (c *FContext) AddResponseHeader(name, value string) {
	if name == opID {
		return
	}
	c.addResponseHeader(name, value)
}

// ResponseHeader gets the named response header
func (c *FContext) ResponseHeader(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.responseHeaders[name]
	return val, ok
}

// ResponseHeaders returns the response headers map
func (c *FContext) ResponseHeaders() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	headers := make(map[string]string, len(c.responseHeaders))
	for name, value := range c.responseHeaders {
		headers[name] = value
	}
	return headers
}

// SetTimeout sets the request timeout. Default is 1 minute.
func (c *FContext) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	c.timeout = timeout
	c.mu.Unlock()
}

// Timeout returns the request timeout.
func (c *FContext) Timeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.timeout
}

func (c *FContext) setResponseOpID(id string) {
	c.mu.Lock()
	c.responseHeaders[opID] = id
	c.mu.Unlock()
}

// addRequestHeader bypasses the check for reserved headers.
func (c *FContext) addRequestHeader(name, value string) {
	c.mu.Lock()
	c.requestHeaders[name] = value
	c.mu.Unlock()
}

// addResponseHeader bypasses the check for reserved headers.
func (c *FContext) addResponseHeader(name, value string) {
	c.mu.Lock()
	c.responseHeaders[name] = value
	c.mu.Unlock()
}

// generateCorrelationID returns a random string id. It's assigned to a var for
// testability purposes.
var generateCorrelationID = func() string {
	return strings.Replace(uuid.RandomUUID().String(), "-", "", -1)
}
