package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FClientRegistry;
import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;

import static org.junit.Assert.*;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

public class FMuxTransportTest {

    private FMuxTransport muxTransport;
    private TTransport mockTrans;
    private FRegistry registry;

    @Before
    public void setUp() throws Exception {
        mockTrans = mock(TTransport.class);
        muxTransport = new FMuxTransport.Factory(4).getTransport(mockTrans);
    }

    @Test
    public void testIsOpenTrue() throws Exception {
        when(mockTrans.isOpen()).thenReturn(true);

        registry = new FClientRegistry();

        muxTransport.setRegistry(registry);

        assertTrue(muxTransport.isOpen());
    }

    @Test
    public void testIsOpenFalseWhenTransportClosed() throws Exception {
        when(mockTrans.isOpen()).thenReturn(false);

        registry = new FClientRegistry();

        muxTransport.setRegistry(registry);

        assertFalse(muxTransport.isOpen());
    }

}
