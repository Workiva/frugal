package shim

import (
	"github.com/apache/thrift/lib/go/thrift"
)

// TStruct interface as defined by Thrift 0.13
type TStruct0_13 interface {
	Write(p TProtocol0_13) error
	Read(p TProtocol0_13) error
}

func NewTStruct(ts thrift.TStruct) TStruct0_13 {
	return tStruct{TStruct: ts}
}

type tStruct struct {
	thrift.TStruct
}

func (t tStruct) Write(p TProtocol0_13) error {
	return t.TStruct.Write(p)
}
func (t tStruct) Read(p TProtocol0_13) error {
	return t.TStruct.Read(p)
}
