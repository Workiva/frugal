package frugal

import (
	"fmt"
	"reflect"
	"unicode"
)

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
	InvocationHandler func(service reflect.Value, method reflect.Method, args Arguments) Results

	// ServiceMiddleware is used to implement interceptor logic around API
	// calls. This can be used, for example, to implement retry policies on
	// service calls, logging, telemetry, or authentication and authorization.
	// ServiceMiddleware can be applied to both RPC services and pub/sub
	// scopes.
	//
	// ServiceMiddleware returns an InvocationHandler which proxies the given
	// InvocationHandler. This can be used to apply middleware logic around a
	// service call.
	ServiceMiddleware func(InvocationHandler) InvocationHandler

	// Method contains an InvocationHandler and a handle to the method it
	// proxies. This should only be used by generated code.
	Method struct {
		handler       InvocationHandler
		proxiedStruct reflect.Value
		proxiedMethod reflect.Method
	}
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

// SetError sets the last return value as the given error. This will result in
// a panic if Results has not been properly allocated. Also note that returned
// errors should match your IDL definition.
func (r Results) SetError(err error) {
	r[len(r)-1] = err
}

// Invoke the Method and return its results. This should only be called by
// generated code.
func (m *Method) Invoke(args Arguments) Results {
	return m.handler(m.proxiedStruct, m.proxiedMethod, args)
}

// NewMethod creates a new Method which proxies the given handler.
// ProxiedHandler must be a struct and method must be a function. This should
// only be called by generated code.
func NewMethod(proxiedHandler, method interface{}, methodName string, middleware []ServiceMiddleware) *Method {
	var (
		reflectHandler     = reflect.ValueOf(proxiedHandler)
		reflectMethodValue = reflect.ValueOf(method)
		reflectMethod      reflect.Method
	)
	// As of go1.7, reflect.MethodByName no longer returns unexported methods
	// (https://golang.org/doc/go1.7). To avoid exporting generated internal
	// methods, construct a reflect.Method ourselves.
	if unicode.IsLower(rune(methodName[0])) {
		reflectMethod = reflect.Method{
			Name: methodName,
			Type: reflect.TypeOf(method),
			Func: reflectMethodValue,
		}
	} else {
		method, ok := reflectHandler.Type().MethodByName(methodName)
		if !ok {
			panic(fmt.Sprintf("frugal: no such method %s on type %s", methodName, reflectHandler))
		}
		reflectMethod = method
	}
	return &Method{
		handler:       composeMiddleware(reflectMethodValue, middleware),
		proxiedStruct: reflectHandler,
		proxiedMethod: reflectMethod,
	}
}

// composeMiddleware applies ServiceMiddleware to the provided function. This
// panics if the first argument is not a function.
func composeMiddleware(method reflect.Value, middleware []ServiceMiddleware) InvocationHandler {
	handler := newInvocationHandler(method)
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

// newInvocationHandler returns the base InvocationHandler which calls the
// actual handler function.
func newInvocationHandler(method reflect.Value) InvocationHandler {
	return func(_ reflect.Value, _ reflect.Method, args Arguments) Results {
		argValues := make([]reflect.Value, len(args))
		for i, arg := range args {
			argValues[i] = reflect.ValueOf(arg)
		}
		returnValues := method.Call(argValues)
		results := make([]interface{}, len(returnValues))
		for i, ret := range returnValues {
			results[i] = ret.Interface()
		}
		return results
	}
}
