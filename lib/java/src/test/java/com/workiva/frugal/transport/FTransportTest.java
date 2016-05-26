package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FException;
import com.workiva.frugal.protocol.FAsyncCallback;
import com.workiva.frugal.protocol.FClientRegistry;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

public class FTransportTest {

    private final FContext context = new FContext();
    private FAsyncCallback callback = new FAsyncCallback() {
        @Override
        public void onMessage(TTransport transport) throws TException {

        }
    };
    private FTransport transport;

    @Before
    public void setUp() throws Exception {
        transport = new FTransportTester();
    }

    @Test
    public void testRegister() throws Exception {
        FRegistry registry = mock(FRegistry.class);

        transport.setRegistry(registry);

        transport.register(context, callback);

        verify(registry).register(context, callback);
    }

    @Test
    public void testRegisterThrowsFExceptionIfRegistryNotSet() throws Exception {
        try{
            transport.register(context, callback);
        } catch(FException ex) {
            assertEquals("registry not set", ex.getMessage());
        }
    }

    @Test
    public void testUnregister() throws Exception {
        FRegistry registry = mock(FRegistry.class);

        transport.setRegistry(registry);

        transport.unregister(context);

        verify(registry).unregister(context);
    }

    @Test
    public void testUnregisterThrowsFExceptionIfRegistryNotSet() throws Exception {
        try{
            transport.unregister(context);
        } catch(FException ex) {
            assertEquals("registry not set", ex.getMessage());
        }
    }

    class FTransportTester extends FTransport {

        @Override
        public void setRegistry(FRegistry registry){
            this.registry = registry;
        }

        @Override
        public boolean isOpen() { return false; }

        @Override
        public void open() throws TTransportException { }

        @Override
        public void close() { }

        @Override
        public int read(byte[] bytes, int i, int i1) throws TTransportException { return 0; }

        @Override
        public void write(byte[] bytes, int i, int i1) throws TTransportException { }
    }
}
