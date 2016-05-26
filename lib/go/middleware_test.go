package frugal

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensure middleware and the proxied method are properly invoked.
func TestServiceMiddleware(t *testing.T) {
	assert := assert.New(t)
	var (
		calledArg1   int
		serviceName1 string
		methodName1  string
		calledArg2   int
		serviceName2 string
		methodName2  string
	)
	middleware1 := newTestMiddleware(&calledArg1, &serviceName1, &methodName1)
	middleware2 := newTestMiddleware(&calledArg2, &serviceName2, &methodName2)
	handler := &testHandler{}
	method := NewMethod(handler, handler.handlerMethod, "handlerMethod", []ServiceMiddleware{middleware1, middleware2})
	arg := 42

	ret := method.Invoke([]interface{}{arg})

	assert.Equal("foo", ret[0])
	assert.Equal(arg+2, handler.calledArg)
	assert.Equal(arg, calledArg2)
	assert.Equal("testHandler", serviceName2)
	assert.Equal("handlerMethod", methodName2)
	assert.Equal(arg+1, calledArg1)
	assert.Equal("testHandler", serviceName1)
	assert.Equal("handlerMethod", methodName1)
}

// Ensure the proxied method is properly invoked if no middleware is provided.
func TestServiceMiddlewareNoMiddleware(t *testing.T) {
	assert := assert.New(t)
	handler := &testHandler{}
	method := NewMethod(handler, handler.handlerMethod, "handlerMethod", nil)
	arg := 42

	ret := method.Invoke([]interface{}{arg})

	assert.Equal("foo", ret[0])
	assert.Equal(arg, handler.calledArg)
}

type testHandler struct {
	calledArg int
}

func (t *testHandler) handlerMethod(x int) string {
	t.calledArg = x
	return "foo"
}

func newTestMiddleware(calledArg *int, serviceName, methodName *string) ServiceMiddleware {
	return func(next InvocationHandler) InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args Arguments) Results {
			*calledArg = args[0].(int)
			*serviceName = service.Type().Elem().Name()
			*methodName = method.Name
			args[0] = args[0].(int) + 1
			return next(service, method, args)
		}
	}
}
