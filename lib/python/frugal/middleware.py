
"""
InvocationHandler is a function whick takes the instance method being invoked
and its arguments. It processes a service method invocation on a proxy
instance and returns the result. The return value should match the arity of the
proxied method and have the same types. The first argument will always be the
FContext.

ServiceMiddleware is a function which takes an InvocationHandler and returns
a new InvocationHandler. It's used to implement interceptor logic around API
calls. This can be used, for example, to implement retry policies on service
calls, logging, telemetry, or authentication and authorization.
ServiceMiddleware can be applied to both RPC services and pub/sub scopes.

Middleware example:

def logging_middleware(next):

    def invocation_handler(method, args):
        service = '%s.%s' % (method.im_self.__module__,
                             method.im_class.__name__)
        print '==== CALLING %s.%s ====' % (service, method.im_func.func_name)
        ret = next(method, args)
        print '==== CALLED  %s.%s ====' % (service, method.im_func.func_name)
        return ret

    return invocation_handler

"""


class Method(object):
    """Method contains an InvocationHandler and a handle to the method it
    proxies. This should only be used by generated code.
    """

    def __init__(self, method, middleware):
        """Initialize a new Method which proxies the given handler method. This
        should only be called by generated code.

        Args:
            method: instance method
            middleware: list of ServiceMiddleware
        """

        self._handler = _compose_middleware(method, middleware)
        self._proxied_method = method

    def __call__(self, *args):
        """Invoke the Method and return its results. The should only be called
        by generated code.
        """

        return self._handler(self._proxied_method, *args)


def _compose_middleware(method, middleware):
    """Apply the given ServiceMiddleware to the provided method and return
    an InvocationHandler.

    Args:
        method: instance method
        middleware: list of ServiceMiddleware

    Returns:
        InvocationHandler
    """

    def base_handler(method, args):
        return method(*args)

    handler = base_handler
    if middleware:
        for m in middleware:
            handler = m(handler)
    return handler

