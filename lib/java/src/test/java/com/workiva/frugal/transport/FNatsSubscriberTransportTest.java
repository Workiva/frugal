package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FAsyncCallback;
import io.nats.client.AsyncSubscription;
import io.nats.client.Connection;
import io.nats.client.Constants;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.mockito.ArgumentCaptor;
import org.mockito.Mockito;

import java.io.IOException;

import static com.workiva.frugal.transport.FNatsTransport.FRUGAL_PREFIX;
import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNull;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FSubscriberTransport}.
 */
public class FNatsSubscriberTransportTest {

    private FNatsSubscriberTransport transport;
    private Connection conn;
    private String topic = "topic";
    private String formattedSubject = FRUGAL_PREFIX + topic;
    private AsyncSubscription mockSub;

    private class Handler implements FAsyncCallback {
        TTransport transport;
        TException exception;
        @Override
        public void onMessage(TTransport transport) throws TException {
            this.transport = transport;
            if (exception != null) {
                throw exception;
            }
        }
    }

    @Before
    public void setUp() throws Exception {
        conn = mock(Connection.class);
        transport = new FNatsSubscriberTransport.Factory(conn).getTransport();
        mockSub = mock(AsyncSubscription.class);
    }

    @Test
    public void testSubscribe() throws Exception {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        ArgumentCaptor<String> topicCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);

        when(conn.subscribe(topicCaptor.capture(), isNull(), handlerCaptor.capture()))
                .thenReturn(mockSub);

        Handler handler = new Handler();
        transport.subscribe(topic, handler);

        // Nats subscription not yet valid
        when(mockSub.isValid()).thenReturn(false);
        assertFalse(transport.isSubscribed());

        // All good now
        when(mockSub.isValid()).thenReturn(true);
        assertTrue(transport.isSubscribed());
        assertEquals(mockSub, transport.sub);

        assertEquals(formattedSubject, topicCaptor.getValue());

        // Handle a good frame
        byte[] frame = new byte[]{0, 0, 0, 4, 1, 2, 3, 4};
        MessageHandler messageHandler = handlerCaptor.getValue();
        messageHandler.onMessage(new Message("foo", null, frame));

        byte[] expectedPayload = new byte[]{1, 2, 3, 4};
        byte[] actualPayload = new byte[4];

        handler.transport.read(actualPayload, 0, 4);
        assertArrayEquals(expectedPayload, actualPayload);

        // Handle a bad frame
        handler.transport = null;
        messageHandler.onMessage(new Message("foo", null, new byte[3]));
        assertNull(handler.transport);

        // Handler an FAsyncCallback error
        handler.exception = new TException("Bad things!");
        messageHandler.onMessage(new Message("foo", null, frame));
        actualPayload = new byte[4];

        handler.transport.read(actualPayload, 0, 4);
        assertArrayEquals(expectedPayload, actualPayload);
    }

    @Test
    public void testSubscribeQueue() throws Exception {
        transport = new FNatsSubscriberTransport.Factory(conn, "foo").getTransport();
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        ArgumentCaptor<String> topicCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<String> queueCaptor = ArgumentCaptor.forClass(String.class);

        when(conn.subscribe(topicCaptor.capture(), queueCaptor.capture(), any(MessageHandler.class)))
                .thenReturn(mockSub);

        Handler handler = new Handler();
        transport.subscribe(topic, handler);
        when(mockSub.isValid()).thenReturn(true);

        assertTrue(transport.isSubscribed());
        assertEquals(mockSub, transport.sub);
        assertEquals("foo", queueCaptor.getValue());
        assertEquals(formattedSubject, topicCaptor.getValue());
    }

    @Test
    public void testSubscribeEmptySubjectThrowsException() throws Exception {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        try {
            transport.subscribe("", new Handler());
            fail();
        } catch (TTransportException ex) {
            assertEquals("Subject cannot be empty.", ex.getMessage());
        }
    }

    @Test(expected = TTransportException.class)
    public void testSubscribeNotConnectedThrowsException() throws Exception {
        when(conn.getState()).thenReturn(Constants.ConnState.DISCONNECTED);
        transport.subscribe("", new Handler());
    }

    @Test
    public void testCloseSubscriber() throws Exception {
        transport.sub = mockSub;
        transport.unsubscribe();
        verify(mockSub).unsubscribe();
        assertFalse(transport.isSubscribed());
        // Make sure unsubscribe doesn't throw an error when called again
        transport.unsubscribe();
    }

    @Test
    public void testCloseSubscriberUnsubscribeException() throws Exception {
        transport.sub = mockSub;
        Mockito.doThrow(new IOException("Problem")).when(mockSub).unsubscribe();
        transport.unsubscribe();
        verify(mockSub).unsubscribe();
        assertFalse(transport.isSubscribed());
    }
}

