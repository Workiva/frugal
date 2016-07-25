package com.workiva.frugal.transport;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * FAdapterTransport is an FTransport which wraps a thrift TTransport's
 * read/write operations in a way that is compatible with Frugal. This
 * allows TTransports which support blocking reads to work with Frugal by
 * starting a thread that reads from the underlying transport and calling
 * the registry on received frames.
 */
public class FAdapterTransport extends FTransport {
    private static final Logger LOGGER = LoggerFactory.getLogger(FAdapterTransport.class);

    protected TFramedTransport framedTransport;
    private ProcessorThread processorThread;

    /**
     * Construct a new FAdapterTransport.
     *
     * @param transport  TTransport to wrap
     */
    public FAdapterTransport(TTransport transport) {
        this.framedTransport = new TFramedTransport(transport);
        this.processorThread = new ProcessorThread();
    }

    /**
     * Factory for creating {@link FAdapterTransport} instances.
     */
    public static class Factory implements FTransportFactory {

        /**
         * Returns a new FMuxTransport wrapping the given TTransport.
         *
         * @param transport TTransport to wrap
         * @return new FTransport
         */
        public FAdapterTransport getTransport(TTransport transport) {
            return new FAdapterTransport(transport);
        }
    }

    @Override
    public boolean isOpen() {
        return framedTransport.isOpen() && processorThread != null && registry != null;
    }

    @Override
    public void open() throws TTransportException {
        try {
            framedTransport.open();
        } catch (TTransportException e) {
            // It's OK if the underlying transport is already open.
            if (e.getType() != TTransportException.ALREADY_OPEN) {
                throw e;
            }
        }
        processorThread = new ProcessorThread();
        processorThread.start();
        LOGGER.info("transport opened");
    }

    @Override
    public void close() {
        close(null);
    }

    @Override
    protected void close(final Exception cause) {
        if (isCleanClose(cause) && !isOpen()) {
            return;
        }
        if (processorThread != null) {
            processorThread.kill();
            processorThread = null;
        }
        framedTransport.close();
        super.close(cause);
    }

    /**
     * Determines if the transport close caused by the given exception was a "clean" close, i.e. the exception is null
     * (closed by user) or it's a TTransportException.END_OF_FILE (remote peer closed).
     *
     * @param cause exception which caused the close
     * @return true if the close was clean, false if not
     */
    private boolean isCleanClose(Exception cause) {
        if (cause == null) {
            return true;
        }
        if (cause instanceof TTransportException) {
            return ((TTransportException) cause).getType() == TTransportException.END_OF_FILE;
        }
        return false;
    }

    @Override
    public void write(byte[] var1, int var2, int var3) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.NOT_OPEN);
        }
        framedTransport.write(var1, var2, var3);
    }

    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.NOT_OPEN);
        }
        framedTransport.flush();
    }

    private class ProcessorThread extends Thread {

        private volatile boolean running;

        public ProcessorThread() {
            setName("processor");
        }

        public void kill() {
            if (this != Thread.currentThread()) {
                interrupt();
            }
            running = false;
        }

        public void run() {
            running = true;
            while (running) {
                byte[] frameBytes;
                try {
                    frameBytes = framedTransport.readFrame();
                } catch (TTransportException e) {
                    if (e.getType() != TTransportException.END_OF_FILE) {
                        LOGGER.warn("error reading frame, closing transport " + e.getMessage());
                    }
                    close(e);
                    return;
                }

                try {
                    registry.execute(frameBytes);
                } catch (TException e) {
                    // An exception here indicates an unrecoverable exception,
                    // tear down transport.
                    LOGGER.error("closing transport due to unrecoverable error processing frame: " + e.getMessage());
                    close(e);
                    return;
                }
            }
        }
    }
}
