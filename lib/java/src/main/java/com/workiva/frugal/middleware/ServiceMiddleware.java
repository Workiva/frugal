package com.workiva.frugal.middleware;

/**
 * ServiceMiddleware is used to apply middleware logic around service handlers.
 */
public interface ServiceMiddleware {

    /**
     * Returns an InvocationHandler which proxies the given InvocationContext. This can be used to apply middleware
     * logic around a service call.
     *
     * @param next the next InvocationContext in the chain.
     * @param <T>  the handler type.
     * @return proxied InvocationHandler.
     */
    <T> InvocationHandler<T> apply(InvocationContext<T> next);

}
