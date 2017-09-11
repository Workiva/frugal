/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.workiva.frugal.protocol;

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
 * Functions for encoding and decoding Frugal HeaderUtils.
 */
public class HeaderUtils {

    // Version 0
    public static final byte V0 = 0x00;

    /**
     * Encode a map of headers into a byte sequence.
     *
     * @param headers headers to encode
     * @return headers encoded as a byte sequence.
     *
     * @throws TException if error encoding headers
     */
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

    /**
     * Read headers in a transport buffer.
     *
     * @param transport a transport buffering header information
     * @return headers as key-value pairs
     *
     * @throws TException if error reading headers
     */
    public static Map<String, String> read(TTransport transport) throws TException {
        byte[] buff = new byte[5];

        // Read version
        transport.readAll(buff, 0, 1);

        // Support more versions when available
        if (buff[0] != V0) {
            throw new TProtocolException(TProtocolException.BAD_VERSION, "unsupported header version " + buff[0]);
        }

        // Read size
        transport.readAll(buff, 1, 4);
        int size = ProtocolUtils.readInt(buff, 1);
        buff = new byte[size];
        transport.readAll(buff, 0, size);

        return readPairs(buff, 0, size);
    }

    /**
     * Decodes header information from a byte sequence.
     *
     * @param bytes a sequence of framed bytes
     * @return Map of headers
     *
     * @throws TException if invalid data
     */
    public static Map<String, String> decodeFromFrame(byte[] bytes) throws TException {
        if (bytes.length < 5) {
            throw new TProtocolException(TProtocolException.INVALID_DATA, "invalid frame size " + bytes.length);
        }

        // Support more versions when available
        if (bytes[0] != V0) {
            throw new TProtocolException(TProtocolException.BAD_VERSION, "unsupported header version " + bytes[0]);
        }

        return readPairs(bytes, 5, ProtocolUtils.readInt(bytes, 1) + 5);
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
                    throw new TProtocolException(TProtocolException.INVALID_DATA, "invalid protocol header name");
                }
                byte[] nameBuff = Arrays.copyOfRange(buff, i, nameSize + i);
                i += nameSize;
                String name = new String(nameBuff, "UTF-8");

                // Read header value
                int valueSize = ProtocolUtils.readInt(buff, i);
                i += 4;
                if (i > end || i + valueSize > end) {
                    throw new TProtocolException(TProtocolException.INVALID_DATA, "invalid protocol header value");
                }
                byte[] valueBuff = Arrays.copyOfRange(buff, i, valueSize + i);
                i += valueSize;
                String value = new String(valueBuff, "UTF-8");

                headers.put(name, value);
            } catch (IOException e) {
                throw new TTransportException("could not read header bytes", e);
            }
        }
        return headers;
    }

}
