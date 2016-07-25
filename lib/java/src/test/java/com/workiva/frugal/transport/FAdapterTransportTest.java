package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FClientRegistry;
import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;
import org.mockito.invocation.InvocationOnMock;
import org.mockito.stubbing.Answer;

import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.anyInt;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FAdapterTransport}.
 */
@RunWith(JUnit4.class)
public class FAdapterTransportTest {

    private FAdapterTransport adapterTransport;
    private TTransport mockTrans;
    private FRegistry registry;

    @Before
    public void setUp() throws Exception {
        mockTrans = mock(TTransport.class);
        adapterTransport = new FAdapterTransport.Factory().getTransport(mockTrans);
    }

    @Test
    public void testIsOpenFalseWhenTransportClosed() throws Exception {
        when(mockTrans.isOpen()).thenReturn(false);

        registry = new FClientRegistry();
        adapterTransport.setRegistry(registry);

        assertFalse(adapterTransport.isOpen());
    }

    @Test
    public void testCloseCleanCloseNotOpen() {
        when(mockTrans.isOpen()).thenReturn(false);

        adapterTransport.close();

        verify(mockTrans, times(0)).close();
    }

    @Test
    public void testCloseCleanClose() throws TTransportException {
        when(mockTrans.isOpen()).thenReturn(true);
        FRegistry mockRegistry = mock(FRegistry.class);
        adapterTransport.setRegistry(mockRegistry);
        adapterTransport.open();

        adapterTransport.close();

        verify(mockTrans).close();
        verify(mockRegistry).close();
    }

    @Test
    public void testCloseUncleanCloseNotOpen() throws TTransportException {
        when(mockTrans.isOpen()).thenReturn(false);
        FRegistry mockRegistry = mock(FRegistry.class);
        adapterTransport.setRegistry(mockRegistry);
        adapterTransport.open();

        adapterTransport.close(new Exception());

        verify(mockTrans).close();
        verify(mockRegistry).close();
    }

    @Test
    public void testCloseUncleanClose() throws TTransportException {
        when(mockTrans.isOpen()).thenReturn(true);
        FRegistry mockRegistry = mock(FRegistry.class);
        adapterTransport.setRegistry(mockRegistry);
        adapterTransport.open();

        adapterTransport.close(new Exception());

        verify(mockTrans).close();
        verify(mockRegistry).close();
    }

    @Test
    public void testWriteFlush() throws TTransportException {
        when(mockTrans.isOpen()).thenReturn(true);
        FRegistry mockRegistry = mock(FRegistry.class);
        adapterTransport.setRegistry(mockRegistry);
        adapterTransport.open();

        // Create writer buffer
        byte[] buff = new byte[1024];
        for (int i = 0; i < 5; i++) {
            buff[i] = (byte) i;
        }
        adapterTransport.write(buff, 0, 5);
        // Verify the wrapping framed transport is buffer the data
        verify(mockTrans, never()).write(any(byte[].class), any(int.class), any(int.class));

        adapterTransport.flush();
        // Verify that flushing writes the frame size then the buffered data
        byte[] frameSize = new byte[]{0, 0, 0, 5};
        verify(mockTrans, times(1)).write(frameSize, 0, 4);
        verify(mockTrans, times(1)).write(buff, 0, 5);
        verify(mockTrans, times(1)).flush();
    }

    @Test
    public void testReadFrame() throws TException {
        when(mockTrans.isOpen()).thenReturn(true);
        FRegistry mockRegistry = mock(FRegistry.class);
        adapterTransport.setRegistry(mockRegistry);

        // Setup close callback
        ArrayBlockingQueue<Throwable> closeQueue = new ArrayBlockingQueue<>(1);
        adapterTransport.setClosedCallback(new FTransportClosedCallback() {
            @Override
            public void onClose(Exception cause) {
                try {
                    closeQueue.put(cause);
                } catch (InterruptedException ignored) {
                }
            }
        });

        // Set frame size
        byte[] buff = new byte[]{1, 2, 3, 4, 5};
        ArrayBlockingQueue<Object[]> argQueue = new ArrayBlockingQueue<>(1);
        when(mockTrans.readAll(any(byte[].class), anyInt(), anyInt())).thenAnswer(new Answer<Integer>() {
            int callNum = 0;

            @Override
            public Integer answer(InvocationOnMock invocation) throws Throwable {
                Object[] args = invocation.getArguments();
                byte[] readBuff = (byte[]) args[0];
                int len = (int) args[2];

                if (callNum == 0) {
                    argQueue.put(args);
                    readBuff[3] = 5;
                } else if (callNum == 1) {
                    argQueue.put(args);
                    System.arraycopy(buff, 0, readBuff, 0, 5);
                } else if (callNum == 2) {
                    try {
                        Thread.sleep(50);
                    } catch (InterruptedException ignored){
                        // This thread should get interrupted on close
                    }
                } else {
                    // Close with EOF
                    throw new TTransportException(TTransportException.END_OF_FILE);
                }
                callNum++;
                return len;
            }
        });
        adapterTransport.open();

        // Verify the first call is to read frame size
        try {
            Object[] args = argQueue.poll(100, TimeUnit.MILLISECONDS);
            assertEquals(0, (int) args[1]);
            assertEquals(4, (int) args[2]);
        } catch (InterruptedException e) {
            assertFalse(true);
        }

        // Verify the first call is to read bytes
        try {
            Object[] args = argQueue.poll(100, TimeUnit.MILLISECONDS);
            assertEquals(0, (int) args[1]);
            assertEquals(5, (int) args[2]);
        } catch (InterruptedException e) {
            assertFalse(true);
        }

        // Verify registry is called
        verify(mockRegistry, times(1)).execute(buff);

        // Wait for EOF
        try {
            TTransportException err = (TTransportException) closeQueue.poll(100, TimeUnit.MILLISECONDS);
            assertEquals(TTransportException.END_OF_FILE,  err.getType());
        } catch (InterruptedException e) {
            assertFalse(true);
        }

        assertFalse(adapterTransport.isOpen());
    }

    @Test(expected = UnsupportedOperationException.class)
    public void testRead() throws TTransportException {
        adapterTransport.read(new byte[0], 0, 0);
    }
}
