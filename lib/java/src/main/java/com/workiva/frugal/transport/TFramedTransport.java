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

import com.workiva.frugal.util.ProtocolUtils;
import org.apache.thrift.TByteArrayOutputStream;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.apache.thrift.transport.TTransportFactory;

/**
 * TFramedTransport is a buffered TTransport that ensures a fully read message
 * every time by preceding messages with a 4-byte frame size.
 */
class TFramedTransport extends TTransport {

    protected static final int DEFAULT_MAX_LENGTH = 2147483647;

    private int maxLength;

    /**
     * Underlying transport.
     */
    private TTransport transport = null;

    /**
     * Buffer for output.
     */
    protected final TByteArrayOutputStream writeBuffer =
            new TByteArrayOutputStream(1024);

    public static class Factory extends TTransportFactory {
        private int maxLength;

        public Factory() {
            maxLength = TFramedTransport.DEFAULT_MAX_LENGTH;
        }

        public Factory(int maxLength) {
            this.maxLength = maxLength;
        }

        @Override
        public TTransport getTransport(TTransport base) {
            return new TFramedTransport(base, maxLength);
        }
    }

    /**
     * Constructor wraps around another transport.
     */
    public TFramedTransport(TTransport transport, int maxLength) {
        this.transport = transport;
        this.maxLength = maxLength;
    }

    public TFramedTransport(TTransport transport) {
        this.transport = transport;
        maxLength = TFramedTransport.DEFAULT_MAX_LENGTH;
    }

    public void open() throws TTransportException {
        transport.open();
    }

    public boolean isOpen() {
        return transport.isOpen();
    }

    public void close() {
        transport.close();
    }

    public int read(byte[] buf, int off, int len) throws TTransportException {
        throw new TTransportException("Cannot read directly from " + getClass().getName());
    }

    private final byte[] readi32buf = new byte[4];
    private final byte[] writei32buf = new byte[4];

    public byte[] readFrame() throws TTransportException {
        transport.readAll(readi32buf, 0, 4);
        int size = ProtocolUtils.readInt(readi32buf, 0);

        if (size < 0) {
            close();
            throw new TTransportException("Read a negative frame size (" + size + ")!");
        }

        if (size > maxLength) {
            close();
            throw new TTransportException(
                    "Frame size (" + size + ") larger than max length (" + maxLength + ")!");
        }

        byte[] buff = new byte[size];
        transport.readAll(buff, 0, size);
        return buff;
    }

    public void write(byte[] buf, int off, int len) throws TTransportException {
        writeBuffer.write(buf, off, len);
    }

    @Override
    public void flush() throws TTransportException {
        byte[] buf = writeBuffer.get();
        int len = writeBuffer.len();
        writeBuffer.reset();

        ProtocolUtils.writeInt(len, writei32buf, 0);
        transport.write(writei32buf, 0, 4);
        transport.write(buf, 0, len);
        transport.flush();
    }
}
