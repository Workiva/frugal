package com.workiva.frugal.transport.monitor;


import com.workiva.frugal.protocol.FRegistry;
import com.workiva.frugal.transport.FTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

public class BaseFTransportMonitorTest {

    /**
     * Ensure that onClosedUncleanly returns -1 if max attempts is 0.
     */
    @Test
    public void testOnClosedUncleanlyMaxZero() {
        FTransportMonitor monitor = new BaseFTransportMonitor(0, 0, 0);
        long wait = monitor.onClosedUncleanly(new Exception("error"));
        assertEquals(-1, wait);
    }

    /**
     * Ensure that onClosedUncleanly returns the expected wait period when max attemts > 0.
     */
    @Test
    public void testOnClosedUncleanly() {
        FTransportMonitor monitor = new BaseFTransportMonitor(1, 1, 1);
        long wait = monitor.onClosedUncleanly(new Exception("error"));
        assertEquals(1, wait);
    }

    /**
     * Ensure that onReopenFailed returns -1 if max attempts is reached.
     */
    @Test
    public void testOnReopenFailedMaxAttempts() {
        FTransportMonitor monitor = new BaseFTransportMonitor(1, 0, 0);
        long wait = monitor.onReopenFailed(1, 0);
        assertEquals(-1, wait);
    }

    /**
     * Ensure that onReopenFailed returns double the previous wait.
     */
    @Test
    public void testOnReopenFailed() {
        FTransportMonitor monitor = new BaseFTransportMonitor(6, 1, 10);
        long wait = monitor.onReopenFailed(0, 1);
        assertEquals(2, wait);
    }

    /**
     * Ensure that onReopenFailed respects the max wait.
     */
    @Test
    public void testOnReopenFailedMaxWait() {
        FTransportMonitor monitor = new BaseFTransportMonitor(6, 1, 1);
        long wait = monitor.onReopenFailed(0, 1);
        assertEquals(1, wait);
    }

    /**
     * Ensure that the monitor handles a clean close.
     */
    @Test
    public void testCleanClose() throws InterruptedException {
        FTransport transport = new MockFTransport();
        FTransportMonitor monitor = mock(FTransportMonitor.class);
        transport.setMonitor(monitor);

        transport.close();
        Thread.sleep(5);

        verify(monitor).onClosedCleanly();
    }

    /**
     * Ensure that the monitor handles an unclean close.
     */
    @Test
    public void testUncleanClose() throws InterruptedException {
        MockFTransport transport = new MockFTransport();
        FTransportMonitor monitor = mock(FTransportMonitor.class);
        Exception cause = new Exception("error");
        when(monitor.onClosedUncleanly(cause)).thenReturn((long) -1);
        transport.setMonitor(monitor);

        transport.close(cause);
        Thread.sleep(5);

        verify(monitor).onClosedUncleanly(cause);
    }

    /**
     * Ensure that handleUncleanClose attempts to reopen when onClosedUncleanly instructs it to do so.
     */
    @Test
    public void testHandleUncleanClose() throws InterruptedException {
        MockFTransport transport = new MockFTransport();
        FTransportMonitor monitor = mock(FTransportMonitor.class);
        Exception cause = new Exception("error");
        when(monitor.onClosedUncleanly(cause)).thenReturn((long) 0);
        transport.setMonitor(monitor);

        transport.close(cause);
        Thread.sleep(5);

        verify(monitor).onClosedUncleanly(cause);
        verify(monitor).onReopenSucceeded();
    }

    /**
     * Ensure that attemptReopen retries when OnReopenFailed instructs to do so.
     */
    @Test
    public void testAttemptReopenFailRetrySucceed() throws InterruptedException {
        MockFTransport transport = new MockFTransport(1);
        FTransportMonitor monitor = mock(FTransportMonitor.class);
        Exception cause = new Exception("error");
        when(monitor.onClosedUncleanly(cause)).thenReturn((long) 1);
        when(monitor.onReopenFailed(1, 1)).thenReturn((long) 2);
        transport.setMonitor(monitor);

        transport.close(cause);
        Thread.sleep(10);

        verify(monitor).onClosedUncleanly(cause);
        verify(monitor).onReopenFailed(1, 1);
        verify(monitor).onReopenSucceeded();
    }

    /**
     * Ensure that attemptReopen does not retry when onReopenFailed instructs it not to do so.
     */
    @Test
    public void testAttemptReopenFailNoRetry() throws InterruptedException {
        MockFTransport transport = new MockFTransport(1);
        FTransportMonitor monitor = mock(FTransportMonitor.class);
        Exception cause = new Exception("error");
        when(monitor.onClosedUncleanly(cause)).thenReturn((long) 1);
        when(monitor.onReopenFailed(1, 1)).thenReturn((long) -1);
        transport.setMonitor(monitor);

        transport.close(cause);
        Thread.sleep(5);

        verify(monitor).onClosedUncleanly(cause);
        verify(monitor).onReopenFailed(1, 1);
    }

    private class MockFTransport extends FTransport {

        private int openErrorCount;
        private int errorCount;

        public MockFTransport() {
        }

        public MockFTransport(int openErrorCount) {
            this.openErrorCount = openErrorCount;
        }

        @Override
        public void setRegistry(FRegistry registry) {

        }

        @Override
        public boolean isOpen() {
            return false;
        }

        @Override
        public void open() throws TTransportException {
            if (errorCount < openErrorCount) {
                errorCount++;
                throw new TTransportException(0, "open error");
            }
        }

        @Override
        public void close() {
            signalClose(null);
        }

        @Override
        public int read(byte[] bytes, int i, int i1) throws TTransportException {
            return 0;
        }

        @Override
        public void write(byte[] bytes, int i, int i1) throws TTransportException {

        }

        public void close(Exception cause) {
            signalClose(cause);
        }

    }

}
