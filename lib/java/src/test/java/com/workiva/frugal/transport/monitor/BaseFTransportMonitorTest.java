package com.workiva.frugal.transport.monitor;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.FrugalTTransportExceptionType;
import com.workiva.frugal.transport.FTransport;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;
import org.mockito.Mockito;
import org.mockito.invocation.InvocationOnMock;
import org.mockito.stubbing.Answer;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link BaseFTransportMonitor}.
 */
@RunWith(JUnit4.class)
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
        CountDownLatch latch = new CountDownLatch(1);
        Mockito.doAnswer(invocation -> {
            latch.countDown();
            return null;
        }).when(monitor).onClosedCleanly();
        transport.setMonitor(monitor);

        transport.close();
        latch.await(1, TimeUnit.SECONDS);

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
        CountDownLatch latch = new CountDownLatch(1);
        when(monitor.onClosedUncleanly(cause)).thenAnswer(new Answer<Long>() {
            public Long answer(InvocationOnMock invocation) {
                latch.countDown();
                return -1L;
            }
        });
        transport.setMonitor(monitor);

        transport.close(cause);
        latch.await(1, TimeUnit.SECONDS);

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
        CountDownLatch latch = new CountDownLatch(2);
        when(monitor.onClosedUncleanly(cause)).thenAnswer(new Answer<Long>() {
            public Long answer(InvocationOnMock invocation) {
                latch.countDown();
                return 0L;
            }
        });
        Mockito.doAnswer(new Answer<Void>() {
            public Void answer(InvocationOnMock invocation) {
                latch.countDown();
                return null;
            }
        }).when(monitor).onReopenSucceeded();
        transport.setMonitor(monitor);

        transport.close(cause);
        latch.await(1, TimeUnit.SECONDS);

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
        CountDownLatch latch = new CountDownLatch(3);
        when(monitor.onClosedUncleanly(cause)).thenAnswer(new Answer<Long>() {
            public Long answer(InvocationOnMock invocation) {
                latch.countDown();
                return 1L;
            }
        });
        when(monitor.onReopenFailed(1, 1)).thenAnswer(new Answer<Long>() {
            public Long answer(InvocationOnMock invocation) {
                latch.countDown();
                return 2L;
            }
        });
        Mockito.doAnswer(new Answer<Void>() {
            public Void answer(InvocationOnMock invocation) {
                latch.countDown();
                return null;
            }
        }).when(monitor).onReopenSucceeded();
        transport.setMonitor(monitor);

        transport.close(cause);
        latch.await(1, TimeUnit.SECONDS);

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
        CountDownLatch latch = new CountDownLatch(2);
        when(monitor.onClosedUncleanly(cause)).thenAnswer(new Answer<Long>() {
            public Long answer(InvocationOnMock invocation) {
                latch.countDown();
                return 1L;
            }
        });
        when(monitor.onReopenFailed(1, 1)).thenAnswer(new Answer<Long>() {
            public Long answer(InvocationOnMock invocation) {
                latch.countDown();
                return -1L;
            }
        });
        transport.setMonitor(monitor);

        transport.close(cause);
        latch.await(1, TimeUnit.SECONDS);

        verify(monitor).onClosedUncleanly(cause);
        verify(monitor).onReopenFailed(1, 1);
    }

    private class MockFTransport extends FTransport {

        private int openErrorCount;
        private int errorCount;

        public MockFTransport() {
            super();
        }

        public MockFTransport(int openErrorCount) {
            super();
            this.openErrorCount = openErrorCount;
        }

        @Override
        public boolean isOpen() {
            return false;
        }

        @Override
        public void open() throws TTransportException {
            if (errorCount < openErrorCount) {
                errorCount++;
                throw new TTransportException(FrugalTTransportExceptionType.UNKNOWN, "open error");
            }
        }

        @Override
        public void close() {
            signalClose(null);
        }

        public void close(Exception cause) {
            signalClose(cause);
        }

        public void oneway(FContext context, byte[] payload) throws TTransportException {

        }

        @Override
        public TTransport request(FContext context, byte[] payload) throws TTransportException {
            return null;
        }

    }

}
