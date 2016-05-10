from io import BytesIO
import logging
from threading import Lock
import struct

from thrift.transport.TTransport import TTransportException
from tornado import gen

from .scope_transport import FScopeTransport
from .transport_factory import FScopeTransportFactory
from frugal.exceptions import FException, FMessageSizeException

_MAX_NATS_MESSAGE_SIZE = 1024 * 1024
_FRAME_BUFFER_SIZE = 5
_FRUGAL_PREFIX = "frugal."

logger = logging.getLogger(__name__)


class FNatsScopeTransport(FScopeTransport):

    def __init__(self, nats_client=None, queue=""):
        """Create a new instance of an FNatsScopeTransport for pub/sub."""
        self._nats_client = nats_client
        self._queue = queue
        self._subject = ""
        self._topic_lock = Lock()
        self._pull = False
        self._is_open = False
        self._write_buffer = None

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

    def is_open(self):
        return self._nats_client.is_connected() and self._is_open

    @gen.coroutine
    def open(self):
        """ Asynchronously opens the transport. Throws exception if the provided
        NATS client is not connected or if the transport is already open.

        Throws:
            TTransportException: if NOT_OPEN or ALREADY_OPEN
        """
        if not self._nats_client.is_connected():
            raise TTransportException(TTransportException.NOT_OPEN,
                                      "Nats not connected!")
        if self.is_open():
            raise TTransportException(TTransportException.ALREADY_OPEN,
                                      "Nats is already open!")
        # If _pull is False the transport belongs to a publisher.  Allocate a
        # write buffer, mark open and short circuit
        if not self._pull:
            self._write_buffer = BytesIO()
            self._is_open = True
            return

        if not self._subject:
            raise TTransportException(message="Subject cannot be empty.")

        yield self._nats_client.subscribe(self._subject,
                                   self._queue,
                                   self._on_message_callback)

        self._is_open = True

        raise gen.Return(self)

    def _on_message_callback(self, msg=None):
        if len(msg.data) < 4:
            logger.warning("discarding invalid message frame")

    @gen.coroutine
    def close(self):
        yield self._nats_client.close()

    def read(self):
        pass

    def write(self, buff):
        """Write takes a bytearray and attempts to write it to an underlying
        BytesIO instance.  It will raise an exception if NATS is not connected
        or if writing causes the buffer to exceed 1 MB message size.

        Args:
            buff: bytearray buffer of bytes to write
        Throws:
            TTransportException: if NATS not connected
            FMessageSizeException: if writing to the buffer exceeds 1MB length
        """
        if not self.is_open():
            raise TTransportException(TTransportException.NOT_OPEN,
                                      "Nats not connected!")
        size = len(buff) + len(self._write_buffer.getvalue()) + 4
        if size > _MAX_NATS_MESSAGE_SIZE:
            raise FMessageSizeException("Message exceeds NATS max message size")

        self._write_buffer.write(buff)

    @gen.coroutine
    def flush(self):
        if not self.is_open():
            raise TTransportException(TTransportException.NOT_OPEN,
                                      "Nats not connected!")
        frame = self._write_buffer.getvalue()
        frame_length = struct.pack('!I', len(frame))
        self._write_buffer = BytesIO()
        formatted_subject = self._get_formatted_subject()
        print(formatted_subject)
        yield self._nats_client.publish(formatted_subject,
                                 frame_length + frame)

    def _get_formatted_subject(self):
        return "{}{}".format(_FRUGAL_PREFIX, self._subject)


class FNatsScopeTransportFactory(FScopeTransportFactory):

    def __init__(self, nats_client, queue=None):
        self._nats_client = nats_client
        self._queue = queue

    def get_transport(self):
        return FNatsScopeTransport(self._nats_client, self._queue)
