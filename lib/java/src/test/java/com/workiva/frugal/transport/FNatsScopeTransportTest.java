package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FException;
import io.nats.client.*;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;

import java.util.concurrent.ArrayBlockingQueue;

import static org.junit.Assert.*;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

public class FNatsScopeTransportTest {

    private FNatsScopeTransport transport;
    private Connection conn;
    private String topic = "topic";
    private String formattedSubject = TNatsServiceTransport.FRUGAL_PREFIX + topic;
    private AsyncSubscription mockSub;


    @Before
    public void setUp() throws Exception {
        conn = mock(Connection.class);

        transport = new FNatsScopeTransport.Factory(conn).getTransport();
        mockSub = mock(AsyncSubscription.class);
    }

    @Test
    public void testLockTopicSetsSubject() throws Exception {
        transport.lockTopic(topic);

        assertEquals(topic, transport.subject);
    }

    @Test
    public void testLockTopicThrowsExceptionIfPull() throws Exception {
        try {
            transport.pull = true;
            transport.lockTopic(topic);
            fail();
        } catch (FException ex) {
            assertEquals("subscriber cannot lock topic", ex.getMessage());
        }
    }

    @Test
    public void testUnlockTopicClearsSubject() throws Exception {
        transport.lockTopic(topic);

        assertEquals(topic, transport.subject);

        transport.unlockTopic();

        assertEquals("", transport.subject);
    }

    @Test
    public void testUnlockTopicThrowsExceptionIfPull() throws Exception {
        try {
            transport.pull = true;
            transport.unlockTopic();
            fail();
        } catch (FException ex) {
            assertEquals("subscriber cannot unlock topic", ex.getMessage());
        }
    }

    @Test
    public void testSubscribe() throws Exception {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        when(conn.subscribe(any(String.class), any(MessageHandler.class))).thenReturn(mockSub);

        transport.subscribe(topic);

        assertTrue(transport.isOpen());
        assertEquals(mockSub, transport.sub);
    }

    @Test
    public void testSubscribeEmptySubjectThrowsException() throws Exception {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);

        try {
            transport.subscribe("");
            fail();
        } catch(TTransportException ex) {
            assertEquals("Subject cannot be empty.", ex.getMessage());
        }
    }

    @Test
    public void testOpen() throws Exception {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);

        transport.open();

        assertTrue(transport.isOpen());
    }

    @Test
    public void testClosePublisher() throws Exception {
        transport.isOpen = true;

        transport.close();

        // should we be closing nats?
        assertFalse(transport.isOpen);
    }

    @Test
    public void testCloseSubscriber() throws Exception {
        transport.isOpen = true;
        transport.pull = true;
        transport.sub = mockSub;

        // should i have to do this? when does this get initialized
        transport.frameBuffer = new ArrayBlockingQueue<>(4);
        transport.close();

        verify(mockSub).unsubscribe();
        assertFalse(transport.isOpen);

    }
}
