package frugal

import "reflect"

type (
	// InvocationHandler processes a service method invocation on a proxy
	// instance and returns the result.
	InvocationHandler func(service, method string, args []interface{}) []interface{}

	// ServiceMiddleware returns an InvocationHandler which proxies the given
	// InvocationHandler. This can be used to apply middleware logic around a
	// service call.
	ServiceMiddleware func(InvocationHandler) InvocationHandler
)

// ComposeMiddleware applies ServiceMiddleware to the provided function. This
// panics if the first argument is not a function. This should only be called
// by generated code.
func ComposeMiddleware(method interface{}, middleware []ServiceMiddleware) InvocationHandler {
	handler := newInvocationHandler(method)
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func newInvocationHandler(method interface{}) InvocationHandler {
	return func(_, _ string, args []interface{}) []interface{} {
		argValues := make([]reflect.Value, len(args))
		for i, arg := range args {
			argValues[i] = reflect.ValueOf(arg)
		}
		returnValues := reflect.ValueOf(method).Call(argValues)
		results := make([]interface{}, len(returnValues))
		for i, ret := range returnValues {
			results[i] = ret.Interface()
		}
		return results
	}
}
