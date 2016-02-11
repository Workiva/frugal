package com.workiva.frugal.internal;

import com.workiva.frugal.exception.FException;
import com.workiva.frugal.util.ProtocolUtils;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

import java.io.*;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

/**
 * This is an internal-only class. Don't use it!
 */
public class Headers {
    // Version 0
    private static final byte V0 = 0x00;

    public static byte[] encode(Map<String, String> headers) throws FException {
        int size = 0;
        if (headers == null) {
            headers = new HashMap<>();
        }

        // Get total frame size headers
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            size += 8 + pair.getKey().length() + pair.getValue().length();
        }

        byte[] buff = new byte[size + 5];

        // Write version
        buff[0] = V0;

        // Write size
        ProtocolUtils.writeInt(size, buff, 1);

        int i = 5;
        // Write headers
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            // Write key
            String key = pair.getKey();
            ProtocolUtils.writeInt(key.length(), buff, i);
            i += 4;
            ProtocolUtils.writeString(key, buff, i);
            i += key.length();

            // Write value
            String value = pair.getValue();
            ProtocolUtils.writeInt(value.length(), buff, i);
            i += 4;
            ProtocolUtils.writeString(value, buff, i);
            i += value.length();
        }
        return buff;
    }

    public static Map<String, String> read(TTransport transport) throws FException {
        byte[] buff = new byte[5];

        // Read version
        try {
            transport.readAll(buff, 0, 1);
        } catch (TTransportException e) {
            throw new FException("could not read header version", e);
        }

        // Support more versions when available
        if (buff[0] != V0) {
            throw new FException("unsupported header version " + buff[0]);
        }

        // Read size
        try {
            transport.readAll(buff, 1, 4);
        } catch (TTransportException e) {
            throw new FException("could not read header version", e);
        }
        int size = ProtocolUtils.readInt(buff, 1);
        buff = new byte[size];
        try {
            transport.readAll(buff, 0, size);
        } catch (TTransportException e) {
            throw new FException("could not read headers from transport ", e);
        }

        return readPairs(buff, 0, size);
    }

    public static Map<String, String> decodeFromFrame(byte[] frame) throws FException {
        if (frame.length < 5) {
            throw new FException("invalid frame size " + frame.length);
        }

        // Support more versions when available
        if (frame[0] != V0) {
            throw new FException("unsupported header version " + frame[0]);
        }

        return readPairs(frame, 5, ProtocolUtils.readInt(frame, 1) + 5);
    }

    private static Map<String, String> readPairs(byte[] buff, int start, int end) throws FException {
        Map<String, String> headers = new HashMap<>();
        int i = start;
        while (i < end) {
            try {
                // Read header name
                int nameSize = ProtocolUtils.readInt(buff, i);
                i += 4;
                if (i > end || i + nameSize > end) {
                    throw new FException("invalid protocol header name");
                }
                byte[] nameBuff = Arrays.copyOfRange(buff, i, nameSize + i);
                i += nameSize;
                String name = new String(nameBuff, "UTF-8");

                // Read header value
                int valueSize = ProtocolUtils.readInt(buff, i);
                i += 4;
                if (i > end || i + valueSize > end) {
                    throw new FException("invalid protocol header value");
                }
                byte[] valueBuff = Arrays.copyOfRange(buff, i, valueSize + i);
                i += valueSize;
                String value = new String(valueBuff, "UTF-8");

                headers.put(name, value);
            } catch (IOException e) {
                throw new FException("could not read header bytes, possible protocol error", e);
            }
        }
        return headers;
    }

}
