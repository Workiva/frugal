package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.logging.Logger;

public class FMuxTransport extends FTransport {
    protected TFramedTransport framedTransport;
    protected BlockingQueue<FrameWrapper> workQueue;
    private ProcessorThread processorThread;
    private WorkerThread[] workerThreads;

    private static Logger LOGGER = Logger.getLogger(FMuxTransport.class.getName());

    /**
     * Construct a new FMuxTransport.
     *
     * @param transport  TTransport to wrap
     * @param numWorkers number of worker thread for the FTransport
     */
    public FMuxTransport(TTransport transport, int numWorkers) {
        this.framedTransport = new TFramedTransport(transport);
        this.workQueue = new ArrayBlockingQueue<>(numWorkers);
        this.processorThread = new ProcessorThread();
        this.workerThreads = new WorkerThread[numWorkers];
    }

    public static class Factory implements FTransportFactory {

        private final int numWorkers;

        /**
         * Construct a new FMuxTransport factory.
         *
         * @param numWorkers number of worker thread for the FTransport
         */
        public Factory(int numWorkers) {
            this.numWorkers = numWorkers;
        }

        /**
         * Returns a new FMuxTransport wrapping the given TTransport.
         *
         * @param transport TTransport to wrap
         * @return new FTransport
         */
        public FMuxTransport getTransport(TTransport transport) {
            return new FMuxTransport(transport, numWorkers);
        }
    }

    private static class FrameWrapper {
        byte[] frameBytes;
        long timestamp;

        protected FrameWrapper(byte[] frameByes, long timestamp) {
            this.frameBytes = frameByes;
            this.timestamp = timestamp;
        }

        protected byte[] getFrameBytes() {
            return frameBytes;
        }

        protected long getTimestamp() {
            return this.timestamp;
        }
    }

    public synchronized void setRegistry(FRegistry registry) {
        if (registry == null) {
            throw new RuntimeException("registry cannot by null");
        }
        if (this.registry != null) {
            return;
        }
        this.registry = registry;
    }

    public synchronized boolean isOpen() {
        return framedTransport.isOpen();
    }

    public synchronized void open() throws TTransportException {
        if (isOpen()) {
            throw new TTransportException("transport already open");
        }
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
        startWorkers();
    }

    public synchronized void close() {
        close(null);
    }

    private synchronized void close(Exception cause) {
        if (registry == null) {
            return;
        }
        framedTransport.close();
        processorThread.kill();
        for (WorkerThread workerThread : workerThreads) {
            workerThread.kill();
        }
        signalClose(cause);
        registry.close();
    }

    public int read(byte[] var1, int var2, int var3) throws TTransportException {
        return framedTransport.read(var1, var2, var3);
    }

    public void write(byte[] var1, int var2, int var3) throws TTransportException {
        framedTransport.write(var1, var2, var3);
    }

    public void flush() throws TTransportException {
        framedTransport.flush();
    }

    private void startWorkers() {
        for (int i = 0; i < workerThreads.length; i++) {
            WorkerThread workerThread = new WorkerThread();
            workerThread.start();
            workerThreads[i] = workerThread;
        }
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
                        LOGGER.warning("error reading frame, closing transport " + e.getMessage());
                    }
                    close(e);
                    return;
                }

                try {
                    FrameWrapper frame = new FrameWrapper(frameBytes, System.currentTimeMillis());
                    workQueue.put(frame);
                } catch (InterruptedException e) {
                    LOGGER.warning("could not put frame in work queue. Dropping frame.");
                }
            }
        }
    }

    private class WorkerThread extends Thread {

        private volatile boolean running;

        public WorkerThread() {
            setName("worker");
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
                FrameWrapper frame;
                try {
                    frame = workQueue.take();
                } catch (InterruptedException e) {
                    // Just keep trying!
                    continue;
                }
                long duration = System.currentTimeMillis() - frame.getTimestamp();
                if (duration > getHighWatermark()) {
                    LOGGER.warning("frame spent " + duration + "ms in the transport buffer, your consumer might be backed up");
                }
                try {
                    registry.execute(frame.getFrameBytes());
                } catch (TException e) {
                    // An exception here indicates an unrecoverable exception,
                    // tear down transport.
                    LOGGER.severe("closing transport due to unrecoverable error processing frame: " + e.getMessage());
                    close(e);
                    return;
                }
            }
        }
    }
}
