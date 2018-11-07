// Autogenerated by Frugal Compiler (2.23.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package variety

import (
	"errors"

	"github.com/Workiva/frugal/lib/gopherjs/frugal"
)

// FooTransitiveDeps is a service or a client.
type FooTransitiveDeps interface {
	Ping(ctx frugal.Context) (err error)
}

// FooTransitiveDepsClient is the client.
type FooTransitiveDepsClient struct {
	call frugal.CallFunc
}

// NewFooTransitiveDepsClient constructs a FooTransitiveDepsClient.
func NewFooTransitiveDepsClient(cf frugal.CallFunc) *FooTransitiveDepsClient {
	return &FooTransitiveDepsClient{
		call: cf,
	}
}

// Ping calls a server.
func (c *FooTransitiveDepsClient) Ping(ctx frugal.Context) (err error) {
	args := &FooTransitiveDepsPingArgs{}
	res := &FooTransitiveDepsPingResult{}
	err = c.call(ctx, "fooTransitiveDeps", "ping", args, res)
	if err != nil {
		return
	}
	return nil
}

// FooTransitiveDepsProcessor is the client.
type FooTransitiveDepsProcessor struct {
	handler FooTransitiveDeps
}

// NewFooTransitiveDepsProcessor constructs a FooTransitiveDepsProcessor.
func NewFooTransitiveDepsProcessor(handler FooTransitiveDeps) *FooTransitiveDepsProcessor {
	return &FooTransitiveDepsProcessor{
		handler: handler,
	}
}

// Invoke handles internal processing of FooTransitiveDeps invocations.
func (p *FooTransitiveDepsProcessor) Invoke(ctx frugal.Context, method string, in frugal.Protocol) (frugal.Packer, error) {
	switch method {
	case "ping":
		args := &FooTransitiveDepsPingArgs{}
		args.Unpack(in)
		err := in.Err()
		if err != nil {
			return nil, err
		}
		res := &FooTransitiveDepsPingResult{}
		res.Success, err = p.handler.Ping(ctx)
		switch terr := err.(type) {
		}
		return res, err
	default:
		return nil, errors.New("FooTransitiveDeps: unsupported method " + method)
	}
}

// FooTransitiveDepsPingArgs is a frual serializable object.
type FooTransitiveDepsPingArgs struct {
}

// NewFooTransitiveDepsPingArgs constructs a FooTransitiveDepsPingArgs.
func NewFooTransitiveDepsPingArgs() *FooTransitiveDepsPingArgs {
	return &FooTransitiveDepsPingArgs{
		// TODO: default values

	}
}

// Unpack deserializes FooTransitiveDepsPingArgs objects.
func (p *FooTransitiveDepsPingArgs) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("FooTransitiveDepsPingArgs")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes FooTransitiveDepsPingArgs objects.
func (p *FooTransitiveDepsPingArgs) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("FooTransitiveDepsPingArgs")
	prot.PackFieldStop()
	prot.PackStructEnd()
}

// FooTransitiveDepsPingResult is a frual serializable object.
type FooTransitiveDepsPingResult struct {
}

// NewFooTransitiveDepsPingResult constructs a FooTransitiveDepsPingResult.
func NewFooTransitiveDepsPingResult() *FooTransitiveDepsPingResult {
	return &FooTransitiveDepsPingResult{
		// TODO: default values

	}
}

// Unpack deserializes FooTransitiveDepsPingResult objects.
func (p *FooTransitiveDepsPingResult) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("FooTransitiveDepsPingResult")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes FooTransitiveDepsPingResult objects.
func (p *FooTransitiveDepsPingResult) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("FooTransitiveDepsPingResult")
	prot.PackFieldStop()
	prot.PackStructEnd()
}
