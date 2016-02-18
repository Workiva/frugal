

class FSubscription(object):
    """FSubscription to a pub/sub topic."""

    def __init__(self, topic, transport):
        """Initialize FSubscription.

        Args:
            topic: pub/sub topic string.
            transport: FScopeTransport for the subscription.
        """
        self._topic = topic
        self._transport = transport

    def unsubscribe(self):
        self._transport.close()

    def get_topic(self):
        return self._topic


