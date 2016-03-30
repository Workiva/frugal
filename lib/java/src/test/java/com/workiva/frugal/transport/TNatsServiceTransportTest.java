package com.workiva.frugal.transport;


import io.nats.client.Connection;
import io.nats.client.Constants;
import io.nats.client.Message;
import io.nats.client.SyncSubscription;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;

import java.util.concurrent.TimeUnit;

import static org.junit.Assert.assertEquals;

import static org.junit.Assert.assertFalse;

import static org.junit.Assert.fail;
import static org.mockito.Mockito.*;
import static org.mockito.Mockito.when;

public class TNatsServiceTransportTest {

    private final Connection mockConn = mock(Connection.class);
    private final String inbox = "_INBOX";
    private SyncSubscription mockSyncSubscription = mock(SyncSubscription.class);

    private final int maxMissedHeartbeats = 3;
    private final String listenTo = "listenTo";
    private long timeout = 20000L;
    private TNatsServiceTransport client;

    @Before
    public void setUp() throws Exception {
        client = TNatsServiceTransport.client(mockConn, listenTo, timeout, maxMissedHeartbeats);
    }

    @Test
    public void testClientIsOpenFalseWhenNatsDisconnected() throws Exception {
        when(mockConn.getState()).thenReturn(Constants.ConnState.DISCONNECTED);

        assertFalse(client.isOpen());
    }

    @Test
    public void testClientOpenThrowsTransportExceptionIfNatsNotConnected() throws Exception {
        when(mockConn.getState()).thenReturn(Constants.ConnState.DISCONNECTED);

        try {
            client.open();
            fail();
        } catch(TTransportException ex) {
            assertEquals(TTransportException.NOT_OPEN, ex.getType());
            assertEquals("NATS not connected, has status DISCONNECTED", ex.getMessage());
        }
    }

    @Test
    public void testClientOpenThrowsTransportAlreadyOpen() throws Exception {
        when(mockConn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        client.isOpen = true;

        try {
            client.open();
            fail();
        } catch(TTransportException ex) {
            assertEquals(TTransportException.ALREADY_OPEN, ex.getType());
            assertEquals("NATS transport already open", ex.getMessage());
        }
    }

    @Test
    public void testClientThrowsTransportExceptionWhenOpenCalledAlreadyOpen() throws Exception {
        when(mockConn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        when(mockConn.newInbox()).thenReturn(inbox);
        when(mockConn.subscribeSync(TNatsServiceTransport.FRUGAL_PREFIX + inbox, null)).thenReturn(mockSyncSubscription);

        Message result = mock(Message.class);
        when(result.getReplyTo()).thenReturn("replyTo");
        when(result.getSubject()).thenReturn("subject");

        when(mockSyncSubscription.nextMessage(timeout, TimeUnit.MILLISECONDS)).thenReturn(result);
        when(result.getData()).thenReturn("heartbeatListen heartbeatReply 20000".getBytes());

        client.open();

        verify(mockSyncSubscription).autoUnsubscribe(1);

        assertEquals("heartbeatListen", client.heartbeatListen);
        assertEquals("heartbeatReply", client.heartbeatReply);
        assertEquals(20000, client.heartbeatInterval);
        assertEquals("subject", client.listenTo);
        assertEquals("replyTo", client.writeTo);

    }

}
