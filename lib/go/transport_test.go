package frugal

import (
	"errors"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// Ensures IsErrTooLarge correctly classifies errors.
func TestIsErrTooLarge(t *testing.T) {
	assert.True(t, IsErrTooLarge(thrift.NewTTransportException(TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE, "error")))
	assert.True(t, IsErrTooLarge(thrift.NewTTransportException(TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE, "error")))
	assert.False(t, IsErrTooLarge(nil))
	assert.False(t, IsErrTooLarge(errors.New("error")))
	assert.False(t, IsErrTooLarge(thrift.NewTTransportException(TRANSPORT_EXCEPTION_NOT_OPEN, "error")))
	assert.False(t, IsErrTooLarge(thrift.NewTApplicationException(0, "error")))
}
