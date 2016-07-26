package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;

import java.util.concurrent.ExecutorService;

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
            assertEquals(TTransportException.ALREADY_OPEN, e.getType());
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
        FRegistry mockRegistry = mock(FRegistry.class);
        tr.setRegistry(mockRegistry);
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        when(mockTr.isOpen()).thenReturn(true);

        tr.open();

        verify(mockExecutorFactory).newExecutor();
        verify(mockExecutor).submit(any(Runnable.class));

        assertTrue(tr.isOpen());

        tr.close();

        verify(mockRegistry).close();
        verify(mockTr).close();
        verify(mockExecutor).shutdownNow();

        assertFalse(tr.isOpen());
    }

    /**
     * Ensures TransportReader reads frames from the transport and executes them on the registry. When it encounters an
     * EOF it closes the transport and returns.
     */
    @Test
    public void testTransportReader() throws TException {
        FRegistry mockRegistry = mock(FRegistry.class);
        tr.setRegistry(mockRegistry);
        FAdapterTransport.ExecutorFactory mockExecutorFactory = mock(FAdapterTransport.ExecutorFactory.class);
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        tr.setExecutorFactory(mockExecutorFactory);
        tr.open();
        when(mockTr.isOpen()).thenReturn(true);
        when(mockTr.readAll(any(byte[].class), any(int.class), any(int.class)))
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    buff[3] = 10;
                    return 4;
                }) // Read frame 1 size
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    for (int i = 0; i < 10; i++) {
                        buff[i] = 1;
                    }
                    return 10;
                }) // Read frame 1
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    buff[3] = 5;
                    return 4;
                }) // Read frame 2 size
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    for (int i = 0; i < 5; i++) {
                        buff[i] = 2;
                    }
                    return 5;
                }) // Read frame 2
                .thenThrow(new TTransportException(TTransportException.END_OF_FILE));
        Runnable reader = tr.newTransportReader();

        reader.run();

        assertFalse(tr.isOpen());
        verify(mockRegistry).execute(new byte[]{1, 1, 1, 1, 1, 1, 1, 1, 1, 1});
        verify(mockRegistry).execute(new byte[]{2, 2, 2, 2, 2});
    }

    /**
     * Ensures TransportReader reads frames from the transport and closes it when
     * an error is encountered.
     */
    @Test
    public void testTransportReader_error() throws TException {
        FRegistry mockRegistry = mock(FRegistry.class);
        tr.setRegistry(mockRegistry);
        FAdapterTransport.ExecutorFactory mockExecutorFactory = mock(FAdapterTransport.ExecutorFactory.class);
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        tr.setExecutorFactory(mockExecutorFactory);
        tr.open();
        when(mockTr.isOpen()).thenReturn(true);
        TTransportException cause = new TTransportException(TTransportException.UNKNOWN, "error");
        when(mockTr.readAll(any(byte[].class), any(int.class), any(int.class))).thenThrow(cause);
        Runnable reader = tr.newTransportReader();

        reader.run();

        assertFalse(tr.isOpen());
    }

    /**
     * Ensures write buffers up write and flush calls through to the underlying transport.
     */
    @Test
    public void testWriteFlush() throws TTransportException {
        byte[] buff = new byte[]{1, 2, 3, 4};
        int off = 0;
        int len = 4;

        when(mockTr.isOpen()).thenReturn(true);
        tr.open();

        tr.write(buff, off, len);
        tr.flush();

        verify(mockTr).write(new byte[]{0, 0, 0, 4}, 0, 4);
        byte[] expected = new byte[1024];
        expected[0] = 1;
        expected[1] = 2;
        expected[2] = 3;
        expected[3] = 4;
        verify(mockTr).write(expected, 0, 4);
        verify(mockTr).flush();
    }

    /**
     * Ensures write throws TTransportException if the transport is not open.
     */
    @Test(expected = TTransportException.class)
    public void testWrite_notOpen() throws TTransportException {
        tr.write(new byte[1]);
    }


    /**
     * Ensures flush throws TTransportException if the transport is not open.
     */
    @Test(expected = TTransportException.class)
    public void testFlush_notOpen() throws TTransportException {
        tr.flush();
    }

}
