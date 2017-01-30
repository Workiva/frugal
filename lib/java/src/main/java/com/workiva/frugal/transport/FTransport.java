package com.workiva.frugal.transport;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.FrugalTTransportExceptionType;
import com.workiva.frugal.transport.monitor.FTransportMonitor;
import com.workiva.frugal.transport.monitor.MonitorRunner;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * FTransport is comparable to Thrift's TTransport in that it represent the transport
 * layer for frugal clients. However, frugal is callback based and sends only framed data.
 * Therefore, instead of exposing <code>read</code>, <code>write</code>, and <code>flush</code>,
 * the transport has a simple <code>request</code> method that sends framed frugal messages and
 * returns the response.
 */
public abstract class FTransport {

    private static final Logger LOGGER = LoggerFactory.getLogger(FTransport.class);

    private volatile FTransportClosedCallback closedCallback;
    private volatile FTransportClosedCallback monitor;
    private boolean isOpen;

    protected int requestSizeLimit;

    public synchronized boolean isOpen() {
        return isOpen;
    }

    /**
     * Opens the transport.
     *
     * @throws TTransportException
     */
    public synchronized void open() throws TTransportException {
        isOpen = true;
    }

    /**
     * Closes the transport.
     */
    public void close() {
        close(null);
    }

    /**
     * Signal close with the given cause.
     *
     * @param cause Exception if not a clean close (null otherwise)
     */
    protected synchronized void close(final Exception cause) {
        isOpen = false;
        signalClose(cause);
    }

    /**
     * Send the given framed frugal payload over the transport.
     * Implementations of <code>oneway</code> should be thread-safe.
     *
     * @param context FContext associated with the request (used for timeout and logging)
     * @param payload framed frugal bytes
     * @throws TTransportException if the request times out or encounters other problems
     */
    public abstract void oneway(FContext context, byte[] payload) throws TTransportException;

    /**
     * Send the given framed frugal payload over the transport and returns the response.
     * Implementations of <code>request</code> should be thread-safe.
     *
     * @param context FContext associated with the request (used for timeout and logging)
     * @param payload framed frugal bytes
     * @return the response in TTransport form
     * @throws TTransportException if the request times out or encounters other problems
     */
    public abstract TTransport request(FContext context, byte[] payload) throws TTransportException;

    /**
     * Get the maximum request size permitted by the transport. If <code>getRequestSizeLimit</code>
     * returns a non-positive number, the transport is assumed to have no request size limit.
     *
     * @return the request size limit
     */
    public int getRequestSizeLimit() {
        return requestSizeLimit;
    }

    /**
     * Set the closed callback for the FTransport.
     *
     * @param closedCallback callback to be invoked when the transport is closed.
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

    protected synchronized void signalClose(final Exception cause) {
        if (closedCallback != null) {
            closedCallback.onClose(cause);
        }
        if (monitor != null) {
            new Thread(() -> monitor.onClose(cause), "transport-monitor").start();
        }
    }

    // Make sure that the transport is in a state that we can send data.
    protected void preflightRequestCheck(int length) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(FrugalTTransportExceptionType.NOT_OPEN);
        }

        int requestSizeLimit = getRequestSizeLimit();
        if (requestSizeLimit > 0 && length > requestSizeLimit) {
            throw new TTransportException(FrugalTTransportExceptionType.REQUEST_TOO_LARGE,
                    String.format("Message exceeds %d bytes, was %d bytes",
                            requestSizeLimit, length));
        }
    }
}
