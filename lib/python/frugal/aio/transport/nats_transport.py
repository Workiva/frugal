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

from nats.aio.client import Client
from nats.aio.utils import new_inbox
from thrift.transport.TTransport import TTransportException

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.aio.transport import FAsyncTransport
from frugal.exceptions import TTransportExceptionType


class FNatsTransport(FAsyncTransport):
    """
    FNatsTransport is an extension of FAsyncTransport that uses nats as the
    underlying transport. This is "stateless" in the sense there is no
    connection with a server. A request is published on a subject and responses
    are received on another subject. To use this, requests and responses MUST
    fit within a single nats message.
    """
    def __init__(
        self,
        nats_client: Client,
        subject: str,
        inbox=''
    ):
        super().__init__(request_size_limit=_NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._sub_id = None

    def is_open(self):
        """Check whether the transport is open."""
        return self._is_open and self._nats_client.is_connected

    async def open(self):
        """Subscribe to the inbox subject."""
        if not self._nats_client.is_connected:
            raise TTransportException(TTransportExceptionType.NOT_OPEN,
                                      'Nats not connected')

        if self.is_open():
            raise TTransportException(TTransportExceptionType.ALREADY_OPEN,
                                      'Transport is already open')

        self._sub_id = await self._nats_client.subscribe_async(
            self._inbox,
            cb=self._on_message_callback
        )
        self._is_open = True

    async def _on_message_callback(self, message):
        await self.handle_response(message.data[4:])

    async def close(self):
        """Unsubscribe from the inbox subject."""
        if not self._sub_id:
            return

        await self._nats_client.unsubscribe(self._sub_id)
        self._is_open = False
        self._sub_id = None

    async def flush(self, data):
        await self._nats_client.publish_request(
            self._subject,
            self._inbox,
            data
        )
