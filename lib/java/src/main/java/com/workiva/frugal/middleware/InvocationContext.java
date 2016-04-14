package com.workiva.frugal.middleware;

/**
 * InvocationContext contains a target handler, which may or may not be a dynamic proxy, and the service name,
 * representing the execution context of a middleware call.
 *
 * @param <T> the service handler type.
 */
public final class InvocationContext<T> {

    protected String service;
    protected T target;

    /**
     * InvocationContext shouldn't be constructed by user code. User code should just pass through the InvocationContext
     * provided by {@link ServiceMiddleware#apply(InvocationContext)}.
     *
     * @param service the service name.
     * @param target  the service handler.
     */
    protected InvocationContext(String service, T target) {
        this.service = service;
        this.target = target;
    }

}