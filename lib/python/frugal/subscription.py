

class FSubscription(object):
    """FSubscription to a pub/sub topic.  This is used only by generated code
    and should not be called directly."""

    def __init__(self, topic, transport):
        """Initialize FSubscription.

        Args:
            topic: pub/sub topic string.
            transport: FScopeTransport for the subscription.
        """
        self._topic = topic
        self._transport = transport

    def get_topic(self):
        """Return subscription topic."""
        return self._topic

    def unsubscribe(self):
        """Unsubscribe from the topic."""
        self._transport.close()


