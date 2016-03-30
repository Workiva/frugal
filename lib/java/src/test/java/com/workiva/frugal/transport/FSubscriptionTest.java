package com.workiva.frugal.transport;

import org.junit.Before;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

public class FSubscriptionTest {

    private final String topic = "topic";

    private FScopeTransport mockTransport;
    private FSubscription subscription;

    @Before
    public void setUp() throws Exception {
        mockTransport = mock(FScopeTransport.class);
        subscription = new FSubscription(topic, mockTransport);
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
