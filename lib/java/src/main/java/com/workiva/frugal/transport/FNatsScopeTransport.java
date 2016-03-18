package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FException;
import com.workiva.frugal.exception.FMessageSizeException;
import com.workiva.frugal.util.ProtocolUtils;
import io.nats.client.*;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;

import java.io.IOException;
import java.io.PipedInputStream;
import java.io.PipedOutputStream;
import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.locks.ReentrantLock;
import java.util.logging.Logger;

/**
 * FNatsScopeTransport implements FScopeTransport by using NATS as the pub/sub message broker. Messages are limited to
 * 1MB in size.
 */
public class FNatsScopeTransport extends FScopeTransport {

    // The number of message frames to buffer on the subscriber.
    private static final int FRAME_BUFFER_SIZE = 5;
    private static final byte[] FRAME_BUFFER_CLOSED = new byte[0];

    private final Connection conn;
    private String subject;
    private BlockingQueue<byte[]> frameBuffer;
    private byte[] currentFrame;
    private int currentFramePos;
    private ByteBuffer writeBuffer;
    private Subscription sub;
    private boolean pull;
    private boolean isOpen;
    private final ReentrantLock lock;

    private static Logger LOGGER = Logger.getLogger(FNatsScopeTransport.class.getName());

    protected FNatsScopeTransport(Connection conn) {
        this.conn = conn;
        this.lock = new ReentrantLock();
    }

    public static class Factory implements FScopeTransportFactory {

        private Connection conn;

        public Factory(Connection conn) {
            this.conn = conn;
        }

        /**
         * Get a new FScopeTransport instance.
         *
         * @return A new FScopeTransport instance.
         */
        public FNatsScopeTransport getTransport() {
            return new FNatsScopeTransport(this.conn);
        }
    }

    @Override
    public void lockTopic(String topic) throws TException {
        if (pull) {
            throw new FException("subscriber cannot lock topic");
        }
        lock.lock();
        subject = topic;
    }

    @Override
    public void unlockTopic() throws TException {
        if (pull) {
            throw new FException("subscriber cannot unlock topic");
        }
        lock.unlock();
        subject = "";
    }

    @Override
    public void subscribe(String topic) throws TException {
        pull = true;
        subject = topic;
        open();
    }

    @Override
    public synchronized boolean isOpen() {
        return conn.getState() == Constants.ConnState.CONNECTED && isOpen;
    }

    @Override
    public synchronized void open() throws TTransportException {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            throw new TTransportException(TTransportException.NOT_OPEN,
                    "NATS not connected, has status " + conn.getState());
        }
        if (isOpen) {
            throw new TTransportException(TTransportException.ALREADY_OPEN, "NATS transport already open");
        }

        if (!pull) {
            writeBuffer = ByteBuffer.allocate(TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE);
            isOpen = true;
            return;
        }

        if ("".equals(subject)) {
            throw new TTransportException("Subject cannot be empty.");
        }

        frameBuffer = new ArrayBlockingQueue<>(FRAME_BUFFER_SIZE);

        sub = conn.subscribe(getFormattedSubject(), new MessageHandler() {
            @Override
            public void onMessage(Message msg) {
                if (msg.getData().length < 4) {
                    LOGGER.warning("discarding invalid scope message frame");
                    return;
                }
                try {
                    // Discard frame size.
                    frameBuffer.put(Arrays.copyOfRange(msg.getData(), 4, msg.getData().length));
                } catch (InterruptedException ignored) {
                }
            }
        });
        isOpen = true;
    }

    @Override
    public synchronized void close() {
        if (!isOpen) {
            return;
        }

        if (!pull) {
            isOpen = false;
            return;
        }
        try {
            sub.unsubscribe();
        } catch (IOException e) {
            LOGGER.warning("could not unsubscribe from subscription. " + e.getMessage());
        }
        sub = null;
        try {
            frameBuffer.put(FRAME_BUFFER_CLOSED);
        } catch (InterruptedException e) {
            LOGGER.warning("could not close write frame buffer. " + e.getMessage());
        }
        isOpen = false;
    }

    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.END_OF_FILE);
        }
        if (currentFrame == null) {
            try {
                currentFrame = frameBuffer.take();
            } catch (InterruptedException e) {
                throw new TTransportException(TTransportException.END_OF_FILE, e.getMessage());
            }
        }
        if (currentFrame == FRAME_BUFFER_CLOSED) {
            throw new TTransportException(TTransportException.END_OF_FILE);
        }
        int size = Math.min(len, currentFrame.length);
        System.arraycopy(currentFrame, currentFramePos, bytes, off, size);
        currentFramePos += size;
        if (currentFramePos == currentFrame.length) {
            // The entire frame was copied, clear it.
            discardFrame();
        }
        return size;
    }

    @Override
    public void discardFrame() {
        currentFrame = null;
        currentFramePos = 0;
    }

    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw TNatsServiceTransport.getClosedConditionException(conn, "write:");
        }
        // Include 4 bytes for frame size.
        if (writeBuffer.remaining() < len + 4) {
            int size = 4 + len + TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE - writeBuffer.remaining();
            writeBuffer.clear();
            throw new FMessageSizeException(
                    String.format("Message exceeds %d bytes, was %d bytes",
                            TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE, size));
        }
        writeBuffer.put(bytes, off, len);
    }

    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw TNatsServiceTransport.getClosedConditionException(conn, "flush:");
        }
        byte[] data = new byte[writeBuffer.position()];
        writeBuffer.flip();
        writeBuffer.get(data);
        if (data.length == 0) {
            return;
        }
        // Include 4 bytes for frame size.
        if (data.length + 4 > TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE) {
            throw new FMessageSizeException(String.format(
                    "Message exceeds %d bytes, was %d bytes",
                    TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE, 4 + data.length));
        }
        byte[] frame = new byte[data.length + 4];
        ProtocolUtils.writeInt(data.length, frame, 0);
        System.arraycopy(data, 0, frame, 4, data.length);
        try {
            conn.publish(getFormattedSubject(), frame);
        } catch (IOException e) {
            throw new TTransportException("flush: unable to publish data: " + e.getMessage());
        }
        writeBuffer.clear();
    }

    private String getFormattedSubject() {
        return TNatsServiceTransport.FRUGAL_PREFIX + this.subject;
    }

}
