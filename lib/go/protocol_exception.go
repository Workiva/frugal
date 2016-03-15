package frugal

import (
	"encoding/base64"
	"git.apache.org/thrift.git/lib/go/thrift"
)

// Frugal Protocol exception
type FProtocolException interface {
	thrift.TProtocolException
}

type fProtocolException struct {
	typeId  int
	message string
}

func (f *fProtocolException) TypeId() int {
	return f.typeId
}

func (f *fProtocolException) String() string {
	return f.message
}

func (f *fProtocolException) Error() string {
	return f.message
}

func NewFProtocolException(err error) FProtocolException {
	if err == nil {
		return nil
	}
	if e, ok := err.(FProtocolException); ok {
		return e
	}
	if _, ok := err.(base64.CorruptInputError); ok {
		return &fProtocolException{thrift.INVALID_DATA, err.Error()}
	}
	return &fProtocolException{thrift.UNKNOWN_PROTOCOL_EXCEPTION, err.Error()}
}

func NewFProtocolExceptionWithType(errType int, message string) FProtocolException {
	return &fProtocolException{errType, message}
}
