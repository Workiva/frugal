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

import asyncio

import mock
from nats.aio.client import Client
from nats.aio.subscription import Subscription
from thrift.transport.TTransport import TTransportException

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.aio.transport import FNatsTransport
from frugal.exceptions import TTransportExceptionType
from frugal.tests.aio import utils


class TestFNatsTransport(utils.AsyncIOTestCase):

    def setUp(self):
        super().setUp()
        self.mock_nats_client = mock.Mock(spec=Client)
        self.subject = 'foo'
        self.inbox = 'bar'
        self.transport = FNatsTransport(
            self.mock_nats_client,
            self.subject,
            inbox=self.inbox
        )

    def test_init(self):
        self.assertEqual(self.mock_nats_client, self.transport._nats_client)
        self.assertEqual(self.subject, self.transport._subject)
        self.assertEqual(self.inbox, self.transport._inbox)

        self.mock_nats_client.new_inbox.return_value = 'a new inbox'
        transport = FNatsTransport(self.mock_nats_client,
                                   self.subject)
        self.mock_nats_client.new_inbox.assert_called_once_with()
        self.assertEqual('a new inbox', transport._inbox)

    @utils.async_runner
    async def test_open_nats_not_connected(self):
        self.mock_nats_client.is_connected = False

        with self.assertRaises(TTransportException) as cm:
            await self.transport.open()
        self.assertEqual(TTransportExceptionType.NOT_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_open_already_open(self):
        self.mock_nats_client.is_connected = True
        self.transport._is_open = True

        with self.assertRaises(TTransportException) as cm:
            await self.transport.open()
        self.assertEqual(TTransportExceptionType.ALREADY_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_open_subscribes(self):
        future = asyncio.Future()
        mock_sub = mock.Mock(spec=Subscription)
        future.set_result(mock_sub)
        self.mock_nats_client.subscribe.return_value = future
        await self.transport.open()

        self.mock_nats_client.subscribe.assert_called_once_with(
            self.inbox + ".*",
            cb=self.transport._on_message_callback
        )
        self.assertTrue(self.transport._is_open)

    @utils.async_runner
    async def test_on_message_callback(self):
        message = mock.Mock()
        message.data = [1, 2, 3, 4, 5, 6, 7, 8, 9]
        callback = mock.Mock()
        future = asyncio.Future()
        future.set_result(None)
        callback.return_value = future
        self.transport.handle_response = callback
        await self.transport._on_message_callback(message)
        callback.assert_called_once_with(message)

    @utils.async_runner
    async def test_close_not_subscribed(self):
        self.transport._subscription = None
        await self.transport.close()
        assert(True)

    @utils.async_runner
    async def test_close_unsubscribes(self):
        self.transport._is_open = True
        mock_sub = mock.Mock(spec=Subscription)
        self.transport._subscription = mock_sub
        future = asyncio.Future()
        future.set_result(None)
        mock_sub.unsubscribe.return_value = future
        await self.transport.close()

        self.assertIsNone(self.transport._subscription)
        self.assertFalse(self.transport._is_open)
        mock_sub.unsubscribe.assert_called_once()        

    @utils.async_runner
    async def test_flush(self):
        self.transport._is_open = True
        data = bytearray([2, 3, 4, 5, 6, 7])
        data_len = bytearray([0, 0, 0, 6])
        frame = data_len + data
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.publish.return_value = future
        await self.transport.flush(frame)

        self.mock_nats_client.publish.assert_called_once_with(
            self.subject,
            frame,
            self.inbox
        )

    @utils.async_runner
    async def test_flush_op(self):
        self.transport._is_open = True
        op_id = 1
        data = bytearray([2, 3, 4, 5, 6, 7])
        data_len = bytearray([0, 0, 0, 6])
        frame = data_len + data
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.publish.return_value = future
        await self.transport.flush_op(op_id, frame)

        self.mock_nats_client.publish.assert_called_once_with(
            self.subject,
            frame,
            f"{self.inbox}.{op_id}",
        )

    def test_request_size_limit(self):
        self.assertEqual(_NATS_MAX_MESSAGE_SIZE,
                         self.transport.get_request_size_limit())

    @utils.async_runner
    async def test_handle_status_message(self):
        message = mock.Mock()
        message.data = []
        message.subject = "subject.1"
        await self.transport._on_message_callback(message)