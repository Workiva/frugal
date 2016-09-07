package com.workiva.frugal.internal;

import com.workiva.frugal.exception.FProtocolException;
import com.workiva.frugal.protocol.HeaderUtils;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import java.util.HashMap;
import java.util.Map;

import static org.junit.Assert.assertEquals;

/**
 * Tests for {@link HeaderUtils}.
 */
@RunWith(JUnit4.class)
public class HeaderUtilsTest {

    private static final Map<String, String> HEADERS;

    static {
        HEADERS = new HashMap<>();
        HEADERS.put("foo", "bar");
        HEADERS.put("blah", "baz");
    }

    private static final byte[] LIST = new byte[]{
          0, 0, 0, 0, 29, 0, 0, 0, 3, 102, 111, 111, 0, 0, 0, 3, 98, 97,
          114, 0, 0, 0, 4, 98, 108, 97, 104, 0, 0, 0, 3, 98, 97, 122
    };

    @Rule
    public final ExpectedException thrown = ExpectedException.none();

    @Test
    public void testReadOutOfTransport() throws TException {
        TTransport transport = new TMemoryInputTransport(LIST);

        Map<String, String> decodedHeaders = HeaderUtils.read(transport);
        assertEquals(HEADERS, decodedHeaders);
    }

    @Test
    public void testReadThrowsFExceptionForUnsupportedVersion() throws TException {
        TTransport transport = new TMemoryInputTransport(new byte[]{1});

        thrown.expect(TException.class);
        thrown.expectMessage("unsupported header version 1");
        HeaderUtils.read(transport);
    }

    @Test
    public void testReadThrowsTExceptionForTTransportException() throws TException {
        TTransport transport = new TMemoryInputTransport(new byte[]{0, 0, 0});

        thrown.expect(TException.class);
        // CHECKSTYLE:OFF
        thrown.expectMessage("Cannot read. Remote side has closed. Tried to read 4 bytes, but only got 2 bytes. (This is often indicative of an internal error on the server side. Please check your server logs.");
        // CHECKSTYLE:ON
        HeaderUtils.read(transport);
    }

    @Test
    public void testDecodeFromFrame() throws TException {
        Map<String, String> decodedHeaders = HeaderUtils.decodeFromFrame(LIST);
        assertEquals(HEADERS, decodedHeaders);
    }

    @Test
    public void testEncodeDecode() throws TException {
        byte[] encodedHeaders = HeaderUtils.encode(HEADERS);
        Map<String, String> decodedHeaders = HeaderUtils.decodeFromFrame(encodedHeaders);
        assertEquals(HEADERS, decodedHeaders);
    }

    @Test
    public void testEncodeDecodeNull() throws TException {
        Map<String, String> empty = new HashMap<>();
        byte[] encodedHeaders = HeaderUtils.encode(null);
        Map<String, String> decodedHeaders = HeaderUtils.decodeFromFrame(encodedHeaders);
        assertEquals(empty, decodedHeaders);
    }

    @Test
    public void testEncodeDecodeEmpty() throws TException {
        Map<String, String> empty = new HashMap<>();
        byte[] encodedHeaders = HeaderUtils.encode(empty);
        Map<String, String> decodedHeaders = HeaderUtils.decodeFromFrame(encodedHeaders);
        assertEquals(empty, decodedHeaders);
    }

    @Test
    public void testEncodeDecodeUTF8() throws TException {
        Map<String, String> headers = new HashMap<>();
        headers.put("Đ¥ÑØ", "δάüΓ");
        headers.put("good\u00F1ight", "moo\u00F1");
        byte[] encodedHeaders = HeaderUtils.encode(headers);
        Map<String, String> decodedHeaders = HeaderUtils.decodeFromFrame(encodedHeaders);
        assertEquals(headers, decodedHeaders);
    }

    @Test
    public void testDecodeHeadersFromFrameThrowsFExceptionForBadFrame() throws TException {
        thrown.expect(FProtocolException.class);
        thrown.expectMessage("invalid frame size 3");
        HeaderUtils.decodeFromFrame(new byte[3]);
    }

    @Test
    public void testDecodeHeadersFromFrameThrowsFExceptionForUnsupportedVersion() throws TException {
        thrown.expect(FProtocolException.class);
        thrown.expectMessage("unsupported header version 1");
        HeaderUtils.decodeFromFrame(new byte[]{1, 0, 0, 0, 0});
    }

}
