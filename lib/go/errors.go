package frugal

import (
	"errors"
	"git.apache.org/thrift.git/lib/go/thrift"
)

const (
	// REQUEST_TOO_LARGE is a TTransportException error type indicating the
	// request exceeded the size limit.
	REQUEST_TOO_LARGE = 100

	// RESPONSE_TOO_LARGE is a TTransportException error type indicating the
	// response exceeded the size limit.
	RESPONSE_TOO_LARGE = 101

	// RATE_LIMIT_EXCEEDED is a TApplicationException error type indicating the
	// client exceeded its rate limit.
	RATE_LIMIT_EXCEEDED = 100
)

var (
	// ErrTransportClosed is returned by service calls when the transport is
	// unexpectedly closed, perhaps as a result of the transport entering an
	// invalid state. If this is returned, the transport should be
	// reinitialized.
	ErrTransportClosed = errors.New("frugal: transport was unexpectedly closed")

	// ErrTooLarge is returned when attempting to write a message which exceeds
	// the transport's message size limit.
	ErrTooLarge = thrift.NewTTransportException(REQUEST_TOO_LARGE,
		"request was too large for the transport")

	// ErrRateLimitExceeded is returned when the client exceeds its rate limit.
	ErrRateLimitExceeded = thrift.NewTApplicationException(RATE_LIMIT_EXCEEDED,
		"client exceeded rate limit")
)

// IsErrTooLarge indicates if the given error is an ErrTooLarge.
func IsErrTooLarge(err error) bool {
	if err == ErrTooLarge {
		return true
	}
	if e, ok := err.(thrift.TTransportException); ok {
		return e.TypeId() == REQUEST_TOO_LARGE || e.TypeId() == RESPONSE_TOO_LARGE
	}
	return false
}
