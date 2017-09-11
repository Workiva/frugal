package com.workiva.frugal.transport.monitor;

import com.workiva.frugal.transport.FTransport;
import com.workiva.frugal.transport.FTransportClosedCallback;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Runs a transport monitor.
 */
public class MonitorRunner implements FTransportClosedCallback {

    private static final Logger LOGGER = LoggerFactory.getLogger(MonitorRunner.class);

    private FTransportMonitor monitor;
    private FTransport transport;

    public MonitorRunner(FTransportMonitor monitor, FTransport transport) {
        this.monitor = monitor;
        this.transport = transport;
    }

    @Override
    public void onClose(Exception cause) {
        if (cause == null) {
            handleCleanClose();
        } else {
            handleUncleanClose(cause);
        }
    }

    private void handleCleanClose() {
        LOGGER.info("FTransport Monitor: FTransport was closed cleanly.");
        monitor.onClosedCleanly();
    }

    private void handleUncleanClose(Exception cause) {
        LOGGER.warn("FTransport Monitor: FTransport was closed uncleanly because: " + cause.getMessage());
        long wait = monitor.onClosedUncleanly(cause);
        if (wait < 0) {
            LOGGER.warn("FTransport Monitor: Instructed not to reopen.");
            return;
        }
        attemptReopen(wait);
    }

    private void attemptReopen(long initialWait) {
        long wait = initialWait;
        long prevAttempts = 0;

        while (wait >= 0) {
            LOGGER.info("FTransport Monitor: Attempting to reopen after " + wait + " ms");
            try {
                Thread.sleep(wait);
            } catch (InterruptedException e) {
                LOGGER.warn("FTransport Monitor: Reconnect wait interrupted: " + e.getMessage());
            }

            try {
                transport.open();
            } catch (TTransportException e) {
                LOGGER.warn("FTransport Monitor: Failed to reopen transport due to: " + e.getMessage());
                prevAttempts++;

                wait = monitor.onReopenFailed(prevAttempts, wait);
                continue;
            }

            LOGGER.info("FTransport Monitor: Successfully reopened!");
            monitor.onReopenSucceeded();
            return;
        }

        LOGGER.warn("FTransport Monitor: ReopenFailed callback instructed not to reopen. Terminating...");
    }

}
