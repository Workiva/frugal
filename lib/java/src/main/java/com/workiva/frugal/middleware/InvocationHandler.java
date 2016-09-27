package com.workiva.frugal.middleware;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;

/**
 * InvocationHandler processes a service method invocation on a proxy instance and returns the result.
 *
 * @param <T> the service handler type.
 */
public abstract class InvocationHandler<T> implements java.lang.reflect.InvocationHandler {

    private final T next;

    /**
     * Creates a new InvocationHandler wrapping the given InvocationContext.
     *
     * @param next the next target in the call chain.
     */
    public InvocationHandler(T next) {
        this.next = next;
    }

    @Override
    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
        try {
            return invoke(method, next, args);
        } catch (InvocationTargetException e) {
            throw e.getCause();
        }
    }

    /**
     * Called when the middleware is invoked by a service call.
     *
     * @param method   the method being invoked.
     * @param receiver the method receiver being invoked.
     * @param args     the method arguments. The first argument will always be the FContext.
     * @return the method return value.
     * @throws Throwable thrown by the wrapped method.
     */
    public abstract Object invoke(Method method, T receiver, Object[] args) throws Throwable;

    /**
     * Applies ServiceMiddleware to the provided target object by constructing a dynamic proxy. This should only be
     * called by generated code.
     *
     * @param target     the service handler.
     * @param iface      the interface class the handler implements.
     * @param middleware the middleware to apply.
     * @param <T>        the handler type.
     * @return handler proxy.
     */
    @SuppressWarnings("unchecked")
    public static <T> T composeMiddleware(T target, Class iface, ServiceMiddleware[] middleware) {
        ClassLoader classLoader = target.getClass().getClassLoader();
        Class[] ifaces = new Class[]{iface};
        for (ServiceMiddleware m : middleware) {
            InvocationHandler handler = m.apply(target);
            target = (T) Proxy.newProxyInstance(classLoader, ifaces, handler);
        }
        return target;
    }
}
