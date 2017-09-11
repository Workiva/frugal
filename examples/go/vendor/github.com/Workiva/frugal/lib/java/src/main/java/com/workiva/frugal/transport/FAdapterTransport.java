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
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;


/**
 * An implementation of FTransport which uses a provided TTransport for read/write operations in a way that is
 * compatible with Frugal. This allows TTransports which support blocking reads to work with Frugal by starting a
 * thread that reads from the underlying transport and calling <code>handleResponse</code> on received frames.
 */
public class FAdapterTransport extends FAsyncTransport {

    private static final Logger LOGGER = LoggerFactory.getLogger(FAdapterTransport.class);

    private final TTransport transport;
    private final TFramedTransport framedTransport;
    private ExecutorFactory executorFactory;
    private ExecutorService readExecutor;

    /**
     * Factory for creating FAdapterTransport instances.
     */
    public static class Factory implements FTransportFactory {

        @Override
        public FTransport getTransport(TTransport transport) {
            return new FAdapterTransport(transport);
        }

    }

    /**
     * Creates a new FAdapterTransport which wraps the given TTransport.
     *
     * @param tr TTransport to adapt to
     */
    public FAdapterTransport(TTransport tr) {
        transport = tr;
        framedTransport = new TFramedTransport(tr);
        executorFactory = Executors::newSingleThreadExecutor;
    }

    @Override
    public synchronized boolean isOpen() {
        return super.isOpen() && framedTransport.isOpen();
    }

    @Override
    public synchronized void open() throws TTransportException {
        if (isOpen()) {
            throw new TTransportException(TTransportExceptionType.ALREADY_OPEN, "Transport already open");
        }

        try {
            framedTransport.open();
        } catch (TTransportException e) {
            // It's OK if the underlying transport is already open.
            if (e.getType() != TTransportExceptionType.ALREADY_OPEN) {
                throw e;
            }
        }

        readExecutor = executorFactory.newExecutor();
        readExecutor.submit(newTransportReader());
        super.open();
    }

    protected void setExecutorFactory(ExecutorFactory factory) {
        executorFactory = factory;
    }

    @Override
    public void close() {
        close(null);
    }

    @Override
    protected synchronized void close(Exception cause) {
        if (isCleanClose(cause) && !isOpen()) {
            return;
        }
        super.close(cause);
        readExecutor.shutdownNow(); // No need to do a clean shutdown
        framedTransport.close();
        if (isCleanClose(cause)) {
            LOGGER.info("transport closed");
        } else {
            LOGGER.info("transport closed with cause: " + cause.getMessage());
        }
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
            return ((TTransportException) cause).getType() == TTransportExceptionType.END_OF_FILE;
        }
        return false;
    }

    @Override
    protected void flush(byte[] payload) throws TTransportException {

        // We need to write to the wrapped transport, not the framed transport, since
        // data given to request is already framed.
        transport.write(payload);
        transport.flush();
    }

    protected Runnable newTransportReader() {
        return new TransportReader();
    }

    private class TransportReader implements Runnable {

        @Override
        public void run() {
            while (true) {
                byte[] frame;
                try {
                    frame = framedTransport.readFrame();
                } catch (TTransportException e) {
                    if (e.getType() == TTransportExceptionType.END_OF_FILE) {
                        // EOF indicates remote peer disconnected.
                        close();
                        return;
                    }

                    LOGGER.error("error reading protocol frame, closing transport: " + e.getMessage());
                    close(e);
                    return;
                }

                try {
                    handleResponse(frame);
                } catch (TException e) {
                    LOGGER.error("closing transport due to unrecoverable error processing frame: " + e.getMessage());
                    close(e);
                    return;
                }
            }
        }

    }

    /**
     * ExecutorFactory creates ExecutorServices.
     */
    protected interface ExecutorFactory {
        ExecutorService newExecutor();
    }

}
