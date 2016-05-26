package frugal

import (
	"bytes"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
)

var nextOpID uint64 = 0

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
type FRegistry interface {
	// Register a callback for the given Context.
	Register(ctx *FContext, callback FAsyncCallback) error
	// Unregister a callback for the given Context.
	Unregister(*FContext)
	// Execute dispatches a single Thrift message frame.
	Execute([]byte) error
}

type clientRegistry struct {
	mu       sync.RWMutex
	handlers map[uint64]FAsyncCallback
}

// NewClientRegistry creates a Registry intended for use by Frugal clients.
// This is only to be called by generated code.
func NewFClientRegistry() FRegistry {
	return &clientRegistry{handlers: make(map[uint64]FAsyncCallback)}
}

// Register a callback for the given Context.
func (c *clientRegistry) Register(ctx *FContext, callback FAsyncCallback) error {
	opID := ctx.opID()
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.handlers[opID]
	if ok {
		return errors.New("frugal: context already registered")
	}
	opID = atomic.AddUint64(&nextOpID, 1)
	ctx.setOpID(opID)
	c.handlers[opID] = callback
	return nil
}

// Unregister a callback for the given Context.
func (c *clientRegistry) Unregister(ctx *FContext) {
	opID := ctx.opID()
	c.mu.Lock()
	delete(c.handlers, opID)
	c.mu.Unlock()
}

// Execute dispatches a single Thrift message frame.
func (c *clientRegistry) Execute(frame []byte) error {
	headers, err := getHeadersFromFrame(frame)
	if err != nil {
		log.Warn("frugal: invalid protocol frame headers:", err)
		return err
	}

	opid, err := strconv.ParseUint(headers[opID], 10, 64)
	if err != nil {
		log.Warn("frugal: invalid protocol frame:", err)
		return err
	}

	c.mu.RLock()
	handler, ok := c.handlers[opid]
	if !ok {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	return handler(&thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(frame)})
}

type serverRegistry struct {
	processor            FProcessor
	inputProtocolFactory *FProtocolFactory
	outputProtocol       *FProtocol
}

// NewServerRegistry creates a Registry intended for use by Frugal servers.
// This is only to be called by generated code.
func NewServerRegistry(processor FProcessor, inputProtocolFactory *FProtocolFactory,
	outputProtocol *FProtocol) FRegistry {

	return &serverRegistry{
		processor:            processor,
		inputProtocolFactory: inputProtocolFactory,
		outputProtocol:       outputProtocol,
	}
}

// Register is a no-op for serverRegistry.
func (s *serverRegistry) Register(*FContext, FAsyncCallback) error {
	return nil
}

// Unregister is a no-op for serverRegistry.
func (s *serverRegistry) Unregister(*FContext) {}

// Execute dispatches a single Thrift message frame.
func (s *serverRegistry) Execute(frame []byte) error {
	tr := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(frame)}
	return s.processor.Process(s.inputProtocolFactory.GetProtocol(tr), s.outputProtocol)
}
