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
from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FNatsPublisherTransport
from frugal.aio.transport import FNatsSubscriberTransport
from frugal.exceptions import TTransportExceptionType
from frugal.tests.aio import utils


class TestFNatsScopeTransport(utils.AsyncIOTestCase):
    def setUp(self):
        super().setUp()
        self.mock_nats_client = mock.Mock()
        self.callback = mock.Mock()
        self.queue = 'foo'
        self.pub_trans = FNatsPublisherTransport(self.mock_nats_client)
        self.sub_trans = FNatsSubscriberTransport(
            self.mock_nats_client, self.queue)

    @utils.async_runner
    async def test_publisher_open_nats_not_connected(self):
        self.mock_nats_client.is_connected = False
        with self.assertRaises(TTransportException) as cm:
            await self.pub_trans.open()
        self.assertEqual(TTransportExceptionType.NOT_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_publisher_close_not_open(self):
        self.mock_nats_client.is_connected = False
        await self.pub_trans.close()
        self.assertFalse(self.mock_nats_client.flush.called)

    @utils.async_runner
    async def test_close_publisher(self):
        self.pub_trans._is_open = True
        future = asyncio.Future()
        future.set_result(None)
        flush_mock = mock.Mock()
        flush_mock.return_value = future
        self.mock_nats_client.flush = flush_mock

        await self.pub_trans.close()
        self.assertTrue(flush_mock.called)

    @utils.async_runner
    async def test_publish(self):
        self.pub_trans._is_open = True
        data = bytearray([0, 0, 5, 2, 3, 4, 5, 6])
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.publish.return_value = future
        self.pub_trans._subject = 'bar'
        await self.pub_trans.publish('bar', data)

        self.mock_nats_client.publish.assert_called_once_with(
            'frugal.bar',
            data
        )

    @utils.async_runner
    async def test_subscribe(self):
        future = asyncio.Future()
        future.set_result(235)
        self.mock_nats_client.subscribe.return_value = future

        topic = 'bar'
        await self.sub_trans.subscribe(topic, self.callback)
        self.mock_nats_client.subscribe.assert_called_once_with(
            'frugal.bar',
            queue=self.queue,
            cb=mock.ANY,
        )
        self.assertEqual(self.sub_trans._sub_id, 235)

    @utils.async_runner
    async def test_subscribe_nats_not_connected(self):
        self.mock_nats_client.is_connected = False
        with self.assertRaises(TTransportException) as cm:
            await self.sub_trans.subscribe('foo', None)
        self.assertEqual(TTransportExceptionType.NOT_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_subscribe_open_already(self):
        self.mock_nats_client.is_connected = True
        self.sub_trans._is_subscribed = True
        with self.assertRaises(TTransportException) as cm:
            await self.sub_trans.subscribe('foo', None)
        self.assertEqual(TTransportExceptionType.ALREADY_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_unsubscribe(self):
        self.sub_trans._is_subscribed = True
        self.sub_trans._sub_id = 235
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = future
        await self.sub_trans.unsubscribe()

        self.assertIsNone(self.sub_trans._sub_id)
        self.assertFalse(self.sub_trans._is_subscribed)
        self.mock_nats_client.unsubscribe.assert_called_once_with(235)
