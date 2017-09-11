package com.workiva.frugal.transport;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TApplicationExceptionType;
import com.workiva.frugal.exception.TTransportExceptionType;
import com.workiva.frugal.util.ProtocolUtils;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.mockito.Mockito;

import java.io.UnsupportedEncodingException;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;

import static com.workiva.frugal.transport.FAsyncTransportTest.mockFrame;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for FAdapterTransport.
 */
public class FAdapterTransportTest {

    private TTransport mockTr;
    private FAdapterTransport tr;

    @Before
    public void setUp() {
        mockTr = mock(TTransport.class);
        tr = new FAdapterTransport(mockTr);
    }

    /**
     * Ensures open throws a TTransportException if the wrapped transport fails to open.
     */
    @Test(expected = TTransportException.class)
    public void testOpen_error() throws TTransportException {
        doThrow(TTransportException.class).when(mockTr).open();

        tr.open();

        verify(mockTr).open();
    }

    /**
     * Ensures open throws an ALREADY_OPEN TTransportException if the transport is already open.
     */
    @Test
    public void testOpen_alreadyOpen() throws TTransportException {
        when(mockTr.isOpen()).thenReturn(true);

        tr.open();

        try {
            tr.open();
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.ALREADY_OPEN, e.getType());
        }

        verify(mockTr).open();
        verify(mockTr).isOpen();
    }

    /**
     * Ensures open opens the underlying transport and starts the read thread and close shuts down the executor.
     */
    @Test
    public void testOpenClose() throws TTransportException {
        FAdapterTransport.ExecutorFactory mockExecutorFactory = mock(FAdapterTransport.ExecutorFactory.class);
        tr.setExecutorFactory(mockExecutorFactory);
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        when(mockTr.isOpen()).thenReturn(true);

        tr.open();

        verify(mockExecutorFactory).newExecutor();
        verify(mockExecutor).submit(any(Runnable.class));

        assertTrue(tr.isOpen());

        tr.close();

        verify(mockTr).close();
        verify(mockExecutor).shutdownNow();

        assertFalse(tr.isOpen());
    }

    /**
     * Ensures TransportReader reads frames from the transport and passes them to handleResponse. When it encounters
     * an EOF it closes the transport and returns.
     */
    @Test
    public void testTransportReader() throws TException, InterruptedException, UnsupportedEncodingException {
        FContext context1 = new FContext();
        BlockingQueue<byte[]> mockQueue1 = mock(BlockingQueue.class);
        byte[] mockFrame1 = mockFrame(context1);
        FContext context2 = new FContext();
        BlockingQueue<byte[]> mockQueue2 = mock(BlockingQueue.class);
        byte[] mockFrame2 = mockFrame(context2);

        tr.queueMap.put(FAsyncTransport.getOpId(context1), mockQueue1);
        tr.queueMap.put(FAsyncTransport.getOpId(context2), mockQueue2);

        FAdapterTransport.ExecutorFactory mockExecutorFactory = mock(FAdapterTransport.ExecutorFactory.class);
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        tr.setExecutorFactory(mockExecutorFactory);
        tr.open();
        when(mockTr.isOpen()).thenReturn(true);
        when(mockTr.readAll(any(byte[].class), any(int.class), any(int.class)))
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    ProtocolUtils.writeInt(mockFrame1.length, buff, 0);
                    return 4;
                }) // Read frame 1 size
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    System.arraycopy(mockFrame1, 0, buff, 0, mockFrame1.length);
                    return mockFrame1.length;
                }) // Read frame 1
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    ProtocolUtils.writeInt(mockFrame2.length, buff, 0);
                    return 4;
                }) // Read frame 2 size
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    System.arraycopy(mockFrame2, 0, buff, 0, mockFrame1.length);
                    return 5;
                }) // Read frame 2
                .thenThrow(new TTransportException(TTransportExceptionType.END_OF_FILE));
        Runnable reader = tr.newTransportReader();

        reader.run();

        assertFalse(tr.isOpen());
        verify(mockQueue1).put(mockFrame1);
        verify(mockQueue2).put(mockFrame2);
    }

    /**
     * Ensures TransportReader reads frames from the transport and closes it when
     * an error is encountered.
     */
    @Test
    public void testTransportReader_error() throws TException {
        FAdapterTransport.ExecutorFactory mockExecutorFactory = mock(FAdapterTransport.ExecutorFactory.class);
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        tr.setExecutorFactory(mockExecutorFactory);
        tr.open();
        when(mockTr.isOpen()).thenReturn(true);
        TTransportException cause = new TTransportException(TApplicationExceptionType.UNKNOWN, "error");
        when(mockTr.readAll(any(byte[].class), any(int.class), any(int.class))).thenThrow(cause);
        Runnable reader = tr.newTransportReader();

        reader.run();

        assertFalse(tr.isOpen());
    }

    /**
     * Ensures flush calls through to write and flush the underlying transport.
     */
    @Test
    public void testFlush() throws TTransportException {
        when(mockTr.isOpen()).thenReturn(true);
        Mockito.doNothing().when(mockTr).open();
        tr.open();

        byte[] buff = "helloworld".getBytes();
        tr.flush(buff);

        verify(mockTr).write(buff);
        verify(mockTr).flush();
    }
}
