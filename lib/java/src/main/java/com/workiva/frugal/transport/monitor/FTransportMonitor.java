package com.workiva.frugal.transport.monitor;

/**
 * FTransportMonitor watches and heals an FTransport.
 */
public interface FTransportMonitor {

    /**
     * Called when the transport is closed cleanly by a call to close().
     */
    void onClosedCleanly();

    /**
     * Called when the transport is closed for a reason *other* than a call to Close(). Returns the number of
     * milliseconds to wait before attempting to reopen the transport or a negative number indicating not to attempt to
     * reopen.
     *
     * @param cause the exception causing the transport to be closed.
     * @return milliseconds to wait before attempting to reopen the transport. A negative value means the transport will
     * not attempt to be reopened.
     */
    long onClosedUncleanly(Exception cause);

    /**
     * Called when an attempt to reopen the transport fails.
     *
     * @param prevAttempts the number of previous attempts to reopen the transport.
     * @param prevWait     the length, in milliseconds, of the previous wait.
     * @return milliseconds to wait before attempting to reopen the transport. A negative value means the transport will
     * not attempt to be reopened.
     */
    long onReopenFailed(long prevAttempts, long prevWait);

    /**
     * Called after the transport has been successfully reopened.
     */
    void onReopenSucceeded();

}
