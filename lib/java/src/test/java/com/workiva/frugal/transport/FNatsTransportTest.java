package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;
import com.workiva.frugal.protocol.FRegistry;
import io.nats.client.AsyncSubscription;
import io.nats.client.Connection;
import io.nats.client.Constants;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.mockito.ArgumentCaptor;

import java.io.IOException;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.fail;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FNatsTransport}.
 */
public class FNatsTransportTest {

    private Connection conn;
    private String subject = "foo";
    private String inbox = "bar";
    private FNatsTransport transport;

    @Before
    public void setUp() {
        conn = mock(Connection.class);
        transport = new FNatsTransport(conn, subject, inbox);
    }

    @Test(expected = TTransportException.class)
    public void testOpen_natsDisconnected() throws TTransportException {
        assertFalse(transport.isOpen());
        when(conn.getState()).thenReturn(Constants.ConnState.CLOSED);

        transport.open();
    }

    @Test
    public void testOpenCallbackClose() throws TException, IOException, InterruptedException {
        assertFalse(transport.isOpen());
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        ArgumentCaptor<String> inboxCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(inboxCaptor.capture(), handlerCaptor.capture())).thenReturn(sub);

        FRegistry mockRegistry = mock(FRegistry.class);
        transport.setRegistry(mockRegistry);
        transport.open();

        verify(conn).subscribe(inboxCaptor.getValue(), handlerCaptor.getValue());
        assertEquals(inbox, inboxCaptor.getValue());

        MessageHandler handler = handlerCaptor.getValue();
        byte[] framedPayload = new byte[]{0, 1, 2, 3, 4, 5, 6, 7};
        byte[] payload = new byte[]{4, 5, 6, 7};
        Message msg = new Message("foo", "bar", framedPayload);
        handler.onMessage(msg);
        verify(mockRegistry).execute(payload);

        try {
            transport.open();
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(TTransportException.ALREADY_OPEN, e.getType());
        }

        FTransportClosedCallback mockCallback = mock(FTransportClosedCallback.class);
        transport.setClosedCallback(mockCallback);
        transport.close();

        verify(sub).unsubscribe();
        verify(mockCallback).onClose(null);
        verify(mockRegistry).close();
    }

    @Test(expected = UnsupportedOperationException.class)
    public void testRead() throws TTransportException {
        transport.read(new byte[0], 0, 0);
    }

    @Test(expected = TTransportException.class)
    public void testWrite_notOpen() throws TTransportException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        transport.write(new byte[5]);
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

        byte[] expected = new byte[4 + buff.length];
        expected[3] = (byte) buff.length;
        System.arraycopy(buff, 0, expected, 4, buff.length);

        verify(conn).publish(subject, inbox, expected);
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
