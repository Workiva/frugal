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

from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportException
from tornado import gen

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.exceptions import TTransportExceptionType
from frugal.tornado.transport import FAsyncTransport

_NOT_OPEN = 'NATS not connected.'
_ALREADY_OPEN = 'NATS transport already open.'


class FNatsTransport(FAsyncTransport):
    """
    FNatsTransport is an extension of FAsyncTransport. This is a "stateless"
    transport in the sense that there is no connection with a server. A request
    is simply published to a subject and responses are received on another
    subject. This assumes requests/responses fit within a single NATS message.
    """

    def __init__(self, nats_client, subject, inbox=""):
        """
        Create a new instance of FNatsTransport

        Args:
            nats_client: connected instance of nats.io.Client
            subject: subject to publish to
        """
        super(FNatsTransport, self).__init__(
            request_size_limit=_NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._sub_id = None

    def is_open(self):
        return self._is_open and self._nats_client.is_connected

    @gen.coroutine
    def open(self):
        """
        Subscribes to the configured inbox subject.
        """
        if not self._nats_client.is_connected:
            raise TTransportException(
                type=TTransportExceptionType.NOT_OPEN,
                message=_NOT_OPEN)

        elif self.is_open():
            already_open = TTransportExceptionType.ALREADY_OPEN
            raise TTransportException(already_open, _ALREADY_OPEN)

        cb = self._on_message_callback
        inbox = self._inbox
        self._sub_id = yield self._nats_client.subscribe_async(inbox, cb=cb)

        self._is_open = True

    @gen.coroutine
    def _on_message_callback(self, msg):
        yield self.handle_response(msg.data[4:])

    @gen.coroutine
    def close(self):
        """
        Unsubscribes from the inbox subject
        """
        if not self._sub_id:
            return
        yield self._nats_client.flush()
        yield self._nats_client.unsubscribe(self._sub_id)
        self._is_open = False

    @gen.coroutine
    def flush(self, payload):
        """
        Sends the buffered bytes over NATS
        """
        subject = self._subject
        inbox = self._inbox
        yield self._nats_client.publish_request(subject, inbox, payload)
