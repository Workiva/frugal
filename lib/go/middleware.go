package frugal

import "reflect"

type (
	// Arguments contains the arguments to a service method. The first argument
	// will always be the FContext.
	Arguments []interface{}

	// Results contains the return values from a service method invocation. The
	// last return value will always be an error (or nil).
	Results []interface{}

	// InvocationHandler processes a service method invocation on a proxy
	// instance and returns the result. The args and return value should match
	// the arity of the proxied method and have the same types. The first
	// argument will always be the FContext.
	InvocationHandler func(service, method string, args Arguments) Results

	// ServiceMiddleware returns an InvocationHandler which proxies the given
	// InvocationHandler. This can be used to apply middleware logic around a
	// service call.
	ServiceMiddleware func(InvocationHandler) InvocationHandler
)

// Context returns the first argument value as an FContext.
func (a Arguments) Context() *FContext {
	return a[0].(*FContext)
}

// Error returns the last return value as an error.
func (r Results) Error() error {
	if r[len(r)-1] == nil {
		return nil
	}
	return r[len(r)-1].(error)
}

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
	return func(_, _ string, args Arguments) Results {
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
