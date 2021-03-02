package shim

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"
)

// TProtocol interface as defined in Thrift 0.13
type TProtocol0_13 interface {
	WriteMessageBegin(name string, typeId thrift.TMessageType, seqid int32) error
	WriteMessageEnd() error
	WriteStructBegin(name string) error
	WriteStructEnd() error
	WriteFieldBegin(name string, typeId thrift.TType, id int16) error
	WriteFieldEnd() error
	WriteFieldStop() error
	WriteMapBegin(keyType thrift.TType, valueType thrift.TType, size int) error
	WriteMapEnd() error
	WriteListBegin(elemType thrift.TType, size int) error
	WriteListEnd() error
	WriteSetBegin(elemType thrift.TType, size int) error
	WriteSetEnd() error
	WriteBool(value bool) error
	WriteByte(value int8) error
	WriteI16(value int16) error
	WriteI32(value int32) error
	WriteI64(value int64) error
	WriteDouble(value float64) error
	WriteString(value string) error
	WriteBinary(value []byte) error

	ReadMessageBegin() (name string, typeId thrift.TMessageType, seqid int32, err error)
	ReadMessageEnd() error
	ReadStructBegin() (name string, err error)
	ReadStructEnd() error
	ReadFieldBegin() (name string, typeId thrift.TType, id int16, err error)
	ReadFieldEnd() error
	ReadMapBegin() (keyType thrift.TType, valueType thrift.TType, size int, err error)
	ReadMapEnd() error
	ReadListBegin() (elemType thrift.TType, size int, err error)
	ReadListEnd() error
	ReadSetBegin() (elemType thrift.TType, size int, err error)
	ReadSetEnd() error
	ReadBool() (value bool, err error)
	ReadByte() (value int8, err error)
	ReadI16() (value int16, err error)
	ReadI32() (value int32, err error)
	ReadI64() (value int64, err error)
	ReadDouble() (value float64, err error)
	ReadString() (value string, err error)
	ReadBinary() (value []byte, err error)

	Skip(fieldType thrift.TType) (err error)
	Flush(ctx context.Context) (err error)

	Transport() thrift.TTransport
}

func NewTProtocol(tp thrift.TProtocol) TProtocol0_13 {
	return tProtocol{TProtocol: tp}
}

// TProtocol matches TProtocol0_13 in this hypothetical minor release of Frugal.
// In a future release built against Thrift 0.14, the stable methods from TProtocol0_13
// would need to be added to this struct type.
type tProtocol struct {
	thrift.TProtocol
}
