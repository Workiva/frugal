// Autogenerated by Thrift Compiler (0.9.3-wk-2)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package event

import (
	"bytes"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/example/go/gen-go/base"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = bytes.Equal

var _ = base.GoUnusedProtection__

type Foo interface {
	base.BaseFoo
	//This is a thrift service. Frugal will generate bindings that include
	//a frugal Context for each service call.

	// Ping the server.
	Ping() (err error)
	// Blah the server.
	//
	// Parameters:
	//  - Num
	//  - Str
	//  - Event
	Blah(num int32, Str string, event *Event) (r int64, err error)
	// oneway methods don't receive a response from the server.
	//
	// Parameters:
	//  - ID
	//  - Req
	OneWay(id ID, req Request) (err error)
}

//This is a thrift service. Frugal will generate bindings that include
//a frugal Context for each service call.
type FooClient struct {
	*base.BaseFooClient
}

func NewFooClientFactory(t thrift.TTransport, f thrift.TProtocolFactory) *FooClient {
	return &FooClient{BaseFooClient: base.NewBaseFooClientFactory(t, f)}
}

func NewFooClientProtocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *FooClient {
	return &FooClient{BaseFooClient: base.NewBaseFooClientProtocol(t, iprot, oprot)}
}

// Ping the server.
func (p *FooClient) Ping() (err error) {
	if err = p.sendPing(); err != nil {
		return
	}
	return p.recvPing()
}

func (p *FooClient) sendPing() (err error) {
	oprot := p.OutputProtocol
	if oprot == nil {
		oprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.OutputProtocol = oprot
	}
	p.SeqId++
	if err = oprot.WriteMessageBegin("ping", thrift.CALL, p.SeqId); err != nil {
		return
	}
	args := FooPingArgs{}
	if err = args.Write(oprot); err != nil {
		return
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return
	}
	return oprot.Flush()
}

func (p *FooClient) recvPing() (err error) {
	iprot := p.InputProtocol
	if iprot == nil {
		iprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.InputProtocol = iprot
	}
	method, mTypeId, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return
	}
	if method != "ping" {
		err = thrift.NewTApplicationException(thrift.WRONG_METHOD_NAME, "ping failed: wrong method name")
		return
	}
	if p.SeqId != seqId {
		err = thrift.NewTApplicationException(thrift.BAD_SEQUENCE_ID, "ping failed: out of sequence response")
		return
	}
	if mTypeId == thrift.EXCEPTION {
		error7 := thrift.NewTApplicationException(thrift.UNKNOWN_APPLICATION_EXCEPTION, "Unknown Exception")
		var error8 error
		error8, err = error7.Read(iprot)
		if err != nil {
			return
		}
		if err = iprot.ReadMessageEnd(); err != nil {
			return
		}
		err = error8
		return
	}
	if mTypeId != thrift.REPLY {
		err = thrift.NewTApplicationException(thrift.INVALID_MESSAGE_TYPE_EXCEPTION, "ping failed: invalid message type")
		return
	}
	result := FooPingResult{}
	if err = result.Read(iprot); err != nil {
		return
	}
	if err = iprot.ReadMessageEnd(); err != nil {
		return
	}
	return
}

// Blah the server.
//
// Parameters:
//  - Num
//  - Str
//  - Event
func (p *FooClient) Blah(num int32, Str string, event *Event) (r int64, err error) {
	if err = p.sendBlah(num, Str, event); err != nil {
		return
	}
	return p.recvBlah()
}

func (p *FooClient) sendBlah(num int32, Str string, event *Event) (err error) {
	oprot := p.OutputProtocol
	if oprot == nil {
		oprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.OutputProtocol = oprot
	}
	p.SeqId++
	if err = oprot.WriteMessageBegin("blah", thrift.CALL, p.SeqId); err != nil {
		return
	}
	args := FooBlahArgs{
		Num:   num,
		Str:   Str,
		Event: event,
	}
	if err = args.Write(oprot); err != nil {
		return
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return
	}
	return oprot.Flush()
}

func (p *FooClient) recvBlah() (value int64, err error) {
	iprot := p.InputProtocol
	if iprot == nil {
		iprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.InputProtocol = iprot
	}
	method, mTypeId, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return
	}
	if method != "blah" {
		err = thrift.NewTApplicationException(thrift.WRONG_METHOD_NAME, "blah failed: wrong method name")
		return
	}
	if p.SeqId != seqId {
		err = thrift.NewTApplicationException(thrift.BAD_SEQUENCE_ID, "blah failed: out of sequence response")
		return
	}
	if mTypeId == thrift.EXCEPTION {
		error9 := thrift.NewTApplicationException(thrift.UNKNOWN_APPLICATION_EXCEPTION, "Unknown Exception")
		var error10 error
		error10, err = error9.Read(iprot)
		if err != nil {
			return
		}
		if err = iprot.ReadMessageEnd(); err != nil {
			return
		}
		err = error10
		return
	}
	if mTypeId != thrift.REPLY {
		err = thrift.NewTApplicationException(thrift.INVALID_MESSAGE_TYPE_EXCEPTION, "blah failed: invalid message type")
		return
	}
	result := FooBlahResult{}
	if err = result.Read(iprot); err != nil {
		return
	}
	if err = iprot.ReadMessageEnd(); err != nil {
		return
	}
	if result.Awe != nil {
		err = result.Awe
		return
	} else if result.API != nil {
		err = result.API
		return
	}
	value = result.GetSuccess()
	return
}

// oneway methods don't receive a response from the server.
//
// Parameters:
//  - ID
//  - Req
func (p *FooClient) OneWay(id ID, req Request) (err error) {
	if err = p.sendOneWay(id, req); err != nil {
		return
	}
	return
}

func (p *FooClient) sendOneWay(id ID, req Request) (err error) {
	oprot := p.OutputProtocol
	if oprot == nil {
		oprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.OutputProtocol = oprot
	}
	p.SeqId++
	if err = oprot.WriteMessageBegin("oneWay", thrift.ONEWAY, p.SeqId); err != nil {
		return
	}
	args := FooOneWayArgs{
		ID:  id,
		Req: req,
	}
	if err = args.Write(oprot); err != nil {
		return
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return
	}
	return oprot.Flush()
}

type FooProcessor struct {
	*base.BaseFooProcessor
}

func NewFooProcessor(handler Foo) *FooProcessor {
	self11 := &FooProcessor{base.NewBaseFooProcessor(handler)}
	self11.AddToProcessorMap("ping", &fooProcessorPing{handler: handler})
	self11.AddToProcessorMap("blah", &fooProcessorBlah{handler: handler})
	self11.AddToProcessorMap("oneWay", &fooProcessorOneWay{handler: handler})
	return self11
}

type fooProcessorPing struct {
	handler Foo
}

func (p *fooProcessorPing) Process(seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	args := FooPingArgs{}
	if err = args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
		oprot.WriteMessageBegin("ping", thrift.EXCEPTION, seqId)
		x.Write(oprot)
		oprot.WriteMessageEnd()
		oprot.Flush()
		return false, err
	}

	iprot.ReadMessageEnd()
	result := FooPingResult{}
	var err2 error
	if err2 = p.handler.Ping(); err2 != nil {
		x := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing ping: "+err2.Error())
		oprot.WriteMessageBegin("ping", thrift.EXCEPTION, seqId)
		x.Write(oprot)
		oprot.WriteMessageEnd()
		oprot.Flush()
		return true, err2
	}
	if err2 = oprot.WriteMessageBegin("ping", thrift.REPLY, seqId); err2 != nil {
		err = err2
	}
	if err2 = result.Write(oprot); err == nil && err2 != nil {
		err = err2
	}
	if err2 = oprot.WriteMessageEnd(); err == nil && err2 != nil {
		err = err2
	}
	if err2 = oprot.Flush(); err == nil && err2 != nil {
		err = err2
	}
	if err != nil {
		return
	}
	return true, err
}

type fooProcessorBlah struct {
	handler Foo
}

func (p *fooProcessorBlah) Process(seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	args := FooBlahArgs{}
	if err = args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
		oprot.WriteMessageBegin("blah", thrift.EXCEPTION, seqId)
		x.Write(oprot)
		oprot.WriteMessageEnd()
		oprot.Flush()
		return false, err
	}

	iprot.ReadMessageEnd()
	result := FooBlahResult{}
	var retval int64
	var err2 error
	if retval, err2 = p.handler.Blah(args.Num, args.Str, args.Event); err2 != nil {
		switch v := err2.(type) {
		case *AwesomeException:
			result.Awe = v
		case *base.APIException:
			result.API = v
		default:
			x := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing blah: "+err2.Error())
			oprot.WriteMessageBegin("blah", thrift.EXCEPTION, seqId)
			x.Write(oprot)
			oprot.WriteMessageEnd()
			oprot.Flush()
			return true, err2
		}
	} else {
		result.Success = &retval
	}
	if err2 = oprot.WriteMessageBegin("blah", thrift.REPLY, seqId); err2 != nil {
		err = err2
	}
	if err2 = result.Write(oprot); err == nil && err2 != nil {
		err = err2
	}
	if err2 = oprot.WriteMessageEnd(); err == nil && err2 != nil {
		err = err2
	}
	if err2 = oprot.Flush(); err == nil && err2 != nil {
		err = err2
	}
	if err != nil {
		return
	}
	return true, err
}

type fooProcessorOneWay struct {
	handler Foo
}

func (p *fooProcessorOneWay) Process(seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	args := FooOneWayArgs{}
	if err = args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		return false, err
	}

	iprot.ReadMessageEnd()
	var err2 error
	if err2 = p.handler.OneWay(args.ID, args.Req); err2 != nil {
		return true, err2
	}
	return true, nil
}

// HELPER FUNCTIONS AND STRUCTURES

type FooPingArgs struct {
}

func NewFooPingArgs() *FooPingArgs {
	return &FooPingArgs{}
}

func (p *FooPingArgs) Read(iprot thrift.TProtocol) error {
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

func (p *FooPingArgs) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("ping_args"); err != nil {
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

func (p *FooPingArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("FooPingArgs(%+v)", *p)
}

type FooPingResult struct {
}

func NewFooPingResult() *FooPingResult {
	return &FooPingResult{}
}

func (p *FooPingResult) Read(iprot thrift.TProtocol) error {
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

func (p *FooPingResult) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("ping_result"); err != nil {
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

func (p *FooPingResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("FooPingResult(%+v)", *p)
}

// Attributes:
//  - Num
//  - Str
//  - Event
type FooBlahArgs struct {
	Num   int32  `thrift:"num,1" db:"num" json:"num"`
	Str   string `thrift:"Str,2" db:"Str" json:"Str"`
	Event *Event `thrift:"event,3" db:"event" json:"event"`
}

func NewFooBlahArgs() *FooBlahArgs {
	return &FooBlahArgs{}
}

func (p *FooBlahArgs) GetNum() int32 {
	return p.Num
}

func (p *FooBlahArgs) GetStr() string {
	return p.Str
}

var FooBlahArgs_Event_DEFAULT *Event

func (p *FooBlahArgs) GetEvent() *Event {
	if !p.IsSetEvent() {
		return FooBlahArgs_Event_DEFAULT
	}
	return p.Event
}
func (p *FooBlahArgs) IsSetEvent() bool {
	return p.Event != nil
}

func (p *FooBlahArgs) Read(iprot thrift.TProtocol) error {
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
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		case 3:
			if err := p.ReadField3(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
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

func (p *FooBlahArgs) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		p.Num = v
	}
	return nil
}

func (p *FooBlahArgs) ReadField2(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(); err != nil {
		return thrift.PrependError("error reading field 2: ", err)
	} else {
		p.Str = v
	}
	return nil
}

func (p *FooBlahArgs) ReadField3(iprot thrift.TProtocol) error {
	p.Event = &Event{
		ID: -1,
	}
	if err := p.Event.Read(iprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Event), err)
	}
	return nil
}

func (p *FooBlahArgs) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("blah_args"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := p.writeField3(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *FooBlahArgs) writeField1(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("num", thrift.I32, 1); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:num: ", p), err)
	}
	if err := oprot.WriteI32(int32(p.Num)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.num (1) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 1:num: ", p), err)
	}
	return err
}

func (p *FooBlahArgs) writeField2(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("Str", thrift.STRING, 2); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:Str: ", p), err)
	}
	if err := oprot.WriteString(string(p.Str)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.Str (2) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 2:Str: ", p), err)
	}
	return err
}

func (p *FooBlahArgs) writeField3(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("event", thrift.STRUCT, 3); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 3:event: ", p), err)
	}
	if err := p.Event.Write(oprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Event), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 3:event: ", p), err)
	}
	return err
}

func (p *FooBlahArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("FooBlahArgs(%+v)", *p)
}

// Attributes:
//  - Success
//  - Awe
//  - API
type FooBlahResult struct {
	Success *int64             `thrift:"success,0" db:"success" json:"success,omitempty"`
	Awe     *AwesomeException  `thrift:"awe,1" db:"awe" json:"awe,omitempty"`
	API     *base.APIException `thrift:"api,2" db:"api" json:"api,omitempty"`
}

func NewFooBlahResult() *FooBlahResult {
	return &FooBlahResult{}
}

var FooBlahResult_Success_DEFAULT int64

func (p *FooBlahResult) GetSuccess() int64 {
	if !p.IsSetSuccess() {
		return FooBlahResult_Success_DEFAULT
	}
	return *p.Success
}

var FooBlahResult_Awe_DEFAULT *AwesomeException

func (p *FooBlahResult) GetAwe() *AwesomeException {
	if !p.IsSetAwe() {
		return FooBlahResult_Awe_DEFAULT
	}
	return p.Awe
}

var FooBlahResult_API_DEFAULT *base.APIException

func (p *FooBlahResult) GetAPI() *base.APIException {
	if !p.IsSetAPI() {
		return FooBlahResult_API_DEFAULT
	}
	return p.API
}
func (p *FooBlahResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *FooBlahResult) IsSetAwe() bool {
	return p.Awe != nil
}

func (p *FooBlahResult) IsSetAPI() bool {
	return p.API != nil
}

func (p *FooBlahResult) Read(iprot thrift.TProtocol) error {
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
		switch fieldId {
		case 0:
			if err := p.ReadField0(iprot); err != nil {
				return err
			}
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
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

func (p *FooBlahResult) ReadField0(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(); err != nil {
		return thrift.PrependError("error reading field 0: ", err)
	} else {
		p.Success = &v
	}
	return nil
}

func (p *FooBlahResult) ReadField1(iprot thrift.TProtocol) error {
	p.Awe = &AwesomeException{}
	if err := p.Awe.Read(iprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Awe), err)
	}
	return nil
}

func (p *FooBlahResult) ReadField2(iprot thrift.TProtocol) error {
	p.API = &base.APIException{}
	if err := p.API.Read(iprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.API), err)
	}
	return nil
}

func (p *FooBlahResult) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("blah_result"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField0(oprot); err != nil {
		return err
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *FooBlahResult) writeField0(oprot thrift.TProtocol) (err error) {
	if p.IsSetSuccess() {
		if err := oprot.WriteFieldBegin("success", thrift.I64, 0); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err)
		}
		if err := oprot.WriteI64(int64(*p.Success)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T.success (0) field write error: ", p), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err)
		}
	}
	return err
}

func (p *FooBlahResult) writeField1(oprot thrift.TProtocol) (err error) {
	if p.IsSetAwe() {
		if err := oprot.WriteFieldBegin("awe", thrift.STRUCT, 1); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:awe: ", p), err)
		}
		if err := p.Awe.Write(oprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Awe), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 1:awe: ", p), err)
		}
	}
	return err
}

func (p *FooBlahResult) writeField2(oprot thrift.TProtocol) (err error) {
	if p.IsSetAPI() {
		if err := oprot.WriteFieldBegin("api", thrift.STRUCT, 2); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:api: ", p), err)
		}
		if err := p.API.Write(oprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.API), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 2:api: ", p), err)
		}
	}
	return err
}

func (p *FooBlahResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("FooBlahResult(%+v)", *p)
}

// Attributes:
//  - ID
//  - Req
type FooOneWayArgs struct {
	ID  ID      `thrift:"id,1" db:"id" json:"id"`
	Req Request `thrift:"req,2" db:"req" json:"req"`
}

func NewFooOneWayArgs() *FooOneWayArgs {
	return &FooOneWayArgs{}
}

func (p *FooOneWayArgs) GetID() ID {
	return p.ID
}

func (p *FooOneWayArgs) GetReq() Request {
	return p.Req
}
func (p *FooOneWayArgs) Read(iprot thrift.TProtocol) error {
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
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
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

func (p *FooOneWayArgs) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		temp := ID(v)
		p.ID = temp
	}
	return nil
}

func (p *FooOneWayArgs) ReadField2(iprot thrift.TProtocol) error {
	_, _, size, err := iprot.ReadMapBegin()
	if err != nil {
		return thrift.PrependError("error reading map begin: ", err)
	}
	tMap := make(Request, size)
	p.Req = tMap
	for i := 0; i < size; i++ {
		var _key12 Int
		if v, err := iprot.ReadI32(); err != nil {
			return thrift.PrependError("error reading field 0: ", err)
		} else {
			temp := Int(v)
			_key12 = temp
		}
		var _val13 string
		if v, err := iprot.ReadString(); err != nil {
			return thrift.PrependError("error reading field 0: ", err)
		} else {
			_val13 = v
		}
		p.Req[_key12] = _val13
	}
	if err := iprot.ReadMapEnd(); err != nil {
		return thrift.PrependError("error reading map end: ", err)
	}
	return nil
}

func (p *FooOneWayArgs) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("oneWay_args"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *FooOneWayArgs) writeField1(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("id", thrift.I64, 1); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:id: ", p), err)
	}
	if err := oprot.WriteI64(int64(p.ID)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.id (1) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 1:id: ", p), err)
	}
	return err
}

func (p *FooOneWayArgs) writeField2(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("req", thrift.MAP, 2); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:req: ", p), err)
	}
	if err := oprot.WriteMapBegin(thrift.I32, thrift.STRING, len(p.Req)); err != nil {
		return thrift.PrependError("error writing map begin: ", err)
	}
	for k, v := range p.Req {
		if err := oprot.WriteI32(int32(k)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T. (0) field write error: ", p), err)
		}
		if err := oprot.WriteString(string(v)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T. (0) field write error: ", p), err)
		}
	}
	if err := oprot.WriteMapEnd(); err != nil {
		return thrift.PrependError("error writing map end: ", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 2:req: ", p), err)
	}
	return err
}

func (p *FooOneWayArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("FooOneWayArgs(%+v)", *p)
}
