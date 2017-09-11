package com.workiva.frugal.transport;

import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

/**
 * Tests for {@link TFramedTransport}.
 */
@RunWith(JUnit4.class)
public class TFramedTransportTest {

    TTransport mockTrans = mock(TTransport.class);

    TFramedTransport transport = (TFramedTransport) new TFramedTransport.Factory().getTransport(mockTrans);

    @Test
    public void testOpen() throws Exception {
        transport.open();

        verify(mockTrans).open();
    }

    @Test
    public void testClose() throws Exception {
        transport.close();

        verify(mockTrans).close();
    }

    @Test
    public void testIsOpen() throws Exception {
        transport.isOpen();

        verify(mockTrans).isOpen();
    }

    @Test(expected = TTransportException.class)
    public void testCantCallRead() throws Exception {
        transport.read(new byte[] {}, 1, 2);
    }

    @Test
    public void testReadFrame() throws Exception {
        transport.readFrame();

        verify(mockTrans).readAll(new byte[4], 0, 4);
    }

    @Test
    public void testWrite() throws Exception {
        transport.write(new byte[] {0x01}, 0, 1);

        byte[] result = transport.writeBuffer.get();

        assertEquals(0x01, result[0]);
    }

    @Test
    public void testFlush() throws Exception {
        transport.write(new byte[] {0x00, 0x00, 0x00, 0x01}, 0, 1);

        byte[] result = transport.writeBuffer.get();
        transport.flush();

        verify(mockTrans).write(new byte[] {0x00, 0x00, 0x00, 0x01}, 0, 4);
        verify(mockTrans).write(result, 0, 1);
        verify(mockTrans).flush();
    }
}
