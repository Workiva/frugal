package middleware

import (
	"reflect"
	"github.com/Workiva/frugal/lib/go"
	"time"
	"fmt"
)

func AddTimeoutContextMiddleware(timeout time.Duration) frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			if frugalCtx, ok := args.Context().(*frugal.FContextImpl); ok {
				fmt.Println("adding timeout to fcontext")
				superCtx := frugal.NewSuperFContext(frugalCtx)
				timeoutCtx, cancelFunc := frugal.WithTimeout(superCtx, timeout)
				timeoutCtx.AddRequestHeader("testtesttest", "test")
				args.SetContext(timeoutCtx)


				defer func() {
					cancelFunc()
				}()
			} else {
				fmt.Println("context was not a super context...")
			}
			return next(service, method, args)
		}
	}
}

