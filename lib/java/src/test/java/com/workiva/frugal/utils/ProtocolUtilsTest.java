package com.workiva.frugal.utils;

import com.workiva.frugal.util.ProtocolUtils;
import org.junit.Test;

import static org.junit.Assert.assertEquals;

public class ProtocolUtilsTest {

    private byte[] buff = new byte[]{0x65, 0x10, (byte) 0xf3, 0x29, 0x0, 0x0, 0x0, 0x63, 0x0, 0x0, 0x0, 0x0};

    @Test
    public void testReadInt() throws Exception {
        buff = new byte[]{0x65, 0x10, (byte) 0xf3, 0x29, 0x0, 0x0, 0x0, 0x63};
        long anInt = ProtocolUtils.readInt(buff, 4);

        assertEquals(99, anInt);
    }

    @Test
    public void testWriteInt() throws Exception {
        ProtocolUtils.writeInt(69, buff, 8);

        long anInt = ProtocolUtils.readInt(buff, 8);

        assertEquals(69, anInt);
    }

    @Test
    public void testWriteString() throws Exception {
        ProtocolUtils.writeString("char", buff, 8);

    }
}
