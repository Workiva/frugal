package middleware

import (
	"reflect"
	"github.com/Workiva/frugal/lib/go"
	"time"
)

func AddTimeoutContextMiddleware(timeout time.Duration) frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			if frugalCtx, ok := args.Context().(*frugal.FContextImpl); ok {
				superCtx := frugal.NewSuperFContext(frugalCtx)
				timeoutCtx, cancelFunc := frugal.WithTimeout(superCtx, timeout)
				args.SetContext(timeoutCtx)


				defer func() {
					cancelFunc()
				}()
			}
			return next(service, method, args)
		}
	}
}

