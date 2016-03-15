package com.workiva.frugal.transport.monitor;

/**
 * BaseFTransportMonitor is a default monitor implementation that attempts to reopen a closed transport with exponential
 * backoff behavior and a capped number of retries. Its behavior can be customized by extending this class and
 * overriding desired callbacks.
 */
public class BaseFTransportMonitor implements FTransportMonitor {

    private static final long DEFAULT_MAX_REOPEN_ATTEMPTS = 60;
    private static final long DEFAULT_INITIAL_WAIT = 2 * 1000;
    private static final long DEFAULT_MAX_WAIT = 2 * 1000;

    protected long maxReopenAttempts;
    protected long initialWait;
    protected long maxWait;

    /**
     * Creates a BaseFTransportMonitor with default reconnect options (attempts to reconnect 60 times with 2 seconds
     * between each attempt).
     */
    public BaseFTransportMonitor() {
        this(DEFAULT_MAX_REOPEN_ATTEMPTS, DEFAULT_INITIAL_WAIT, DEFAULT_MAX_WAIT);
    }

    /**
     * Creates a BaseFTransportMonitor with the specified reconnect behavior.
     *
     * @param maxReopenAttempts the max number of reconnect attempts to make.
     * @param initialWait       the number of milliseconds to wait before attempting to reconnect for the first time.
     * @param maxWait           the max number of milliseconds to wait between reconnect attempts (to cap exponential
     *                          backoff).
     */
    public BaseFTransportMonitor(long maxReopenAttempts, long initialWait, long maxWait) {
        this.maxReopenAttempts = maxReopenAttempts;
        this.initialWait = initialWait;
        this.maxWait = maxWait;
    }

    @Override
    public void onClosedCleanly() {
    }

    @Override
    public long onClosedUncleanly(Exception cause) {
        return maxReopenAttempts > 0 ? initialWait : -1;
    }

    @Override
    public long onReopenFailed(long prevAttempts, long prevWait) {
        if (prevAttempts >= maxReopenAttempts) {
            return -1;
        }

        long nextWait = prevWait * 2;
        if (nextWait > maxWait) {
            nextWait = maxWait;
        }
        return nextWait;
    }

    @Override
    public void onReopenSucceeded() {
    }

}
