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

package com.workiva.frugal.util;

import java.nio.charset.Charset;

/**
 * Utilities for reading/writing protocol data.
 */
public class ProtocolUtils {

    /**
     * Read an integer from the buffer starting at an offset.
     *
     * @param buff buffer of bytes to read from
     * @param offset starting point to read from
     *
     * @return int read from buffer
     */
    public static int readInt(byte[] buff, int offset) {
        return ((buff[offset] & 0xff) << 24) |
                ((buff[offset + 1] & 0xff) << 16) |
                ((buff[offset + 2] & 0xff) << 8) |
                (buff[offset + 3] & 0xff);
    }

    /**
     * Writes an integer into a buffer starting at a certain offset.
     *
     * @param i The int to write.
     * @param buff The buffer to write into.
     * @param offset The position in buff to start writing at.
     */
    public static void writeInt(int i, byte[] buff, int offset) {
        buff[offset] = (byte) (0xff & (i >> 24));
        buff[offset + 1] = (byte) (0xff & (i >> 16));
        buff[offset + 2] = (byte) (0xff & (i >> 8));
        buff[offset + 3] = (byte) (0xff & (i));
    }

    /**
     * Writes a string into a buffer, starting at a certain offset.
     *
     * @param s The String to write.
     * @param buff The buffer to write into.
     * @param offset The position in buff to start writing at.
     */
    public static void writeString(String s, byte[] buff, int offset) {
        byte[] strBytes = Charset.forName("UTF-8").encode(s).array();
        System.arraycopy(strBytes, 0, buff, offset, s.length());
    }

    /**
     * Encodes a string using UTF-8.
     *
     * @param s The string to encode.
     * @return The bytes representing the string.
     */
    public static byte[] encodeString(String s) {
        return s.getBytes(Charset.forName("UTF-8"));
    }

    /**
     * Writes the bytes corresponding to a UTF-8 encoded string into a buffer,
     * starting at a certain offset.
     *
     * @param strBytes The bytes to write.
     * @param buff The buffer to write into.
     * @param offset The position in buff to start writing at.
     */
    public static void writeStringBytes(byte[] strBytes, byte[] buff, int offset) {
        System.arraycopy(strBytes, 0, buff, offset, strBytes.length);
    }

}
