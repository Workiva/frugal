package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;
import io.nats.client.*;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;

/**
 * TStatelessNatsTransport is an extension of thrift.TTransport. This is a "stateless" transport in the sense that there
 * is no connection with a server. A request is simply published to a subject and responses are received on another
 * subject. This assumes requests/responses fit within a single NATS message.
 */
public class TStatelessNatsTransport extends TTransport {

    // Controls how many responses to buffer.
    private static final int FRAME_BUFFER_SIZE = 64;
    protected static final byte[] FRAME_BUFFER_CLOSED = new byte[0];
    private static final Logger LOGGER = LoggerFactory.getLogger(TStatelessNatsTransport.class);

    private final Connection conn;
    private final String subject;
    private final String inbox;
    protected final BlockingQueue<byte[]> frameBuffer = new ArrayBlockingQueue<>(FRAME_BUFFER_SIZE);
    private byte[] currentFrame = new byte[0];
    private int currentFramePos;
    private Subscription sub;
    private ByteBuffer requestBuffer = ByteBuffer.allocate(TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE);

    /**
     * Creates a new Thrift TTransport which uses the NATS messaging system as the underlying transport. Unlike
     * TNatsServiceTransport, this TTransport is stateless in that there is no connection maintained between the client
     * and server. A request is simply published to a subject and responses are received on a randomly generated
     * subject. This requires requests to fit within a single NATS message.
     *
     * @param conn    NATS connection
     * @param subject subject to publish requests on
     */
    public TStatelessNatsTransport(Connection conn, String subject) {
        this(conn, subject, conn.newInbox());
    }

    /**
     * Creates a new Thrift TTransport which uses the NATS messaging system as the underlying transport. Unlike
     * TNatsServiceTransport, this TTransport is stateless in that there is no connection maintained between the client
     * and server. A request is simply published to a subject and responses are received on a specified subject. This
     * requires requests to fit within a single NATS message.
     *
     * @param conn    NATS connection
     * @param subject subject to publish requests on
     * @param inbox   subject to receive responses on
     */
    public TStatelessNatsTransport(Connection conn, String subject, String inbox) {
        this.conn = conn;
        this.subject = subject;
        this.inbox = inbox;
    }

    @Override
    public synchronized boolean isOpen() {
        return sub != null && conn.getState() == Constants.ConnState.CONNECTED;
    }

    /**
     * Subscribes to the configured inbox subject.
     *
     * @throws TTransportException
     */
    @Override
    public synchronized void open() throws TTransportException {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            throw new TTransportException(TTransportException.UNKNOWN,
                    String.format("NATS not connected, has status %s", conn.getState().name()));
        }
        if (sub != null) {
            throw new TTransportException(TTransportException.ALREADY_OPEN, "NATS transport already open");
        }
        // Make sure the frame buffer is completely empty (in case this
        // transport was re-opened).
        frameBuffer.clear();

        sub = conn.subscribe(inbox, new MessageHandler() {
            @Override
            public void onMessage(Message message) {
                try {
                    frameBuffer.put(message.getData());
                } catch (InterruptedException ignored) {
                }
            }
        });
    }

    /**
     * Unsubscribes from the inbox subject and closes the response buffer.
     */
    @Override
    public synchronized void close() {
        if (sub == null) {
            return;
        }
        try {
            sub.unsubscribe();
        } catch (IOException e) {
            LOGGER.warn("NATS transport could not unsubscribe from subscription: " + e.getMessage());
        }
        sub = null;
        frameBuffer.clear();
        try {
            frameBuffer.put(FRAME_BUFFER_CLOSED);
        } catch (InterruptedException e) {
            LOGGER.warn("NATS transport could not close frame buffer: " + e.getMessage());
        }
    }

    /**
     * Reads up to len bytes into the buffer.
     *
     * @throws TTransportException
     */
    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.END_OF_FILE);
        }
        if (currentFramePos == currentFrame.length) {
            try {
                currentFrame = frameBuffer.take();
                currentFramePos = 0;
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
        return size;
    }

    /**
     * Writes the bytes to a buffer. Throws FMessageSizeException if the buffer exceeds 1MB.
     *
     * @throws TTransportException
     */
    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw TNatsServiceTransport.getClosedConditionException(conn, "write:");
        }
        if (requestBuffer.remaining() < len) {
            int size = len + TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE - requestBuffer.remaining();
            requestBuffer.clear();
            throw new FMessageSizeException(
                    String.format("Message exceeds %d bytes, was %d bytes",
                            TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE, size));
        }
        requestBuffer.put(bytes, off, len);
    }

    /**
     * Sends the buffered bytes over NATS.
     *
     * @throws TTransportException
     */
    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw TNatsServiceTransport.getClosedConditionException(conn, "flush:");
        }
        byte[] data = new byte[requestBuffer.position()];
        requestBuffer.flip();
        requestBuffer.get(data);
        if (data.length == 0) {
            return;
        }
        try {
            conn.publish(subject, inbox, data);
        } catch (IOException e) {
            throw new TTransportException("flush: unable to publish data: " + e.getMessage());
        }
        requestBuffer.clear();
    }

}
