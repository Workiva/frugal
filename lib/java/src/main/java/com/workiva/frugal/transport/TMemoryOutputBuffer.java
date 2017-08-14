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

package com.workiva.frugal.transport;

import com.workiva.frugal.exception.TTransportExceptionType;
import com.workiva.frugal.util.ProtocolUtils;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

import java.io.ByteArrayOutputStream;

/**
 * An implementation of a framed TTransport using a memory buffer and is used exclusively for writing.
 * The size of this buffer is optionally limited. If limited, writes which cause the buffer to exceed
 * its size limit throw an TTransportException with code TTransportExceptionType.REQUEST_TOO_LARGE.
 */
public class TMemoryOutputBuffer extends TTransport {

    private ByteArrayOutputStream buffer;
    private final int limit;
    private final byte[] emptyFrameSize = new byte[4];

    /**
     * Create an TMemoryOutputBuffer with no buffer size limit.
     */
    public TMemoryOutputBuffer() {
        this(0);
    }

    /**
     * Create an TMemoryOutputBuffer with a buffer size limit.
     *
     * @param size the size limit of the buffer. Note: If <code>size</code> is non-positive,
     *             no limit will be enforced on the buffer.
     */
    public TMemoryOutputBuffer(int size) {
        buffer = new ByteArrayOutputStream();
        limit = size;
        init();
    }

    /**
     * Write the 4 bytes into the buffer. This is an optimization: when {@link #getWriteBytes() getWriteBytes}
     * is called, we write the actual frame size into these 4 bytes.
     */
    private void init() {
        buffer.write(emptyFrameSize, 0, 4);
    }

    @Override
    public boolean isOpen() {
        return true;
    }

    @Override
    public void open() throws TTransportException {
        /* Do nothing */
    }

    @Override
    public void close() {
        /* Do nothing */
    }

    @Override
    public int read(byte[] buf, int off, int len) throws TTransportException {
        throw new UnsupportedOperationException("Cannot read from " + getClass().getCanonicalName());
    }

    @Override
    public void write(byte[] buf, int off, int len) throws TTransportException {
        if (limit > 0 && buffer.size() + len > limit) {
            reset();
            throw new TTransportException(
                    TTransportExceptionType.REQUEST_TOO_LARGE, String.format("Buffer size reached (%d)", limit));
        }
        buffer.write(buf, off, len);
    }

    /**
     * Query if data has been written to the transport.
     *
     * @return true if data written to transport.
     */
    public boolean hasWriteData() {
        return size() > 4;
    }

    /**
     * Return the number of bytes that have been written to the transport.
     *
     * @return the number of bytes written to the transport including the frame size.
     */
    public int size() {
        return buffer.size();
    }

    /**
     * Get the framed bytes that have been written to the transport.
     *
     * @return the bytes written to the transport prepended with 4 frame size bytes.
     */
    public byte[] getWriteBytes() {
        byte[] unframed = buffer.toByteArray();
        // Frame the bytes
        ProtocolUtils.writeInt(unframed.length - 4, unframed, 0);
        return unframed;
    }

    /**
     * Clear the write buffer and initialize the frame size.
     */
    public void reset() {
        buffer.reset();
        init();
    }

}
