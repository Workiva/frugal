// TODO: Remove this with 2.0
package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;
import io.nats.client.AsyncSubscription;
import io.nats.client.Connection;
import io.nats.client.Constants;
import io.nats.client.MessageHandler;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;
import org.mockito.ArgumentCaptor;

import java.io.IOException;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.fail;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link TStatelessNatsTransport}.
 */
@RunWith(JUnit4.class)
@Deprecated
public class TStatelessNatsTransportTest {

    private Connection conn;
    private String subject = "foo";
    private String inbox = "bar";
    private TStatelessNatsTransport transport;

    @Before
    public void setUp() {
        conn = mock(Connection.class);
        transport = new TStatelessNatsTransport(conn, subject, inbox);
    }

    @Test(expected = TTransportException.class)
    public void testOpen_natsDisconnected() throws TTransportException {
        assertFalse(transport.isOpen());
        when(conn.getState()).thenReturn(Constants.ConnState.CLOSED);

        transport.open();
    }

    @Test
    public void testOpenClose() throws TTransportException, IOException, InterruptedException {
        assertFalse(transport.isOpen());
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        ArgumentCaptor<String> inboxCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(inboxCaptor.capture(), handlerCaptor.capture())).thenReturn(sub);

        transport.open();

        verify(conn).subscribe(inboxCaptor.getValue(), handlerCaptor.getValue());
        assertEquals(inbox, inboxCaptor.getValue());

        try {
            transport.open();
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(TTransportException.ALREADY_OPEN, e.getType());
        }

        transport.close();

        verify(sub).unsubscribe();
        assertEquals(FTransport.FRAME_BUFFER_CLOSED, transport.fNatsTransport.frameBuffer.poll(1, TimeUnit.SECONDS));
    }

    @Test(expected = TTransportException.class)
    public void testRead_notOpen() throws TTransportException {
        transport.read(new byte[5], 0, 5);
    }

    @Test
    public void testRead() throws TTransportException, InterruptedException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(any(String.class), any(MessageHandler.class))).thenReturn(sub);
        transport.open();

        transport.fNatsTransport.frameBuffer.put("hello".getBytes());
        transport.fNatsTransport.frameBuffer.put("world".getBytes());

        byte[] buff = new byte[5];
        assertEquals(5, transport.read(buff, 0, 5));
        assertArrayEquals("hello".getBytes(), buff);
        assertEquals(5, transport.read(buff, 0, 5));
        assertArrayEquals("world".getBytes(), buff);

        transport.fNatsTransport.frameBuffer.put(FTransport.FRAME_BUFFER_CLOSED);

        try {
            transport.read(buff, 0, 5);
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(TTransportException.END_OF_FILE, e.getType());
        }
    }

    @Test(expected = FMessageSizeException.class)
    public void testWrite_sizeException() throws TTransportException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(any(String.class), any(MessageHandler.class))).thenReturn(sub);
        transport.open();

        transport.write(new byte[TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE + 1]);
    }

    @Test
    public void testWriteFlush() throws TTransportException, IOException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(any(String.class), any(MessageHandler.class))).thenReturn(sub);
        transport.open();

        byte[] buff = "helloworld".getBytes();
        transport.write(buff);
        transport.flush();

        verify(conn).publish(subject, inbox, buff);
    }

    @Test(expected = TTransportException.class)
    public void testFlush_notOpen() throws TTransportException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        transport.flush();
    }

    @Test
    public void testFlush_noData() throws TTransportException, IOException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(any(String.class), any(MessageHandler.class))).thenReturn(sub);
        transport.open();

        transport.flush();

        verify(conn, times(0)).publish(any(String.class), any(String.class), any(byte[].class));
    }
}
