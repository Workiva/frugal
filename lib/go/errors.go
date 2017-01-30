package frugal

import (
	"git.apache.org/thrift.git/lib/go/thrift"
)

// TTransportException types used in frugal instantiated
// TTransportExceptions.
const (
	// Inherited from thrift
	TRANSPORT_EXCEPTION_UNKNOWN      = thrift.UNKNOWN_TRANSPORT_EXCEPTION
	TRANSPORT_EXCEPTION_NOT_OPEN     = thrift.NOT_OPEN
	TRANSPORT_EXCEPTION_ALREADY_OPEN = thrift.ALREADY_OPEN
	TRANSPORT_EXCEPTION_TIMED_OUT    = thrift.TIMED_OUT
	TRANSPORT_EXCEPTION_END_OF_FILE  = thrift.END_OF_FILE

	// TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE is a TTransportException
	// error type indicating the request exceeded the size limit.
	TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE = 100

	// TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE is a TTransportException
	// error type indicating the response exceeded the size limit.
	TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE = 101
)

// TApplicationException types used in frugal instantiated
// TApplicationExceptions.
const (
	// Inherited from thrift
	APPLICATION_EXCEPTION_UNKNOWN                 = thrift.UNKNOWN_APPLICATION_EXCEPTION
	APPLICATION_EXCEPTION_UNKNOWN_METHOD          = thrift.UNKNOWN_METHOD
	APPLICATION_EXCEPTION_INVALID_MESSAGE_TYPE    = thrift.INVALID_MESSAGE_TYPE_EXCEPTION
	APPLICATION_EXCEPTION_WRONG_METHOD_NAME       = thrift.WRONG_METHOD_NAME
	APPLICATION_EXCEPTION_BAD_SEQUENCE_ID         = thrift.BAD_SEQUENCE_ID
	APPLICATION_EXCEPTION_MISSING_RESULT          = thrift.MISSING_RESULT
	APPLICATION_EXCEPTION_INTERNAL_ERROR          = thrift.INTERNAL_ERROR
	APPLICATION_EXCEPTION_PROTOCOL_ERROR          = thrift.PROTOCOL_ERROR
	APPLICATION_EXCEPTION_INVALID_TRANSFORM       = 8
	APPLICATION_EXCEPTION_INVALID_PROTOCOL        = 9
	APPLICATION_EXCEPTION_UNSUPPORTED_CLIENT_TYPE = 10

	// APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE is a TApplicationException
	// error type indicating the response exceeded the size limit.
	APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE = 100
)

// IsErrTooLarge indicates if the given error is a TTransportException
// indicating an oversized request or response.
func IsErrTooLarge(err error) bool {
	if e, ok := err.(thrift.TTransportException); ok {
		return e.TypeId() == TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE ||
			e.TypeId() == TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE
	}
	return false
}
