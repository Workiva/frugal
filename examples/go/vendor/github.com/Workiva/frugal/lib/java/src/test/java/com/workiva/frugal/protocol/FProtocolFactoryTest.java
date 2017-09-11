package com.workiva.frugal.protocol;

import org.apache.thrift.protocol.TProtocol;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransport;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FProtocolFactory}.
 */
@RunWith(JUnit4.class)
public class FProtocolFactoryTest {

    @Test
    public void testProtocolFactory() throws Exception {
        TProtocolFactory tProtocolFactory = mock(TProtocolFactory.class);
        TTransport transport = mock(TTransport.class);
        TProtocol protocol = mock(TProtocol.class);

        when(tProtocolFactory.getProtocol(transport)).thenReturn(protocol);

        FProtocolFactory fProtocolFactory = new FProtocolFactory(tProtocolFactory);

        FProtocol fProtocol = fProtocolFactory.getProtocol(transport);

        assertEquals(protocol.getTransport(), fProtocol.getTransport());
    }
}
