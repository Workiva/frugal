package frugal

import (
	"errors"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// Ensures IsErrTooLarge correctly classifies errors.
func TestIsErrTooLarge(t *testing.T) {
	assert.True(t, IsErrTooLarge(ErrTooLarge))
	assert.True(t, IsErrTooLarge(thrift.PrependError("error", ErrTooLarge)))
	assert.True(t, IsErrTooLarge(thrift.NewTTransportException(REQUEST_TOO_LARGE, "error")))
	assert.True(t, IsErrTooLarge(thrift.NewTTransportException(RESPONSE_TOO_LARGE, "error")))
	assert.False(t, IsErrTooLarge(nil))
	assert.False(t, IsErrTooLarge(errors.New("error")))
	assert.False(t, IsErrTooLarge(thrift.NewTTransportException(thrift.NOT_OPEN, "error")))
	assert.False(t, IsErrTooLarge(thrift.NewTApplicationException(0, "error")))
}

// Ensures SetRegistry panics when the registry is nil.
func TestFBaseTransportSetRegistryNilPanic(t *testing.T) {
	tr := newFBaseTransport(0)
	defer func() {
		assert.NotNil(t, recover())
	}()
	tr.SetRegistry(nil)
}

// Ensures SetRegistry does nothing if the registry is already set.
func TestFBaseTransportSetRegistryAlreadySet(t *testing.T) {
	registry := NewFClientRegistry()
	tr := newFBaseTransport(0)
	tr.SetRegistry(registry)
	assert.Equal(t, registry, tr.registry)
	tr.SetRegistry(NewServerRegistry(nil, nil, nil))
	assert.Equal(t, registry, tr.registry)
}
