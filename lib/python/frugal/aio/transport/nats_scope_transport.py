from asyncio.locks import Lock

from nats.aio.client import Client
from thrift.transport.TTransport import TTransportException
from thrift.transport.TTransport import TMemoryBuffer

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.aio.transport import FTransportBase
from frugal.exceptions import FException
from frugal.transport import FScopeTransport
from frugal.transport import FScopeTransportFactory


class FNatsScopeTransportFactory(FScopeTransportFactory):
    """FNatsScopeTransportFactory produces FScopeTransports."""
    def __init__(
            self,
            nats_client: Client,
            queue=''
    ):
        """
        Creates a new instance of FNatsScopeTransportFactory for producing new
        pub/sub transports.

        Args:
            nats_client: A connected nats client.
            queue: The queue group to subscribe to. If multiple transport
                   are part of the same queue group, only one transport will
                   receive a message sent to that queue group. The empty string
                   indicates NOT to join a queue group.
        """
        self._nats_client = nats_client
        self._queue = queue

    def get_transport(self):
        return FNatsScopeTransport(
                self._nats_client,
                queue=self._queue
        )


class FNatsScopeTransport(FScopeTransport, FTransportBase):
    """
    FNatsScopeTransport implements FScopeTransport using nats as the underlying
    message broker. Messages are limited to 1MB in size.
    """
    def __init__(
            self,
            nats_client: Client,
            queue=''
    ):
        """
        Creates a new instance of FNatsScopeTransport for pub/sub.

        Args:
            nats_client: A connected nats client.
            queue: The queue group to subscribe to. If multiple transports
                   are part of the same queue group, only one transport will
                   receive a message sent to that queue group. The empty string
                   indicates NOT to join a queue group.
        """
        FTransportBase.__init__(self, _NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client
        self._queue = queue
        self._subject = ''
        self._topic_lock = Lock()
        self._pull = False
        self._is_open = False
        self._sub_id = None
        self._callback = None

    async def lock_topic(self, topic: str):
        """
        Sets the publish topic and locks the transport for exclusive access.
        This should not be used if the transport instance is a subscriber.

        Args:
            topic: The topic name to subscribe to.
        """
        if self._pull:
            raise FException('Subscribers cannot lock topics')

        await self._topic_lock.acquire()
        self._subject = topic

    def unlock_topic(self):
        """
        Unsets the publish topic and unlocks the transport.
        This should not be used if the transport instance is a subscriber.
        """
        if self._pull:
            raise FException('Subscribers cannot lock topics')

        self._subject = ''
        self._topic_lock.release()

    async def subscribe(self, topic: str, callback):
        """
        Opens the transport to receive messages on the subscription.

        Args:
             topic: The topic to subscribe to.
             callback: The function to call when a message is received.
        """
        self._pull = True
        self._subject = topic
        await self.open(callback=callback)

    def isOpen(self):
        """True if the transport is open, False otherwise"""
        return self._nats_client.is_connected and self._is_open

    async def open(self, callback=None):
        """
        Opens the transport. An exception will be thrown if the nats client
        is not connected or the transport is already open.

        Args:
            callback: The function to call when a message is received, this
                      should only be provided if the transport instance
                      is a subscriber.
        """
        if not self._nats_client.is_connected:
            raise TTransportException(TTransportException.NOT_OPEN,
                                      'Nats is not connected')

        if self._is_open:
            raise TTransportException(TTransportException.ALREADY_OPEN,
                                      'Nats is already open')

        if not self._pull:
            self.reset_write_buffer()
            self._is_open = True
            return

        if not self._subject:
            raise TTransportException(TTransportException.UNKNOWN,
                                      'Subscriber cannot have an empty subject')

        self._sub_id = await self._nats_client.subscribe(
            'frugal.{0}'.format(self._subject),
            queue=self._queue,
            cb=lambda message: callback(TMemoryBuffer(message.data[4:]))
        )
        self._is_open = True

    async def close(self):
        """Close the transport."""
        if not self.isOpen():
            return

        if not self._pull:
            await self._nats_client.flush()
            self._is_open = False
            return

        await self._nats_client.unsubscribe(self._sub_id)
        self._sub_id = None
        self._is_open = False

    async def flush(self):
        """Flushes buffered write data over the network."""
        if not self.isOpen():
            raise TTransportException(TTransportException.NOT_OPEN,
                                      'Transport is not connected')

        frame = self.get_write_bytes()
        self.reset_write_buffer()

        await self._nats_client.publish(
            'frugal.{0}'.format(self._subject),
            frame
        )
