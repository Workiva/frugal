package com.workiva.frugal.transport;

/**
 * FSubscription is a subscription to a pub/sub topic created by a scope. The
 * topic subscription is actually handled by an FSubscriberTransport, which the
 * FSubscription wraps. Each FSubscription should have its own FSubscriberTransport.
 * The FSubscription is used to unsubscribe from the topic.
 */
public final class FSubscription {

    private final String topic;
    private final FSubscriberTransport transport;

    private FSubscription(String topic, FSubscriberTransport transport) {
        this.topic = topic;
        this.transport = transport;
    }

    /**
     * Construct a new subscription. This is used only by generated
     * code and should not be called directly.
     *
     * @param topic     for the subscription.
     * @param transport for the subscription.
     *
     * @return FSubscription
     */
    public static FSubscription of(String topic, FSubscriberTransport transport) {
        return new FSubscription(topic, transport);
    }

    /**
     * Queries whether the subscription is active.
     *
     * @return True if the subscription is active.
     */
    boolean isSubscribed() {
        return transport != null && transport.isSubscribed();
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
        if (transport != null) {
            transport.unsubscribe();
        }
    }
}
