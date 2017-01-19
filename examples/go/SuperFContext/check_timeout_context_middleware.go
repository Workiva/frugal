package middleware

import (
	"reflect"
	"github.com/Workiva/frugal/lib/go"
	"context"
	"fmt"
)

func CheckDeadlineContextMiddleware() frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			if frugalCtx, ok := args.Context().(*frugal.FContextImpl); ok {
				superCtx := frugal.NewSuperFContext(frugalCtx)
				
				if superCtx.Err() == context.DeadlineExceeded {
					fmt.Println("deadline exceeded...")
				}
			}
			return next(service, method, args)
		}
	}
}

