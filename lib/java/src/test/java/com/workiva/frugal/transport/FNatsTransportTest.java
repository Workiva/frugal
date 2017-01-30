package com.workiva.frugal.transport;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.FrugalTTransportExceptionType;
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
import java.util.concurrent.BlockingQueue;

import static com.workiva.frugal.transport.FAsyncTransportTest.mockFrame;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.fail;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
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
        transport = FNatsTransport.of(conn, subject).withInbox(inbox);
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

        transport.open();

        verify(conn).subscribe(inboxCaptor.getValue(), handlerCaptor.getValue());
        assertEquals(inbox, inboxCaptor.getValue());

        MessageHandler handler = handlerCaptor.getValue();
        FContext context = new FContext();
        BlockingQueue<byte[]> mockQueue = mock(BlockingQueue.class);
        transport.queueMap.put(FAsyncTransport.getOpId(context), mockQueue);

        byte[] mockFrame = mockFrame(context);
        byte[] framedPayload = new byte[mockFrame.length + 4];
        System.arraycopy(mockFrame, 0, framedPayload, 4, mockFrame.length);

        Message msg = new Message("foo", "bar", framedPayload);
        handler.onMessage(msg);

        try {
            transport.open();
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(FrugalTTransportExceptionType.ALREADY_OPEN, e.getType());
        }

        FTransportClosedCallback mockCallback = mock(FTransportClosedCallback.class);
        transport.setClosedCallback(mockCallback);
        transport.close();

        verify(sub).unsubscribe();
        verify(mockCallback).onClose(null);
        verify(mockQueue).put(mockFrame);
    }

    @Test
    public void testFlush() throws TTransportException, IOException, InterruptedException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(any(String.class), any(MessageHandler.class))).thenReturn(sub);
        transport.open();

        byte[] buff = "helloworld".getBytes();
        transport.flush(buff);
        verify(conn).publish(subject, inbox, buff);
    }

    @Test(expected = TTransportException.class)
    public void testRequest_notOpen() throws TTransportException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        transport.request(new FContext(), "helloworld".getBytes());
    }
}
