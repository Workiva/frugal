package com.workiva.frugal.middleware;

/**
 * ServiceMiddleware is used to implement interceptor logic around API calls. This
 * can be used, for example, to implement retry policies on service calls,
 * logging, telemetry, or authentication and authorization. ServiceMiddleware can
 * be applied to both RPC services and pub/sub scopes.
 */
public interface ServiceMiddleware {

    /**
     * Returns an InvocationHandler which proxies the given target. This can be used to apply middleware
     * logic around a service call.
     *
     * @param next the next target in the call chain.
     * @param <T>  the handler type.
     * @return proxied InvocationHandler.
     */
    <T> InvocationHandler<T> apply(T next);

}
