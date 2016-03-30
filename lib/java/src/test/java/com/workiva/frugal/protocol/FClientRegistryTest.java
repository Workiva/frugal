package com.workiva.frugal.protocol;

import com.workiva.frugal.exception.FException;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.junit.Before;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.fail;

public class FClientRegistryTest {

    private FClientRegistry registry;

    @Before
    public void setUp() throws Exception {
        registry = new FClientRegistry();

    }

    @Test
    public void testRegister() throws Exception {
        FContext context = new FContext();
        registry.register(context, new FAsyncCallback() {
            @Override
            public void onMessage(TTransport transport) throws TException {

            }
        });

        assertEquals(1, registry.handlers.size());
    }

    @Test
    public void testRegisterThrowsExceptionForMultipleOpIds() throws Exception {
        FContext context = new FContext();
        FAsyncCallback callback = new FAsyncCallback() {
            @Override
            public void onMessage(TTransport transport) throws TException {

            }
        };

        registry.register(context, callback);

        try {
            registry.register(context, callback);
            fail();
        } catch(FException ex) {
            assertEquals("context already registered", ex.getMessage());
        }


    }

    @Test
    public void testUnregister() throws Exception {
        FContext context = new FContext();
        registry.register(context, new FAsyncCallback() {
            @Override
            public void onMessage(TTransport transport) throws TException {

            }
        });

        registry.unregister(context);

        assertEquals(0, registry.handlers.size());
    }

    @Test
    public void testUnregisterWhenContextNotThere() throws Exception {
        registry.unregister(new FContext());
    }
}
