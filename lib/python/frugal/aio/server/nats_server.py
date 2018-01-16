# Copyright 2017 Workiva
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import logging
import struct
from typing import List

from nats.aio.client import Client
from thrift.Thrift import TApplicationException
from thrift.transport.TTransport import TMemoryBuffer

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.aio.processor import FProcessor
from frugal.protocol import FProtocolFactory
from frugal.server import FServer
from frugal.transport import TMemoryOutputBuffer

logger = logging.getLogger(__name__)


class FNatsServer(FServer):
    """FNatsServer is an FServer that uses nats as the underlying transport."""
    def __init__(
            self,
            nats_client: Client,
            subjects: List[str],
            processor: FProcessor,
            protocol_factory: FProtocolFactory,
            queue=''
    ):
        self._nats_client = nats_client
        self._subjects = [subjects] if isinstance(subjects, str) else subjects
        self._processor = processor
        self._protocol_factory = protocol_factory
        self._queue = queue
        self._sub_ids = []

    async def serve(self):
        """Subscribe to the server subject and queue."""
        self._sub_ids = []
        for subject in self._subjects:
            self._sub_ids.append(await self._nats_client.subscribe_async(
                subject,
                queue=self._queue,
                cb=self._on_message_callback,
            ))
        logger.info('Frugal server running...')

    async def stop(self):
        """Unsubscribe from the server subject."""
        for sid in self._sub_ids:
            await self._nats_client.unsubscribe(sid)

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
        otrans = TMemoryOutputBuffer(_NATS_MAX_MESSAGE_SIZE)
        oprot = self._protocol_factory.get_protocol(otrans)

        try:
            await self._processor.process(iprot, oprot)
        except TApplicationException:
            # Continue so the exception is sent to the client
            pass
        except Exception:
            return

        if len(otrans) == 4:
            return

        await self._nats_client.publish(message.reply, otrans.getvalue())
