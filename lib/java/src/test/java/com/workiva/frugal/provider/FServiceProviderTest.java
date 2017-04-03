package com.workiva.frugal.provider;

import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

/**
 * Tests for {@link FServiceProvider}.
 */
@RunWith(JUnit4.class)
public class FServiceProviderTest {

    @Test
    public void testProvide() throws Exception {
        FProtocolFactory protocolFactory = mock(FProtocolFactory.class);
        FTransport transport = mock(FTransport.class);

        FServiceProvider provider = new FServiceProvider(transport, protocolFactory);

        assertEquals(transport, provider.getTransport());
        assertEquals(protocolFactory, provider.getProtocolFactory());
    }

    @Test
    public void testOpen() throws TTransportException {
        FTransport transport = mock(FTransport.class);
        FServiceProvider provider = new FServiceProvider(transport, null);

        provider.open();

        verify(transport).open();
    }

    @Test(expected = TTransportException.class)
    public void testOpenError() throws TTransportException {
        FTransport transport = mock(FTransport.class);
        FServiceProvider provider = new FServiceProvider(transport, null);
        doThrow(TTransportException.class).when(transport).open();

        provider.open();

        verify(transport).open();
    }

    @Test
    public void testClose() throws TTransportException {
        FTransport transport = mock(FTransport.class);
        FServiceProvider provider = new FServiceProvider(transport, null);

        provider.close();

        verify(transport).close();
    }

}
