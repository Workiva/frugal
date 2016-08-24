package com.workiva.frugal.internal;

import com.workiva.frugal.exception.FProtocolException;
import com.workiva.frugal.util.Pair;
import com.workiva.frugal.util.ProtocolUtils;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TProtocolException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * This is an internal-only class. Don't use it!
 */
public class Headers {
    // Version 0
    private static final byte V0 = 0x00;

    public static byte[] encode(Map<String, String> headers) throws TException {
        int size = 0;
        if (headers == null) {
            headers = new HashMap<>();
        }

        List<Pair<byte[], byte[]>> utf8Headers = new ArrayList<>();
        // Get total frame size headers
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            byte[] keyBytes = ProtocolUtils.encodeString(pair.getKey());
            byte[] valueBytes = ProtocolUtils.encodeString(pair.getValue());
            size += 8 + keyBytes.length + valueBytes.length;
            utf8Headers.add(Pair.of(keyBytes, valueBytes));
        }

        byte[] buff = new byte[size + 5];

        // Write version
        buff[0] = V0;

        // Write size
        ProtocolUtils.writeInt(size, buff, 1);

        int i = 5;
        // Write headers
        for (Pair<byte[], byte[]> pair : utf8Headers) {
            // Write key
            byte[] key = pair.getLeft();
            ProtocolUtils.writeInt(key.length, buff, i);
            i += 4;
            ProtocolUtils.writeStringBytes(key, buff, i);
            i += key.length;

            // Write value
            byte[] value = pair.getRight();
            ProtocolUtils.writeInt(value.length, buff, i);
            i += 4;
            ProtocolUtils.writeStringBytes(value, buff, i);
            i += value.length;
        }
        return buff;
    }

    public static Map<String, String> read(TTransport transport) throws TException {
        byte[] buff = new byte[5];

        // Read version
        transport.readAll(buff, 0, 1);

        // Support more versions when available
        if (buff[0] != V0) {
            throw new FProtocolException(FProtocolException.BAD_VERSION, "unsupported header version " + buff[0]);
        }

        // Read size
        transport.readAll(buff, 1, 4);
        int size = ProtocolUtils.readInt(buff, 1);
        buff = new byte[size];
        transport.readAll(buff, 0, size);

        return readPairs(buff, 0, size);
    }

    public static Map<String, String> decodeFromFrame(byte[] frame) throws TException {
        if (frame.length < 5) {
            throw new FProtocolException(FProtocolException.INVALID_DATA, "invalid frame size " + frame.length);
        }

        // Support more versions when available
        if (frame[0] != V0) {
            throw new FProtocolException(FProtocolException.BAD_VERSION, "unsupported header version " + frame[0]);
        }

        return readPairs(frame, 5, ProtocolUtils.readInt(frame, 1) + 5);
    }

    private static Map<String, String> readPairs(byte[] buff, int start, int end) throws TException {
        Map<String, String> headers = new HashMap<>();
        int i = start;
        while (i < end) {
            try {
                // Read header name
                int nameSize = ProtocolUtils.readInt(buff, i);
                i += 4;
                if (i > end || i + nameSize > end) {
                    throw new FProtocolException(TProtocolException.INVALID_DATA, "invalid protocol header name");
                }
                byte[] nameBuff = Arrays.copyOfRange(buff, i, nameSize + i);
                i += nameSize;
                String name = new String(nameBuff, "UTF-8");

                // Read header value
                int valueSize = ProtocolUtils.readInt(buff, i);
                i += 4;
                if (i > end || i + valueSize > end) {
                    throw new FProtocolException(FProtocolException.INVALID_DATA, "invalid protocol header value");
                }
                byte[] valueBuff = Arrays.copyOfRange(buff, i, valueSize + i);
                i += valueSize;
                String value = new String(valueBuff, "UTF-8");

                headers.put(name, value);
            } catch (IOException e) {
                throw new TTransportException(FProtocolException.UNKNOWN, "could not read header bytes", e);
            }
        }
        return headers;
    }

}
