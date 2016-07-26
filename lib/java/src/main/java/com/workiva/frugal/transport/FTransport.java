package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FException;
import com.workiva.frugal.exception.FMessageSizeException;
import com.workiva.frugal.protocol.FAsyncCallback;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FRegistry;
import com.workiva.frugal.transport.monitor.FTransportMonitor;
import com.workiva.frugal.transport.monitor.MonitorRunner;
import com.workiva.frugal.util.ProtocolUtils;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.ByteArrayOutputStream;
import java.util.Arrays;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;

/**
 * FTransport is Frugal's equivalent of Thrift's TTransport. FTransport extends
 * TTransport and exposes some additional methods. An FTransport has an FRegistry,
 * so it provides methods for setting the FRegistry and registering and unregistering
 * an FAsyncCallback to an FContext.
 */
public abstract class FTransport extends TTransport {

    private static final Logger LOGGER = LoggerFactory.getLogger(FTransport.class);

    public static final int REQUEST_TOO_LARGE = 100;
    public static final int RESPONSE_TOO_LARGE = 101;

    private volatile FTransportClosedCallback closedCallback;
    private volatile FTransportClosedCallback monitor;
    protected FRegistry registry;
    private boolean isOpen;

    // Write buffer
    private final int writeBufferSize;
    private final ByteArrayOutputStream writeBuffer;

    // TODO: Remove with 2.0
    // Closed callback
    private volatile FClosedCallback fClosedCallback;
    // Read buffer
    protected final BlockingQueue<byte[]> frameBuffer;
    protected static final byte[] FRAME_BUFFER_CLOSED = new byte[0];
    private byte[] currentFrame = new byte[0];
    private int currentFramePos;
    // Watermark
    public static final long DEFAULT_WATERMARK = 5 * 1000;
    protected long highWatermark = DEFAULT_WATERMARK;

    /**
     * Default constructor for FTransport with no inherent read/write capabilities.
     */
    protected FTransport() {
        this.writeBufferSize = 0;
        this.writeBuffer = null;
        this.frameBuffer = null;
    }

    /**
     * Construct an FTransport with the given writeBufferSize.
     *
     * @param writeBufferSize maximum number of bytes allowed to be written to
     *                        the transport before getWriteBytes is called
     */
    protected FTransport(int writeBufferSize) {
        if (writeBufferSize <= 0) {
            // No size limit
            this.writeBufferSize = 0;
            this.writeBuffer = new ByteArrayOutputStream();
        } else {
            // Cap the size of the buffer
            this.writeBufferSize = writeBufferSize;
            this.writeBuffer = new ByteArrayOutputStream(writeBufferSize);
        }
        this.frameBuffer = null;
    }

    /**
     * Construct an FTransport which supports reads.
     *
     * @deprecated Construct callback-based transports instead
     * TODO: Remove with 2.0
     */
    @Deprecated
    FTransport(int requestBufferSize, int frameBufferSize) {
        this.writeBufferSize = requestBufferSize;
        this.writeBuffer = new ByteArrayOutputStream(requestBufferSize);
        this.frameBuffer = new ArrayBlockingQueue<>(frameBufferSize);
    }

    @Override
    public synchronized boolean isOpen() {
        return isOpen;
    }

    @Override
    public synchronized void open() throws TTransportException {
        isOpen = true;
    }

    /**
     * Closes the transport.
     */
    @Override
    public void close() {
        close(null);
    }

    /**
     * Close the frame buffer and signal close.
     *
     * @param cause Exception if not a clean close (null otherwise)
     */
    protected synchronized void close(final Exception cause) {
        // TODO: Remove all read logic with 2.0
        if (frameBuffer != null) {
            try {
                frameBuffer.put(FRAME_BUFFER_CLOSED);
            } catch (InterruptedException e) {
                LOGGER.warn("could not close frame buffer: " + e.getMessage());
            }
        }

        if (registry != null) {
            registry.close();
        }
        isOpen = false;
        signalClose(cause);
    }

    /**
     * With callback-based FTransports (i.e. all transports with the release of 2.0),
     * this will throw an UnsupportedOperationException.
     *
     * Reads up to len bytes into the buffer.
     *
     * @throws TTransportException
     *
     * TODO: Remove all read logic with 2.0
     */
    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        if (frameBuffer == null) {
            throw new UnsupportedOperationException("Do not call read directly on FTransport");
        }

        // TODO: Remove this with 2.0
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
     * Writes the bytes to a buffer. Throws FMessageSizeException if the buffer exceeds
     * {@code writeBufferSize}.
     *
     * @throws TTransportException
     */
    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        if (writeBuffer == null) {
            throw new UnsupportedOperationException("No write buffer set on FTranspprt");
        }

        if (!isOpen()) {
            throw new TTransportException(TTransportException.NOT_OPEN);
        }
        if (writeBufferSize > 0 && writeBuffer.size() + len > writeBufferSize) {
            int size = writeBuffer.size() + len;
            writeBuffer.reset();
            throw new FMessageSizeException(
                    String.format("Message exceeds %d bytes, was %d bytes",
                            writeBufferSize, size));
        }
        writeBuffer.write(bytes, off, len);
    }

    /**
     * Set the FRegistry on the FTransport.
     *
     * @param registry FRegistry to set on the FTransport.
     */
    public synchronized void setRegistry(FRegistry registry) {
        if (registry == null) {
            throw new IllegalArgumentException("registry cannot by null");
        }
        // TODO: With 2.0, consider throwing a RuntimeException.
        // Currently, the generated code sets the registry for each
        // extending service for a particular base service.
        if (this.registry != null) {
            return;
        }
        this.registry = registry;
    }

    /**
     * Register a callback for the given FContext.
     *
     * @param context  the FContext to register.
     * @param callback the callback to register.
     */
    public synchronized void register(FContext context, FAsyncCallback callback) throws TException {
        if (registry == null) {
            throw new FException("registry not set");
        }
        registry.register(context, callback);
    }

    /**
     * Unregister the callback for the given FContext.
     *
     * @param context the FContext to unregister.
     */
    public synchronized void unregister(FContext context) throws TException {
        if (registry == null) {
            throw new FException("registry not set");
        }
        registry.unregister(context);
    }

    protected synchronized FRegistry getRegistry() {
        return registry;
    }

    /**
     * Set the closed callback for the FTransport.
     *
     * @param closedCallback
     * @deprecated use {@link #setClosedCallback(FTransportClosedCallback)} instead.
     */
    @Deprecated
    public synchronized void setClosedCallback(FClosedCallback closedCallback) {
        this.fClosedCallback = closedCallback;
    }

    /**
     * Set the closed callback for the FTransport.
     *
     * @param closedCallback
     */
    public synchronized void setClosedCallback(FTransportClosedCallback closedCallback) {
        this.closedCallback = closedCallback;
    }

    /**
     * Starts a monitor that can watch the health of, and reopen, the transport.
     *
     * @param monitor the FTransportMonitor to set.
     */
    public synchronized void setMonitor(FTransportMonitor monitor) {
        LOGGER.info("FTransport Monitor: Beginning to monitor transport...");
        this.monitor = new MonitorRunner(monitor, this);
    }

    /**
     * Sets the maximum amount of time a frame is allowed to await processing
     * before triggering transport overload logic.
     *
     * @param watermark the watermark time in milliseconds.
     *
     * @deprecated - Implementing may use a watermark as a constructor option.
     * TODO: Remove this with 2.0
     */
    @Deprecated
    public synchronized void setHighWatermark(long watermark) {
        this.highWatermark = watermark;
    }

    @Deprecated
    protected synchronized long getHighWatermark() {
        return highWatermark;
    }

    /**
     * Queries whether there is write data.
     */
    protected boolean hasWriteData() {
        if (writeBuffer == null) {
            throw new UnsupportedOperationException("No write buffer set on FTranspprt");
        }
        return writeBuffer.size() > 0;
    }

    /**
     * Get the write bytes.
     *
     * @return write bytes
     *
     * @deprecated - Get the framed bytes
     * TODO: Remove this with 2.0
     */
    @Deprecated
    protected byte[] getWriteBytes() {
        if (writeBuffer == null) {
            throw new UnsupportedOperationException("No write buffer set on FTranspprt");
        }
        return writeBuffer.toByteArray();
    }

    /**
     * Get the framed write bytes.
     *
     * @return framed write bytes
     */
    protected byte[] getFramedWriteBytes() {
        if (writeBuffer == null) {
            throw new UnsupportedOperationException("No write buffer set on FTranspprt");
        }
        int numBytes = writeBuffer.size();
        byte[] data = new byte[numBytes + 4];
        ProtocolUtils.writeInt(numBytes, data, 0);
        System.arraycopy(writeBuffer.toByteArray(), 0, data, 4, numBytes);
        return data;
    }

    /**
     * Reset the write buffer.
     */
    protected void resetWriteBuffer() {
        if (writeBuffer == null) {
            throw new UnsupportedOperationException("No write buffer set on FTranspprt");
        }
        writeBuffer.reset();
    }

    /**
     * Execute a frugal frame (NOTE: this frame must include the frame size).
     *
     * @param frame frugal frame
     * @throws TException
     */
    protected void executeFrame(byte[] frame) throws TException {
        if (registry == null) {
            throw new FException("registry not set");
        }
        registry.execute(Arrays.copyOfRange(frame, 4, frame.length));
    }

    protected synchronized void signalClose(final Exception cause) {
        // TODO: Remove deprecated callback in future release.
        if (fClosedCallback != null) {
            fClosedCallback.onClose();
        }
        if (closedCallback != null) {
            closedCallback.onClose(cause);
        }
        if (monitor != null) {
            new Thread(() -> {
                monitor.onClose(cause);
            }, "transport-monitor").start();
        }
    }

}
