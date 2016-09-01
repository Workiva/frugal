package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FException;
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
import org.mockito.Matchers;

import java.util.concurrent.ArrayBlockingQueue;

import static com.workiva.frugal.transport.FNatsTransport.FRUGAL_PREFIX;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FNatsScopeTransport}.
 */
@RunWith(JUnit4.class)
public class FNatsScopeTransportTest {

    private FNatsScopeTransport transport;
    private Connection conn;
    private String topic = "topic";
    private String formattedSubject = FRUGAL_PREFIX + topic;
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
        ArgumentCaptor<String> topicCaptor = ArgumentCaptor.forClass(String.class);

        when(conn.subscribe(topicCaptor.capture(), (String) Matchers.isNull(), any(MessageHandler.class)))
                .thenReturn(mockSub);

        transport.subscribe(topic);

        assertTrue(transport.isOpen());
        assertEquals(mockSub, transport.sub);

        assertEquals(formattedSubject, topicCaptor.getValue());
    }

    @Test
    public void testSubscribeQueue() throws Exception {
        transport = new FNatsScopeTransport.Factory(conn, "foo").getTransport();
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        ArgumentCaptor<String> topicCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<String> queueCaptor = ArgumentCaptor.forClass(String.class);

        when(conn.subscribe(topicCaptor.capture(), queueCaptor.capture(), any(MessageHandler.class)))
                .thenReturn(mockSub);

        transport.subscribe(topic);

        assertTrue(transport.isOpen());
        assertEquals(mockSub, transport.sub);
        assertEquals("foo", queueCaptor.getValue());
        assertEquals(formattedSubject, topicCaptor.getValue());
    }

    @Test
    public void testSubscribeEmptySubjectThrowsException() throws Exception {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);

        try {
            transport.subscribe("");
            fail();
        } catch (TTransportException ex) {
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

        assertFalse(transport.isOpen);
    }

    @Test
    public void testCloseSubscriber() throws Exception {
        transport.isOpen = true;
        transport.pull = true;
        transport.sub = mockSub;

        transport.frameBuffer = new ArrayBlockingQueue<>(4);
        transport.close();

        verify(mockSub).unsubscribe();
        assertFalse(transport.isOpen);

    }
}
