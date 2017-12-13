package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/Workiva/frugal/lib/go"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// ErrorBody represents the default error response
type ErrorBody struct {
	RequestID string            `json:"request_id"`
	Message   string            `json:"message"`
	Errors    []ValidationError `json:"errors"`
}

// ValidationError tracks the resource, field and type of an invalid JSON input
// matching the API guidelines
type ValidationError struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Code     string `json:"code"`
}

func (e *ValidationError) Error() string {
	buf, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("Failed to marshal error message: %v", e)
	}
	return string(buf)
}

// NewValidationError creates a new Validation error for reporting back to the user
func NewValidationError(resource string, field string, code string) *ValidationError {
	return &ValidationError{resource, field, code}
}

// HTTPStatusFromFrugalError converts a Frugal exception into a corresponding HTTP response status.
// https://github.com/Workiva/frugal/blob/master/lib/go/errors.go
func HTTPStatusFromFrugalError(err thrift.TException) int {
	fmt.Println("http status from frugal error")
	switch err.(type) {
	case thrift.TTransportException:
		if e, ok := err.(thrift.TTransportException); ok {
			switch e.TypeId() {
			case frugal.TRANSPORT_EXCEPTION_UNKNOWN, frugal.TRANSPORT_EXCEPTION_NOT_OPEN, frugal.TRANSPORT_EXCEPTION_ALREADY_OPEN, frugal.TRANSPORT_EXCEPTION_END_OF_FILE:
				return http.StatusInternalServerError
			case frugal.TRANSPORT_EXCEPTION_TIMED_OUT:
				return http.StatusRequestTimeout
			case frugal.TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE:
				return http.StatusBadRequest
			default:
				fmt.Printf("unknown TTransportException: %v\n", reflect.TypeOf(err))
				return http.StatusInternalServerError
			}
		}
	case thrift.TApplicationException:
		if e, ok := err.(thrift.TApplicationException); ok {
			switch e.TypeId() {
			case frugal.APPLICATION_EXCEPTION_INVALID_MESSAGE_TYPE, frugal.APPLICATION_EXCEPTION_WRONG_METHOD_NAME, frugal.APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE:
				return http.StatusBadRequest
			case frugal.APPLICATION_EXCEPTION_UNKNOWN, frugal.APPLICATION_EXCEPTION_UNKNOWN_METHOD, frugal.APPLICATION_EXCEPTION_UNSUPPORTED_CLIENT_TYPE, frugal.APPLICATION_EXCEPTION_BAD_SEQUENCE_ID, frugal.APPLICATION_EXCEPTION_MISSING_RESULT, frugal.APPLICATION_EXCEPTION_INTERNAL_ERROR, frugal.APPLICATION_EXCEPTION_PROTOCOL_ERROR, frugal.APPLICATION_EXCEPTION_INVALID_TRANSFORM, frugal.APPLICATION_EXCEPTION_INVALID_PROTOCOL:
				return http.StatusInternalServerError
			default:
				fmt.Printf("unknown TApplicationException: %v\n", reflect.TypeOf(err))
				return http.StatusInternalServerError
			}
		}
	}

	fmt.Printf("Unknown Frugal error: %v\n", reflect.TypeOf(err))
	return http.StatusInternalServerError
}

// DefaultRequestErrorHandler is the default implementation of an error handler for incoming HTTP requests.
func DefaultRequestErrorHandler(marshaler Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	f, _ := w.(http.Flusher)

	var status int
	var response ErrorBody

	// Handle data type errors, etc.
	switch err.(type) {
	case *ValidationError:
		if e, ok := err.(*ValidationError); ok {
			response = ErrorBody{
				"<request_id>",
				"Invalid JSON data",
				[]ValidationError{
					{
						e.Resource,
						e.Field,
						e.Code,
					},
				},
			}
		}
	default:
		status = http.StatusInternalServerError
	}

	// Serialize the error message to JSON
	buf, merr := marshaler.Marshal(response)
	if merr != nil {
		fmt.Printf("Failed to marshal error message %q: %v", response, merr)
		status = http.StatusInternalServerError
	}

	// Write the response
	w.Header().Set("Content-Type", marshaler.ContentType())
	w.WriteHeader(status)
	w.Write(buf)
	f.Flush()
}

// DefaultFrugalErrorHandler is the default implementation for handling errors in the Frugal processer.
//
// If the error is from Frugal, the function replies with the Frugal
// status code mapped to an HTTP status code using HTTPStatusFromFrugalError.
// If an unknown error occurs, it replies with http.StatusInternalServerError.
//
// The response body returned by this function is a JSON object, which
// matches the Workiva API Guidelines
// TODO: Figure out how to handle IDL exceptions and convert to message
func DefaultFrugalErrorHandler(marshaler Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	f, _ := w.(http.Flusher)
	const fallback = "Internal Server Error"
	var status int
	var response ErrorBody

	// Handle data type errors, etc.
	switch err.(type) {
	case thrift.TException:
		status = HTTPStatusFromFrugalError(err)
		response = ErrorBody{
			"<request_id>",
			err.Error(),
			nil,
		}
	default:
		response = ErrorBody{
			"<request_id>",
			fallback,
			nil,
		}
	}

	// Serialize the error message to JSON
	buf, merr := marshaler.Marshal(response)
	if merr != nil {
		fmt.Printf("Failed to marshal error message %q: %v", response, merr)
		status = http.StatusInternalServerError
		response = ErrorBody{
			"<request_id>",
			fallback,
			nil,
		}
	}

	// Write the response
	w.Header().Set("Content-Type", marshaler.ContentType())
	w.WriteHeader(status)
	w.Write(buf)
	f.Flush()
}

// DefaultResponseErrorHandler is the default implementation of an error handler for outgoing HTTP responses.
func DefaultResponseErrorHandler(marshaler Marshaler, w http.ResponseWriter, r *http.Request, err error) {

}
