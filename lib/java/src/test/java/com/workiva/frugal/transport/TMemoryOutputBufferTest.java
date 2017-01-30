package com.workiva.frugal.transport;

import com.workiva.frugal.util.ProtocolUtils;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;

/**
 * Tests for {@link }.
 */
@RunWith(JUnit4.class)
public class TMemoryOutputBufferTest {

    private TMemoryOutputBuffer buffer;
    private byte[] emptyFrameSize = new byte[4];

    @Before
    public void setUp() {
        buffer = new TMemoryOutputBuffer(10);
    }

    @Test
    public void testWrite() throws TTransportException {
        byte[] data = "foo".getBytes();
        byte[] framedData = new byte[4 + data.length];
        ProtocolUtils.writeInt(data.length, framedData, 0);
        System.arraycopy(data, 0, framedData, 4, data.length);
        buffer.write(data);
        assertArrayEquals(framedData, buffer.getWriteBytes());
        buffer.reset();
        assertArrayEquals(emptyFrameSize, buffer.getWriteBytes());
    }

    @Test
    public void testWriteLen() throws TTransportException {
        byte[] data = "foo".getBytes();
        byte[] framedData = new byte[4 + data.length];
        ProtocolUtils.writeInt(data.length, framedData, 0);
        System.arraycopy(data, 0, framedData, 4, data.length);
        buffer.write("foooooooo".getBytes(), 0, 3);
        assertArrayEquals(framedData, buffer.getWriteBytes());
    }

    @Test(expected = TTransportException.class)
    public void testWrite_sizeException() throws TTransportException {
        assertEquals(4, buffer.size());
        buffer.write(new byte[7]);
        assertEquals(0, buffer.size());
    }

    @Test(expected = TTransportException.class)
    public void testWriteLen_size_Exception() throws TTransportException {
        assertEquals(4, buffer.size());
        buffer.write(new byte[7], 0, 6);
        assertEquals(10, buffer.size());
        buffer.write(new byte[7], 6, 1);
        assertEquals(4, buffer.size());
    }

}
