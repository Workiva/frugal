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

from thrift.transport.TTransport import TTransportException, TMemoryBuffer
from tornado import gen

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.transport import FPublisherTransportFactory
from frugal.transport import FPublisherTransport
from frugal.transport import FSubscriberTransportFactory
from frugal.transport import FSubscriberTransport
from frugal.exceptions import TTransportExceptionType

_FRAME_BUFFER_SIZE = 5
_FRUGAL_PREFIX = "frugal."

logger = logging.getLogger(__name__)


class FNatsPublisherTransportFactory(FPublisherTransportFactory):
    def __init__(self, nats_client):
        self._nats_client = nats_client

    def get_transport(self):
        return FNatsPublisherTransport(self._nats_client)


class FNatsPublisherTransport(FPublisherTransport):
    def __init__(self, nats_client):
        super(FNatsPublisherTransport, self).__init__(_NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client

    @gen.coroutine
    def open(self):
        if not self._nats_client.is_connected:
            raise TTransportException(
                type=TTransportExceptionType.NOT_OPEN,
                message="Nats not connected!")

    @gen.coroutine
    def close(self):
        if not self.is_open():
            return

        yield self._nats_client.flush()

    def is_open(self):
        return self._nats_client.is_connected

    @gen.coroutine
    def publish(self, topic, data):
        if not self.is_open():
            raise TTransportException(
                type=TTransportExceptionType.NOT_OPEN,
                message='Nats not connected!')
        if self._check_publish_size(data):
            msg = 'Message exceeds NATS max message size'
            raise TTransportException(
                type=TTransportExceptionType.REQUEST_TOO_LARGE,
                message=msg)
        yield self._nats_client.publish('frugal.{0}'.format(topic), data)


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
        self._sub_id = None

    @gen.coroutine
    def subscribe(self, topic, callback):
        if not self._nats_client.is_connected:
            raise TTransportException(
                type=TTransportExceptionType.NOT_OPEN,
                message="Nats not connected!")

        if self.is_subscribed():
            raise TTransportException(
                type=TTransportExceptionType.ALREADY_OPEN,
                message="Already subscribed to nats topic!")

        self._sub_id = yield self._nats_client.subscribe_async(
            'frugal.{0}'.format(topic),
            queue=self._queue,
            cb=lambda message: callback(TMemoryBuffer(message.data[4:]))
        )
        self._is_subscribed = True

    @gen.coroutine
    def unsubscribe(self):
        if not self.is_subscribed():
            return

        yield self._nats_client.unsubscribe(self._sub_id)
        self._sub_id = None
        self._is_subscribed = False

    def is_subscribed(self):
        return self._is_subscribed and self._nats_client.is_connected
