package com.workiva.frugal.transport;

import org.apache.thrift.TByteArrayOutputStream;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.apache.thrift.transport.TTransportFactory;

/**
 * TFramedTransport is a buffered TTransport that ensures a fully read message
 * every time by preceding messages with a 4-byte frame size.
 */
class TFramedTransport extends TTransport{

    protected static final int DEFAULT_MAX_LENGTH = 2147483647;

    private int maxLength_;

    /**
     * Underlying transport
     */
    private TTransport transport_ = null;

    /**
     * Buffer for output
     */
    protected final TByteArrayOutputStream writeBuffer_ =
            new TByteArrayOutputStream(1024);

    public static class Factory extends TTransportFactory {
        private int maxLength_;

        public Factory() {
            maxLength_ = TFramedTransport.DEFAULT_MAX_LENGTH;
        }

        public Factory(int maxLength) {
            maxLength_ = maxLength;
        }

        @Override
        public TTransport getTransport(TTransport base) {
            return new TFramedTransport(base, maxLength_);
        }
    }

    /**
     * Constructor wraps around another transport
     */
    public TFramedTransport(TTransport transport, int maxLength) {
        transport_ = transport;
        maxLength_ = maxLength;
    }

    public TFramedTransport(TTransport transport) {
        transport_ = transport;
        maxLength_ = TFramedTransport.DEFAULT_MAX_LENGTH;
    }

    public void open() throws TTransportException {
        transport_.open();
    }

    public boolean isOpen() {
        return transport_.isOpen();
    }

    public void close() {
        transport_.close();
    }

    public int read(byte[] buf, int off, int len) throws TTransportException {
        throw new TTransportException("Cannot read directly from " + getClass().getName());
    }

    private final byte[] readi32buf = new byte[4];
    private final byte[] writei32buf = new byte[4];

    public byte[] readFrame() throws TTransportException {
        transport_.readAll(readi32buf, 0, 4);
        int size = decodeFrameSize(readi32buf);

        if (size < 0) {
            close();
            throw new TTransportException("Read a negative frame size (" + size + ")!");
        }

        if (size > maxLength_) {
            close();
            throw new TTransportException(
                    "Frame size (" + size + ") larger than max length (" + maxLength_ + ")!");
        }

        byte[] buff = new byte[size];
        transport_.readAll(buff, 0, size);
        return buff;
    }

    public void write(byte[] buf, int off, int len) throws TTransportException {
        writeBuffer_.write(buf, off, len);
    }

    @Override
    public void flush() throws TTransportException {
        byte[] buf = writeBuffer_.get();
        int len = writeBuffer_.len();
        writeBuffer_.reset();

        encodeFrameSize(len, writei32buf);
        transport_.write(writei32buf, 0, 4);
        transport_.write(buf, 0, len);
        transport_.flush();
    }

    private static void encodeFrameSize(final int frameSize, final byte[] buf) {
        buf[0] = (byte)(0xff & (frameSize >> 24));
        buf[1] = (byte)(0xff & (frameSize >> 16));
        buf[2] = (byte)(0xff & (frameSize >> 8));
        buf[3] = (byte)(0xff & (frameSize));
    }

    private static int decodeFrameSize(final byte[] buf) {
        return
                ((buf[0] & 0xff) << 24) |
                        ((buf[1] & 0xff) << 16) |
                        ((buf[2] & 0xff) <<  8) |
                        ((buf[3] & 0xff));
    }
}
