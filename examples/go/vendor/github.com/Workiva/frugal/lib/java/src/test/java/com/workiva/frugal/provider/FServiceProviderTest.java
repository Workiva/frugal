package com.workiva.frugal.provider;

import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FTransport;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;

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
}
