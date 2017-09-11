package com.workiva.frugal.transport;

import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FSubscription}.
 */
@RunWith(JUnit4.class)
public class FSubscriptionTest {

    private final String topic = "topic";

    private FSubscriberTransport mockTransport;
    private FSubscription subscription;

    @Before
    public void setUp() throws Exception {
        mockTransport = mock(FSubscriberTransport.class);
        subscription = FSubscription.of(topic, mockTransport);
    }

    @Test
    public void testIsSubscribed() throws Exception {
        when(mockTransport.isSubscribed()).thenReturn(true);
        assertTrue(subscription.isSubscribed());

        when(mockTransport.isSubscribed()).thenReturn(false);
        assertFalse(subscription.isSubscribed());
    }

    @Test
    public void testIsSubscribedNoTransport() throws Exception {
        FSubscription sub = FSubscription.of(topic, null);
        assertFalse(sub.isSubscribed());
    }

    @Test
    public void testGetTopic() throws Exception {
        assertEquals(topic, subscription.getTopic());
    }

    @Test
    public void testUnsubscribeCallsCloseOnTransport() throws Exception {
        subscription.unsubscribe();
        verify(mockTransport).unsubscribe();
    }
}
