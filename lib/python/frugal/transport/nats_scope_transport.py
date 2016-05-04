from io import BytesIO
from threading import Lock

from thrift.transport.TTransport import TTransportException
from tornado import gen

from .scope_transport import FScopeTransport
from .transport_factory import FScopeTransportFactory
from frugal.exceptions import FException


class FNatsScopeTransport(FScopeTransport):

    def __init__(self, conn=None, subject=""):
        """Create a new instance of an FNatsScopeTransport for pub/sub."""
        self._conn = conn
        self._subject = subject
        self._topic_lock = Lock()
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

        self._topic_lock.acquire()
        self._subject = topic

    def unlock_topic(self):
        """Unsets the publish topic and unlocks the transport.

        Throws:
            FException: if the instance is a subscriber
        """
        if self._pull:
            raise FException("Subscriber cannot unlock topic.")

        self._subject = ""
        self._topic_lock.release()

    @gen.coroutine
    def subscribe(self, topic):
        """Opens the Transport to receive messages on the subscription.

        Args:
            topic: string topic to subscribe to
        """
        self._pull = True
        self._subject = topic
        yield self.open()

    @gen.coroutine
    def open(self):
        if not self._conn.is_connected():
            raise TTransportException(TTransportException.NOT_OPEN,
                                      "Nats not connected!")
        if self.is_open():
            raise TTransportException(TTransportException.ALREADY_OPEN,
                                      "Nats is already open!")
        if not self._pull:
            # TODO introduce constant
            self._write_buffer = BytesIO(1024 * 1024)
            self._is_open = True
            return

        if not self._subject:
            raise TTransportException("Subject cannot be empty.")




        raise gen.Return(True)


class FNatsScopeTransportFactory(FScopeTransportFactory):

    def __init__(self, nats_client, queue=None):
        self._nats_client = nats_client
        self._queue = queue

    def get_transport(self):
        return FNatsScopeTransport(self._nats_client, self._queue)
