package com.workiva.frugal.transport;

import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

/**
 * Tests for {@link FSubscription}.
 */
@RunWith(JUnit4.class)
public class FSubscriptionTest {

    private final String topic = "topic";

    private FScopeTransport mockTransport;
    private FSubscription subscription;

    @Before
    public void setUp() throws Exception {
        mockTransport = mock(FScopeTransport.class);
        subscription = FSubscription.of(topic, mockTransport);
    }

    @Test
    public void testGetTopic() throws Exception {
        assertEquals(topic, subscription.getTopic());
    }

    @Test
    public void testUnsubscribeCallsCloseOnTransport() throws Exception {
        subscription.unsubscribe();
        verify(mockTransport).close();
    }
}
