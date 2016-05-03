package com.workiva.frugal.transport;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;

/**
 * FSubscription is a subscription to a pub/sub topic created by a scope. The
 * topic subscription is actually handled by an FScopeTransport, which the
 * FSubscription wraps. Each FSubscription should have its own FScopeTransport.
 * The FSubscription is used to unsubscribe from the topic.
 */
public class FSubscription {

    private String topic;
    private FScopeTransport transport;
    private BlockingQueue<Exception> onError;

    /**
     * Construct a new subscription. This is used only by generated
     * code and should not be called directly.
     *
     * @param topic     for the subscription.
     * @param transport for the subscription.
     */
    public FSubscription(String topic, FScopeTransport transport) {
        this.topic = topic;
        this.transport = transport;
        this.onError = new LinkedBlockingQueue<>(1);
    }

    /**
     * Get the subscription topic.
     *
     * @return subscription topic.
     */
    public String getTopic() {
        return topic;
    }

    /**
     * Unsubscribe from the topic.
     */
    public void unsubscribe() {
        transport.close();
    }

    /**
     * Queue which returns any error might have occured during
     * the subscription's lifetime. An error indicates that this
     * subscription has been closed.
     *
     * @return The Exception queue.
     * @deprecated TODO 2.0.0: remove in a future release.
     */
    @Deprecated
    public BlockingQueue<Exception> onError() {
        return onError;
    }

}
