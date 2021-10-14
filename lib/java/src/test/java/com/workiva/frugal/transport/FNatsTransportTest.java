package com.workiva.frugal.transport;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TTransportExceptionType;
import io.nats.client.Connection;
import io.nats.client.Connection.Status;
import io.nats.client.Dispatcher;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import io.nats.client.Options;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.mockito.ArgumentCaptor;

import java.io.IOException;
import java.util.concurrent.BlockingQueue;

import static com.workiva.frugal.transport.FAsyncTransportTest.mockFrame;
import static java.util.Objects.requireNonNull;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.containsString;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.fail;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.doAnswer;
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
        when(conn.getOptions()).thenReturn(new Options.Builder().build());
        transport = FNatsTransport.of(conn, subject).withInbox(inbox);
    }

    @Test(expected = TTransportException.class)
    public void testOpenNatsDisconnected() throws TTransportException {
        assertFalse(transport.isOpen());
        when(conn.getStatus()).thenReturn(Status.CLOSED);
        transport.open();
    }

    @Test
    public void testOpenCallbackClose() throws TException, IOException, InterruptedException {
        assertFalse(transport.isOpen());
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        ArgumentCaptor<String> inboxCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(handlerCaptor.capture())).thenReturn(mockDispatcher);

        transport.open();

        verify(mockDispatcher).subscribe(inboxCaptor.capture());
        assertEquals(inbox + ".*", inboxCaptor.getValue());

        MessageHandler handler = handlerCaptor.getValue();
        FContext context = new FContext();
        @SuppressWarnings("unchecked")
        BlockingQueue<byte[]> mockQueue = mock(BlockingQueue.class);
        transport.queueMap.put(FAsyncTransport.getOpId(context), mockQueue);

        byte[] mockFrame = mockFrame(context);
        byte[] framedPayload = new byte[mockFrame.length + 4];
        System.arraycopy(mockFrame, 0, framedPayload, 4, mockFrame.length);

        Message mockMessage = mock(Message.class);
        when(mockMessage.getData()).thenReturn(framedPayload);
        handler.onMessage(mockMessage);

        try {
            transport.open();
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.ALREADY_OPEN, e.getType());
        }

        FTransportClosedCallback mockCallback = mock(FTransportClosedCallback.class);
        transport.setClosedCallback(mockCallback);
        transport.close();

        verify(conn).closeDispatcher(mockDispatcher);
        verify(mockCallback).onClose(null);
        verify(mockQueue).put(mockFrame);
    }

    @Test
    public void testFlush() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(any(MessageHandler.class))).thenReturn(mockDispatcher);
        transport.open();

        byte[] buff = "helloworld".getBytes();
        transport.flush(buff);
        verify(conn).publish(subject, null, buff);
    }

    @Test
    public void testFlushOp() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(any(MessageHandler.class))).thenReturn(mockDispatcher);
        transport.open();

        byte[] buff = "helloworld".getBytes();
        transport.flushOp(1234L, buff);
        verify(conn).publish(subject, inbox + ".1234", buff);
    }

    @Test
    public void testFlush_closed() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(any(MessageHandler.class))).thenReturn(mockDispatcher);
        transport.open();

        when(conn.getStatus()).thenReturn(Status.CLOSED);

        byte[] buff = "helloworld".getBytes();
        try {
            transport.flush(buff);
            fail();
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.NOT_OPEN, e.getType());
        }
    }

    @Test
    public void testFlushReconnecting() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(any(MessageHandler.class))).thenReturn(mockDispatcher);
        transport.open();

        when(conn.getStatus()).thenReturn(Status.RECONNECTING);
        try {
            transport.flush("helloworld".getBytes());
            fail();
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.DISCONNECTED, e.getType());
        }
    }

    @Test
    public void testRequestNotOpen() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        try {
            transport.request(new FContext(), "helloworld".getBytes());
            fail();
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.NOT_OPEN, e.getType());
        }
    }

    @Test
    public void testRequestDisconnected() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(any(MessageHandler.class))).thenReturn(mockDispatcher);
        transport.open();

        when(conn.getStatus()).thenReturn(Status.DISCONNECTED);
        try {
            transport.request(new FContext(), "helloworld".getBytes());
            fail();
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.DISCONNECTED, e.getType());
        }
    }

    @Test
    public void testRequestReconnecting() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(any(MessageHandler.class))).thenReturn(mockDispatcher);
        transport.open();

        when(conn.getStatus()).thenReturn(Status.RECONNECTING);
        try {
            transport.request(new FContext(), "helloworld".getBytes());
            fail();
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.DISCONNECTED, e.getType());
        }
    }

    @Test
    public void testRequestTimedOut() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(any(MessageHandler.class))).thenReturn(mockDispatcher);
        transport.open();

        try {
            FContext fContext = new FContext();
            fContext.setTimeout(0);
            transport.request(fContext, "helloworld".getBytes());
            fail();
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.TIMED_OUT, e.getType());
            assertThat(e.getMessage(), containsString("foo"));
        }
    }

    @Test
    public void testRequestServiceNotAvailable() throws TTransportException {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(any(MessageHandler.class))).thenReturn(mockDispatcher);
        transport.open();

        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);
        verify(conn).createDispatcher(handlerCaptor.capture());
        MessageHandler handler = requireNonNull(handlerCaptor.getValue());
        doAnswer(inv -> {
            String replyTo = inv.getArgument(1);

            Message message = mock(Message.class);
            when(message.getSubject()).thenReturn(replyTo);
            when(message.isStatusMessage()).thenReturn(true);
            when(message.getStatus()).thenReturn(new io.nats.client.support.Status(io.nats.client.support.Status.NO_RESPONDERS_CODE, null));
            when(message.getData()).thenReturn(new byte[0]);
            handler.onMessage(message);

            return null;
        }).when(conn).publish(any(), any(), any());

        try {
            FContext fContext = new FContext();
            transport.request(fContext, "helloworld".getBytes());
            fail();
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.SERVICE_NOT_AVAILABLE, e.getType());
            assertThat(e.getMessage(), containsString("foo"));
        }
    }

    @Test
    public void testStatusMessage() throws Exception {
        when(conn.getStatus()).thenReturn(Status.CONNECTED);
        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);
        Dispatcher mockDispatcher = mock(Dispatcher.class);
        when(conn.createDispatcher(handlerCaptor.capture())).thenReturn(mockDispatcher);

        transport.open();

        MessageHandler handler = handlerCaptor.getValue();
        Message message = mock(Message.class);
        when(message.isStatusMessage()).thenReturn(true);
        when(message.getStatus()).thenReturn(new io.nats.client.support.Status(0, null));
        when(message.getData()).thenReturn(new byte[0]);

        // Ensure no exception.
        handler.onMessage(message);
    }
}
