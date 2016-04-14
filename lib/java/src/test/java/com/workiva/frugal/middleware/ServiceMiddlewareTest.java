package com.workiva.frugal.middleware;

import org.junit.Test;

import static org.junit.Assert.assertEquals;

import java.lang.reflect.Method;

public class ServiceMiddlewareTest {

    /**
     * Ensure middleware and the proxied method are properly invoked.
     */
    @Test
    public void testApply() {
        TestMiddleware middleware1 = new TestMiddleware();
        TestMiddleware middleware2 = new TestMiddleware();
        String service = "Test";
        int arg = 42;
        TestHandler handler = new TestHandler();
        Handler proxy = InvocationHandler.composeMiddleware(service, handler, Handler.class,
                new ServiceMiddleware[]{middleware1, middleware2});

        String actual = proxy.handlerMethod(arg);

        assertEquals("foo", actual);
        assertEquals(arg + 2, handler.calledArg);
        assertEquals(arg, middleware2.calledArg);
        assertEquals(service, middleware2.serviceName);
        assertEquals(arg, middleware2.calledArg);
        assertEquals(service, middleware1.serviceName);
        assertEquals(arg + 1, middleware1.calledArg);
    }

    /**
     * Ensure the proxied method is properly invoked if no middleware is provided.
     */
    @Test
    public void testApplyNoMiddleware() {
        String service = "Test";
        int arg = 42;
        TestHandler handler = new TestHandler();
        Handler proxy = InvocationHandler.composeMiddleware(service, handler, Handler.class,
                new ServiceMiddleware[0]);

        String actual = proxy.handlerMethod(arg);

        assertEquals("foo", actual);
        assertEquals(arg, handler.calledArg);
    }

    public interface Handler {
        String handlerMethod(int x);
    }

    public class TestHandler implements Handler {

        private int calledArg;

        public String handlerMethod(int x) {
            calledArg = x;
            return "foo";
        }

    }

    public class TestMiddleware implements ServiceMiddleware {

        private int calledArg;
        private String serviceName;

        @Override
        public <T> InvocationHandler<T> apply(InvocationContext<T> next) {
            return new InvocationHandler<T>(next) {
                @Override
                public Object invoke(String service, Method method, T receiver, Object[] args) throws Throwable {
                    calledArg = (int) args[0];
                    serviceName = service;
                    args[0] = ((int) args[0]) + 1;
                    return method.invoke(receiver, args);
                }
            };
        }

    }

}
