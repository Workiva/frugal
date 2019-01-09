// Autogenerated by Frugal Compiler (2.25.3)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package golang

import (
	"fmt"

	"github.com/Workiva/frugal/lib/gopherjs/frugal"
	"github.com/Workiva/frugal/lib/gopherjs/thrift"
)

type FBaseFoo interface {
	BasePing(ctx frugal.FContext) (err error)
}

type FBaseFooClient struct {
	transport       frugal.FTransport
	protocolFactory *frugal.FProtocolFactory
	methods         map[string]*frugal.Method
}

func NewFBaseFooClient(provider *frugal.FServiceProvider, middleware ...frugal.ServiceMiddleware) *FBaseFooClient {
	methods := make(map[string]*frugal.Method)
	client := &FBaseFooClient{
		transport:       provider.GetTransport(),
		protocolFactory: provider.GetProtocolFactory(),
		methods:         methods,
	}
	middleware = append(middleware, provider.GetMiddleware()...)
	methods["basePing"] = frugal.NewMethod(client, client.basePing, "basePing", middleware)
	return client
}

func (f *FBaseFooClient) BasePing(ctx frugal.FContext) (err error) {
	ret := f.methods["basePing"].Invoke([]interface{}{ctx})
	if len(ret) != 1 {
		panic(fmt.Sprintf("Middleware returned %d arguments, expected 1", len(ret)))
	}
	if ret[0] != nil {
		err = ret[0].(error)
	}
	return err
}

func (f *FBaseFooClient) basePing(ctx frugal.FContext) (err error) {
	buffer := frugal.NewTMemoryOutputBuffer(f.transport.GetRequestSizeLimit())
	oprot := f.protocolFactory.GetProtocol(buffer)
	if err = oprot.WriteRequestHeader(ctx); err != nil {
		return
	}
	if err = oprot.WriteMessageBegin("basePing", thrift.CALL, 0); err != nil {
		return
	}
	args := BaseFooBasePingArgs{}
	if err = args.Write(oprot); err != nil {
		return
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return
	}
	if err = oprot.Flush(); err != nil {
		return
	}
	var resultTransport thrift.TTransport
	resultTransport, err = f.transport.Request(ctx, buffer.Bytes())
	if err != nil {
		return
	}
	iprot := f.protocolFactory.GetProtocol(resultTransport)
	if err = iprot.ReadResponseHeader(ctx); err != nil {
		return
	}
	method, mTypeId, _, err := iprot.ReadMessageBegin()
	if err != nil {
		return
	}
	if method != "basePing" {
		err = thrift.NewTApplicationException(frugal.APPLICATION_EXCEPTION_WRONG_METHOD_NAME, "basePing failed: wrong method name")
		return
	}
	if mTypeId == thrift.EXCEPTION {
		error0 := thrift.NewTApplicationException(frugal.APPLICATION_EXCEPTION_UNKNOWN, "Unknown Exception")
		var error1 thrift.TApplicationException
		error1, err = error0.Read(iprot)
		if err != nil {
			return
		}
		if err = iprot.ReadMessageEnd(); err != nil {
			return
		}
		if error1.TypeId() == frugal.APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE {
			err = thrift.NewTTransportException(frugal.TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE, error1.Error())
			return
		}
		err = error1
		return
	}
	if mTypeId != thrift.REPLY {
		err = thrift.NewTApplicationException(frugal.APPLICATION_EXCEPTION_INVALID_MESSAGE_TYPE, "basePing failed: invalid message type")
		return
	}
	result := BaseFooBasePingResult{}
	if err = result.Read(iprot); err != nil {
		return
	}
	if err = iprot.ReadMessageEnd(); err != nil {
		return
	}
	return
}

type FBaseFooProcessor struct {
	*frugal.FBaseProcessor
}

func NewFBaseFooProcessor(handler FBaseFoo, middleware ...frugal.ServiceMiddleware) *FBaseFooProcessor {
	p := &FBaseFooProcessor{frugal.NewFBaseProcessor()}
	p.AddToProcessorMap("basePing", &basefooFBasePing{frugal.NewFBaseProcessorFunction(p.GetWriteMutex(), frugal.NewMethod(handler, handler.BasePing, "BasePing", middleware))})
	return p
}

type basefooFBasePing struct {
	*frugal.FBaseProcessorFunction
}

func (p *basefooFBasePing) Process(ctx frugal.FContext, iprot, oprot *frugal.FProtocol) error {
	args := BaseFooBasePingArgs{}
	var err error
	if err = args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		p.GetWriteMutex().Lock()
		err = basefooWriteApplicationError(ctx, oprot, frugal.APPLICATION_EXCEPTION_PROTOCOL_ERROR, "basePing", err.Error())
		p.GetWriteMutex().Unlock()
		return err
	}

	iprot.ReadMessageEnd()
	result := BaseFooBasePingResult{}
	var err2 error
	ret := p.InvokeMethod([]interface{}{ctx})
	if len(ret) != 1 {
		panic(fmt.Sprintf("Middleware returned %d arguments, expected 1", len(ret)))
	}
	if ret[0] != nil {
		err2 = ret[0].(error)
	}
	if err2 != nil {
		if err3, ok := err2.(thrift.TApplicationException); ok {
			p.GetWriteMutex().Lock()
			oprot.WriteResponseHeader(ctx)
			oprot.WriteMessageBegin("basePing", thrift.EXCEPTION, 0)
			err3.Write(oprot)
			oprot.WriteMessageEnd()
			oprot.Flush()
			p.GetWriteMutex().Unlock()
			return nil
		}
		p.GetWriteMutex().Lock()
		err2 := basefooWriteApplicationError(ctx, oprot, frugal.APPLICATION_EXCEPTION_INTERNAL_ERROR, "basePing", "Internal error processing basePing: "+err2.Error())
		p.GetWriteMutex().Unlock()
		return err2
	}
	p.GetWriteMutex().Lock()
	defer p.GetWriteMutex().Unlock()
	if err2 = oprot.WriteResponseHeader(ctx); err2 != nil {
		if frugal.IsErrTooLarge(err2) {
			basefooWriteApplicationError(ctx, oprot, frugal.APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE, "basePing", err2.Error())
			return nil
		}
		err = err2
	}
	if err2 = oprot.WriteMessageBegin("basePing", thrift.REPLY, 0); err2 != nil {
		if frugal.IsErrTooLarge(err2) {
			basefooWriteApplicationError(ctx, oprot, frugal.APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE, "basePing", err2.Error())
			return nil
		}
		err = err2
	}
	if err2 = result.Write(oprot); err == nil && err2 != nil {
		if frugal.IsErrTooLarge(err2) {
			basefooWriteApplicationError(ctx, oprot, frugal.APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE, "basePing", err2.Error())
			return nil
		}
		err = err2
	}
	if err2 = oprot.WriteMessageEnd(); err == nil && err2 != nil {
		if frugal.IsErrTooLarge(err2) {
			basefooWriteApplicationError(ctx, oprot, frugal.APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE, "basePing", err2.Error())
			return nil
		}
		err = err2
	}
	if err2 = oprot.Flush(); err == nil && err2 != nil {
		if frugal.IsErrTooLarge(err2) {
			basefooWriteApplicationError(ctx, oprot, frugal.APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE, "basePing", err2.Error())
			return nil
		}
		err = err2
	}
	return err
}

func basefooWriteApplicationError(ctx frugal.FContext, oprot *frugal.FProtocol, type_ int32, method, message string) error {
	x := thrift.NewTApplicationException(type_, message)
	oprot.WriteResponseHeader(ctx)
	oprot.WriteMessageBegin(method, thrift.EXCEPTION, 0)
	x.Write(oprot)
	oprot.WriteMessageEnd()
	oprot.Flush()
	return x
}

type BaseFooBasePingArgs struct {
}

func NewBaseFooBasePingArgs() *BaseFooBasePingArgs {
	return &BaseFooBasePingArgs{}
}

func (p *BaseFooBasePingArgs) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		if err := iprot.Skip(fieldTypeId); err != nil {
			return err
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *BaseFooBasePingArgs) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("basePing_args"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *BaseFooBasePingArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BaseFooBasePingArgs(%+v)", *p)
}

type BaseFooBasePingResult struct {
}

func NewBaseFooBasePingResult() *BaseFooBasePingResult {
	return &BaseFooBasePingResult{}
}

func (p *BaseFooBasePingResult) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		if err := iprot.Skip(fieldTypeId); err != nil {
			return err
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *BaseFooBasePingResult) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("basePing_result"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *BaseFooBasePingResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BaseFooBasePingResult(%+v)", *p)
}
