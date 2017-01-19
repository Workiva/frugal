package frugal

import (
	"reflect"
)

func NewContextMiddleware() ServiceMiddleware {
	return func(next InvocationHandler) InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args Arguments) Results {
			if frugalContext, ok := args.Context().(*FContextImpl); ok {
				ctx := NewSuperFContext(frugalContext)
				args.SetContext(ctx)

				defer func() {
					frugalContext.Inject(ctx.Extract())
					args.SetContext(frugalContext)
				}()
			}

			return next(service, method, args)
		}
	}
}
