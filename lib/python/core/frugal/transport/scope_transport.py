from thrift.transport.TTransport import TTransportBase


class FScopeTransport(TTransportBase):

    def lock_topic(self, topic):
        """Sets the publish topic and locks the transport for exclusive access.

        Args:
            topic: string pub/sub topic to publish on
        """
        pass

    def unlock_topic(self):
        """Unsets the publish topic and unlocks the transport.
        """
        pass

    def subscribe(self, topic):
        """Opens the transport to receive messages on the subscription.

        Args:
            topic: string pub/sub topic to subscribe to.
        """
        pass

    def unsubscribe(self):
        pass

