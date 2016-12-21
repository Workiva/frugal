import logging
import gevent

from gnats.client.errors import ErrConnectionClosed, ErrTimeout, ErrMaxPayload

from thrift.transport.TTransport import TTransportException, TMemoryBuffer

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.exceptions import FMessageSizeException
from frugal.transport import FPublisherTransportFactory
from frugal.transport import FPublisherTransport
from frugal.transport import FSubscriberTransportFactory
from frugal.transport import FSubscriberTransport

_FRAME_BUFFER_SIZE = 5
_FRUGAL_PREFIX = "frugal."

logger = logging.getLogger(__name__)


class FNatsPublisherTransportFactory(FPublisherTransportFactory):
    def __init__(self, nats_client):
        self._nats_client = nats_client

    def get_transport(self):
        return FNatsPublisherTranpsort(self._nats_client)


class FNatsPublisherTranpsort(FPublisherTransport):

    def __init__(self, nats_client):
        super(FNatsPublisherTranpsort, self).__init__(_NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client

    def open(self):
        if not self._nats_client.is_connected:
            msg = "Nats not connected!"
            raise TTransportException(TTransportException.NOT_OPEN, msg)

    def close(self):
        if not self.is_open():
            return

        self._nats_client.flush()

    def is_open(self):
        return self._nats_client.is_connected

    def publish(self, topic, data):
        if not self.is_open():
            msg = 'Nats not connected!'
            raise TTransportException(TTransportException.NOT_OPEN, msg)
        if self._check_publish_size(data):
            msg = 'Message exceeds NATS max message size'
            raise FMessageSizeException.request(msg)
        try:
            self._nats_client.publish('frugal.{0}'.format(topic), data)
        except (ErrMaxPayload, ErrConnectionClosed) as e:
            raise TTransportException(
                TTransportException.UNKNOWN,
                'Error publishing to nats: {e}'.format(e=e))



class FNatsSubscriberTransportFactory(FSubscriberTransportFactory):
    def __init__(self, nats_client, queue=''):
        self._nats_client = nats_client
        self._queue = queue

    def get_transport(self):
        return FNatsSubscriberTransport(self._nats_client, self._queue)


class FNatsSubscriberTransport(FSubscriberTransport):

    def __init__(self, nats_client, queue):
        self._nats_client = nats_client
        self._queue = queue
        self._is_subscribed = False
        self._sub = None

    def subscribe(self, topic, callback):
        if not self._nats_client.is_connected:
            msg = "Nats not connected!"
            raise TTransportException(TTransportException.NOT_OPEN, msg)

        if self.is_subscribed():
            msg = 'Already subscribed to nats topic!'
            raise TTransportException(TTransportException.ALREADY_OPEN, msg)

        self._sub = self._nats_client.subscribe(
            'frugal.{0}'.format(topic),
            queue=self._queue,
            cb=lambda message: callback(TMemoryBuffer(message.data[4:]))
        )
        try:
            self._nats_client.flush()
        except (ErrConnectionClosed, ErrTimeout) as e:
            raise TTransportException(
                TTransportException.UNKNOWN,
                'Error flushing nats during subscribe: {e}'.format(e=e))
        self._is_subscribed = True


    def unsubscribe(self):
        if not self.is_subscribed():
            return
        self._nats_client.unsubscribe(self._sub.id)
        self._sub = None
        self._is_subscribed = False

    def is_subscribed(self):
        return self._is_subscribed and self._nats_client.is_connected
