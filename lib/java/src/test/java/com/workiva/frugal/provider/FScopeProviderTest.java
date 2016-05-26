package com.workiva.frugal.provider;

import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FScopeTransport;
import com.workiva.frugal.transport.FScopeTransportFactory;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

public class FScopeProviderTest {

    @Test
    public void testProvide() throws Exception {
        FScopeTransportFactory transportFactory = mock(FScopeTransportFactory.class);
        FProtocolFactory protocolFactory = mock(FProtocolFactory.class);

        FScopeProvider provider = new FScopeProvider(transportFactory, protocolFactory);

        FScopeTransport transport = mock(FScopeTransport.class);
        FProtocol fProtocol = mock(FProtocol.class);

        when(transportFactory.getTransport()).thenReturn(transport);
        when(protocolFactory.getProtocol(transport)).thenReturn(fProtocol);

        FScopeProvider.Client client = provider.build();

        assertEquals(transport, client.getTransport());
        assertEquals(fProtocol, client.getProtocol());
    }
}
