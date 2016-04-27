from threading import Lock

from .scope_transport import FScopeTransport
from frugal.exceptions import FException


class FNatsScopeTransport(FScopeTransport):

    def __init__(self, conn=None):
        """Create a new instance of an FNatsScopeTransport for pub/sub."""
        self._conn = conn
        self._subject = ""
        self._lock = Lock()
        self._pull = False

    def lock_topic(self, topic):
        """Sets the publish topic and locks the transport for exclusive access.

        Args:
            topic: string topic name to subscribe to
        Throws:
            FException: if the instance is a subscriber
        """
        if self._pull:
            raise FException("Subscriber cannot lock topic.")

        self._lock.acquire()
        self._subject = topic

    def unlock_topic(self):
        """Unsets the publish topic and unlocks the transport.

        Throws:
            FException: if the instance is a subscriber
        """
        if self._pull:
            raise FException("Subscriber cannot unlock topic.")

        self._subject = ""
        self._lock.release()

    def subscribe(self, topic):
        """Opens the Transport to receive messages on the subscription.

        Args:
            topic: string topic to subscribe to
        """
        self._pull = True
        self._subject = topic
        self.open()

    def open(self):
        # TODO
        pass
