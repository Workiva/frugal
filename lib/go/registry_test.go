package frugal

import (
	"bytes"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// Ensures Execute returns an error if a bad frugal frame is passed.
func TestClientRegistryBadFrame(t *testing.T) {
	assert := assert.New(t)
	registry := NewFClientRegistry()
	assert.Error(registry.Execute([]byte{0}))
}

// Ensures Execute returns an error if the frame fruagl headers are missing an
// opID.
func TestClientRegistryMissingOpID(t *testing.T) {
	assert := assert.New(t)
	registry := NewFClientRegistry()
	assert.Error(registry.Execute(basicFrame))
}

// Ensures context Register, Execute, and Unregister work as intended with
// a valid frugal frame.
func TestClientRegistry(t *testing.T) {
	assert := assert.New(t)
	called := 0
	cb := func(tr thrift.TTransport) error {
		called++
		return nil
	}
	registry := NewFClientRegistry()
	ctx := NewFContext("")

	// Register the context for the first time
	assert.Nil(registry.Register(ctx, cb))
	opID := ctx.opID()
	assert.True(opID > 0)
	// Encode a frame with this context
	transport := &thrift.TMemoryBuffer{Buffer: new(bytes.Buffer)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}
	err := proto.writeHeader(ctx.RequestHeaders())
	assert.Nil(err)
	// Pass the frame to execute
	frame := transport.Bytes()
	assert.Nil(registry.Execute(frame))
	assert.Equal(1, called)

	// Reregister the same context
	assert.Error(registry.Register(ctx, cb))

	// Unregister the context
	registry.Unregister(ctx)
	_, ok := registry.(*clientRegistry).handlers[ctx.opID()]
	assert.False(ok)
	// But make sure execute sill returns nil when executing a frame with the
	// same opID (it will just drop the frame)
	assert.Nil(registry.Execute(frame))
	assert.Equal(1, called)

	// Now, register the same context again and ensure the opID is increased.
	assert.Nil(registry.Register(ctx, cb))
	assert.True(ctx.opID() > opID)
}

type mockProcessor struct {
	iprot *FProtocol
	oprot *FProtocol
}

func (p *mockProcessor) Process(in, out *FProtocol) error {
	p.iprot = in
	p.oprot = out
	return nil
}

// Ensures registry Execute properly hands off frugal frames to the registry
// Processor.
func TestServerRegistry(t *testing.T) {
	assert := assert.New(t)
	processor := &mockProcessor{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	oprot := protocolFactory.GetProtocol(&mockFTransport{})

	registry := NewServerRegistry(processor, protocolFactory, oprot)
	assert.Nil(registry.Execute(frugalFrame))

	ctx, err := processor.iprot.ReadRequestHeader()
	assert.Nil(err)
	assert.Equal(ctx.CorrelationID(), frugalHeaders[cid])
}
