package com.workiva.frugal.middleware;

import java.lang.reflect.Method;
import java.lang.reflect.Proxy;

/**
 * InvocationHandler processes a service method invocation on a proxy instance and returns the result.
 *
 * @param <T> the service handler type.
 */
public abstract class InvocationHandler<T> implements java.lang.reflect.InvocationHandler {

    private final InvocationContext<T> context;

    /**
     * Creates a new InvocationHandler wrapping the given InvocationContext.
     *
     * @param next the next InvocationContext in the call chain.
     */
    public InvocationHandler(InvocationContext<T> next) {
        context = next;
    }

    @Override
    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
        return invoke(context.service, method, context.target, args);
    }

    /**
     * Called when the middleware is invoked by a service call.
     *
     * @param service  the name of the service being invoked.
     * @param method   the method being invoked.
     * @param receiver the method receiver being invoked.
     * @param args     the method arguments. The first argument will always be the FContext.
     * @return the method return value.
     * @throws Throwable thrown by the wrapped method.
     */
    public abstract Object invoke(String service, Method method, T receiver, Object[] args) throws Throwable;

    /**
     * Applies ServiceMiddleware to the provided target object by constructing a dynamic proxy. This should only be
     * called by generated code.
     *
     * @param service    the name of the service.
     * @param target     the service handler.
     * @param iface      the interface class the handler implements.
     * @param middleware the middleware to apply.
     * @param <T>        the handler type.
     * @return handler proxy.
     */
    @SuppressWarnings("unchecked")
    public static <T> T composeMiddleware(String service, T target, Class iface, ServiceMiddleware[] middleware) {
        InvocationHandler<T> handler = new InvocationHandler<T>(new InvocationContext<>(service, target)) {
            @Override
            public Object invoke(String service, Method method, T receiver, Object[] args) throws Throwable {
                return method.invoke(receiver, args);
            }
        };

        ClassLoader classLoader = target.getClass().getClassLoader();
        Class[] ifaces = new Class[]{iface};
        T proxy = (T) Proxy.newProxyInstance(classLoader, ifaces, handler);
        for (ServiceMiddleware m : middleware) {
            handler = m.apply(new InvocationContext<>(service, proxy));
            proxy = (T) Proxy.newProxyInstance(classLoader, ifaces, handler);
        }
        return proxy;
    }

}
