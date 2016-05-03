package frugal

import (
	"encoding/base64"
	"errors"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// Ensures NewFProtocolException returns nil if the provided error is nil.
func TestNewFProtocolExceptionNil(t *testing.T) {
	assert.Nil(t, NewFProtocolException(nil))
}

// Ensures NewFProtocolException returns the provided error if it is already an
// FProtocolException.
func TestNewFProtocolExceptionIdempotent(t *testing.T) {
	err := &fProtocolException{}
	assert.Equal(t, err, NewFProtocolException(err))
}

// Ensures NewFProtocolException returns an INVALID_DATA exception if the
// provided error is a base64.CorruptInputError.
func TestNewFProtocolExceptionInvalidData(t *testing.T) {
	var err base64.CorruptInputError = 1
	ex := NewFProtocolException(err)
	assert.Equal(t, thrift.INVALID_DATA, ex.TypeId())
}

// Ensures NewFProtocolException returns an UNKNOWN_PROTOCOL_EXCEPTION
// exception if the error is otherwise not known.
func TestNewFProtocolExceptionUnknownProtocolException(t *testing.T) {
	ex := NewFProtocolException(errors.New("error"))
	assert.Equal(t, thrift.UNKNOWN_PROTOCOL_EXCEPTION, ex.TypeId())
}
