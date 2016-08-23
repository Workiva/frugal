import logging
import struct

from nats.aio.client import Client
from thrift.Thrift import TException
from thrift.transport.TTransport import TMemoryBuffer

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.aio.processor import FProcessor
from frugal.protocol import FProtocolFactory
from frugal.server import FServer
from frugal.transport import FBoundedMemoryBuffer

logger = logging.getLogger(__name__)


class FNatsServer(FServer):
    """FNatsServer is an FServer that uses nats as the underlying transport."""
    def __init__(
            self,
            nats_client: Client,
            subject: str,
            processor: FProcessor,
            protocol_factory: FProtocolFactory,
            queue=''
    ):
        self._nats_client = nats_client
        self._subject = subject
        self._processor = processor
        self._protocol_factory = protocol_factory
        self._queue = queue
        self._sub_id = None

    async def serve(self):
        """Subscribe to the server subject and queue."""
        self._sub_id = await self._nats_client.subscribe(
            self._subject,
            queue=self._queue,
            cb=self._on_message_callback
        )
        logger.info('Frugal server running...')

    async def stop(self):
        """Unsubscribe from the server subject."""
        await self._nats_client.unsubscribe(self._sub_id)

    async def _on_message_callback(self, message):
        """The function to be executed when a message is received."""
        if not message.reply:
            logger.warn('no reply present, discarding message')
            return

        frame_size = struct.unpack('!I', message.data[:4])[0]
        if frame_size > _NATS_MAX_MESSAGE_SIZE - 4:
            logger.warning('frame size too large, dropping message')
            return

        # process frame, first four bytes are the frame size
        iprot = self._protocol_factory.get_protocol(
            TMemoryBuffer(message.data[4:])
        )
        out_transport = FBoundedMemoryBuffer(_NATS_MAX_MESSAGE_SIZE - 4)
        oprot = self._protocol_factory.get_protocol(out_transport)

        try:
            await self._processor.process(iprot, oprot)
        except TException as e:
            logger.exception(e)
            return

        if len(out_transport) == 0:
            return

        data = out_transport.getvalue()
        data_length = struct.pack('!I', len(data))
        await self._nats_client.publish(message.reply, data_length + data)
