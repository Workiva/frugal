from io import BytesIO
import logging
from threading import Lock
import struct

from thrift.transport.TTransport import TTransportException, TMemoryBuffer
from tornado import gen

from frugal.transport import FScopeTransport, FScopeTransportFactory
from frugal.exceptions import FException, FMessageSizeException

_NATS_MAX_MESSAGE_SIZE = 1024 * 1024
_FRAME_BUFFER_SIZE = 5
_FRUGAL_PREFIX = "frugal."

logger = logging.getLogger(__name__)


class FNatsScopeTransport(FScopeTransport):

    def __init__(self, nats_client=None, queue=b''):
        """Create a new instance of an FNatsScopeTransport for pub/sub.

            Args:
                nats_client: A connected instance of the Python NATS client.
        """
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
            ex = FException("Subscriber cannot lock topic.")
            logger.exception(ex)
            raise ex

        self._topic_lock.acquire()
        self._subject = topic

    def unlock_topic(self):
        """Unsets the publish topic and unlocks the transport.

        Throws:
            FException: if the instance is a subscriber
        """
        if self._pull:
            ex = FException("Subscriber cannot unlock topic.")
            logger.exception(ex)
            raise ex

        self._subject = ""
        self._topic_lock.release()

    @gen.coroutine
    def subscribe(self, topic, callback=None):
        """Opens the Transport to receive messages on the subscription.

        Args:
            topic: string topic to subscribe to
        """
        self._pull = True
        self._subject = topic
        yield self.open(callback)

    def is_open(self):
        return self._nats_client.is_connected() and self._is_open

    @gen.coroutine
    def open(self, callback=None):
        """ Asynchronously opens the transport. Throws exception if the provided
        NATS client is not connected or if the transport is already open.

        Args:
            callback: function to call when Subscriber receives a message
        Throws:
            TTransportException: if NOT_OPEN or ALREADY_OPEN
        """
        if not self._nats_client.is_connected():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Nats not connected!")
            logger.exception(ex)
            raise ex

        if self.is_open():
            ex = TTransportException(TTransportException.ALREADY_OPEN,
                                     "Nats is already open!")
            logger.exception(ex)
            raise ex

        # If _pull is False the transport belongs to a publisher.  Allocate a
        # write buffer, mark open and short circuit
        if not self._pull:
            self._write_buffer = BytesIO()
            self._is_open = True
            return

        if not self._subject:
            ex = TTransportException(message="Subject cannot be empty.")
            logger.exception(ex)
            raise ex

        def on_message(msg):
            callback(TMemoryBuffer(msg.data[4:]))

        self._sub_id = yield self._nats_client.subscribe(
            "frugal.{}".format(self._subject),
            self._queue,
            on_message
        )

        self._is_open = True
        logger.debug("FNatsScopeTransport open.")

    @gen.coroutine
    def close(self):
        logger.debug("Closing FNatsScopeTransport.")

        if not self._is_open:
            return

        if not self._pull:
            self._is_open = False
            return

        # Unsubscribe
        self._nats_client.auto_unsubscribe(self._sub_id, "")
        self._sub_id = None

        self._is_open = False

    def read(self, sz):
        # Don't call this
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
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Nats not connected!")
            logger.exception(ex)
            raise ex

        wbuf_length = len(self._write_buffer.getvalue())

        size = len(buff) + wbuf_length

        if size > _NATS_MAX_MESSAGE_SIZE:
            ex = FMessageSizeException("Message exceeds NATS max message size")
            logger.exception(ex)
            raise ex

        self._write_buffer.write(buff)

    @gen.coroutine
    def flush(self):
        if not self.is_open():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Nats not connected!")
            logger.exception(ex)
            raise ex

        frame = self._write_buffer.getvalue()
        frame_length = struct.pack('!I', len(frame))
        self._write_buffer = BytesIO()

        formatted_subject = self._get_formatted_subject()

        yield self._nats_client.publish(formatted_subject, frame_length + frame)

    def _get_formatted_subject(self):
        return "{}{}".format(_FRUGAL_PREFIX, self._subject)


class FNatsScopeTransportFactory(FScopeTransportFactory):

    def __init__(self, nats_client, queue=b''):
        self._nats_client = nats_client
        self._queue = queue

    def get_transport(self):
        return FNatsScopeTransport(self._nats_client, self._queue)
