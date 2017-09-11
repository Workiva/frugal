/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package frugal

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Ensure middleware and the proxied method are properly invoked.
func TestServiceMiddleware(t *testing.T) {
	assert := assert.New(t)
	var (
		calledContext1 FContext
		calledArg1     int
		serviceName1   string
		methodName1    string
		calledContext2 FContext
		calledArg2     int
		serviceName2   string
		methodName2    string
	)
	ctx1 := NewFContext("ctx1")
	middleware1 := newTestMiddleware(&calledContext1, nil, &calledArg1, &serviceName1, &methodName1)
	middleware2 := newTestMiddleware(&calledContext2, ctx1, &calledArg2, &serviceName2, &methodName2)
	handler := &testHandler{}
	method := NewMethod(handler, handler.handlerMethod, "handlerMethod", []ServiceMiddleware{middleware1, middleware2})
	called := make(chan bool, 1)
	method.AddMiddleware(newTestSimpleMiddleware(called))

	ctx2 := NewFContext("ctx2")
	arg := 42
	ret := method.Invoke([]interface{}{ctx2, arg})

	assert.Equal("foo", ret[0])
	assert.Equal(ctx1, calledContext1)
	assert.Equal(ctx2, calledContext2)
	assert.Equal(arg+2, handler.calledArg)
	assert.Equal(arg, calledArg2)
	assert.Equal("testHandler", serviceName2)
	assert.Equal("handlerMethod", methodName2)
	assert.Equal(arg+1, calledArg1)
	assert.Equal("testHandler", serviceName1)
	assert.Equal("handlerMethod", methodName1)

	select {
	case <-called:
	case <-time.After(5 * time.Second):
		t.Fatal("Expected middleware to be called")
	}
}

// Ensure the proxied method is properly invoked if no middleware is provided.
func TestServiceMiddlewareNoMiddleware(t *testing.T) {
	assert := assert.New(t)
	handler := &testHandler{}
	method := NewMethod(handler, handler.handlerMethod, "handlerMethod", nil)

	ctx := NewFContext("fooid")
	arg := 42
	ret := method.Invoke([]interface{}{ctx, arg})

	assert.Equal("foo", ret[0])
	assert.Equal(arg, handler.calledArg)
}

type testHandler struct {
	calledContext FContext
	calledArg     int
}

func (t *testHandler) handlerMethod(ctx FContext, x int) string {
	t.calledContext = ctx
	t.calledArg = x
	return "foo"
}

func newTestMiddleware(calledContext *FContext, setContext FContext, calledArg *int, serviceName, methodName *string) ServiceMiddleware {
	return func(next InvocationHandler) InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args Arguments) Results {
			*calledContext = args.Context()
			if setContext != nil {
				args.SetContext(setContext)
			}
			*calledArg = args[1].(int)
			*serviceName = service.Type().Elem().Name()
			*methodName = method.Name
			args[1] = args[1].(int) + 1
			return next(service, method, args)
		}
	}
}

func newTestSimpleMiddleware(called chan<- bool) ServiceMiddleware {
	return func(next InvocationHandler) InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args Arguments) Results {
			select {
			case called <- true:
			default:
			}
			return next(service, method, args)
		}
	}
}
