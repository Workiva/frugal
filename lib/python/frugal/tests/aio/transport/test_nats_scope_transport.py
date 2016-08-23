import asyncio

import mock
from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FNatsScopeTransport
from frugal.exceptions import FException
from frugal.tests.aio import utils


class TestFNatsScopeTransport(utils.AsyncIOTestCase):
    def setUp(self):
        super().setUp()
        self.mock_nats_client = mock.Mock()
        self.callback = mock.Mock()
        self.queue = 'foo'
        self.transport = FNatsScopeTransport(
            self.mock_nats_client,
            queue=self.queue
        )

    @utils.async_runner
    async def test_lock_topic(self):
        topic = 'bar'
        await self.transport.lock_topic(topic)
        self.assertEqual(topic, self.transport._subject)

    @utils.async_runner
    async def test_unlock_topic(self):
        topic = 'bar'
        await self.transport.lock_topic(topic)
        self.transport.unlock_topic()
        self.assertEqual('', self.transport._subject)

    @utils.async_runner
    async def test_subscribe(self):
        open_mock = mock.Mock()
        future = asyncio.Future()
        future.set_result(None)
        open_mock.return_value = future
        self.transport.open = open_mock

        topic = 'bar'
        await self.transport.subscribe(topic, self.callback)
        self.assertTrue(self.transport._pull)
        self.assertEqual(topic, self.transport._subject)
        open_mock.assert_called_once_with(callback=self.callback)

    @utils.async_runner
    async def test_subscriber_lock_unlock_topic(self):
        open_mock = mock.Mock()
        future = asyncio.Future()
        future.set_result(None)
        open_mock.return_value = future
        self.transport.open = open_mock

        topic = 'bar'
        await self.transport.subscribe(topic, self.callback)

        with self.assertRaises(FException):
            await self.transport.lock_topic(topic)

        with self.assertRaises(FException):
            self.transport.unlock_topic()

    @utils.async_runner
    async def test_open_nats_not_connected(self):
        self.mock_nats_client.is_connected = False
        with self.assertRaises(TTransportException) as cm:
            await self.transport.open()
        self.assertEqual(TTransportException.NOT_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_open_already_open(self):
        self.mock_nats_client.is_connected = True
        self.transport._is_open = True
        with self.assertRaises(TTransportException) as cm:
            await self.transport.open()
        self.assertEqual(TTransportException.ALREADY_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_open_publisher(self):
        self.mock_nats_client.is_connected = True
        self.transport._is_open = False

    @utils.async_runner
    async def test_open_subscriber_empty_subject(self):
        self.transport._pull = True
        with self.assertRaises(TTransportException) as cm:
            await self.transport.open()
        self.assertEqual(TTransportException.UNKNOWN, cm.exception.type)

    @utils.async_runner
    async def test_open_subscriber(self):
        self.transport._pull = True
        topic = 'bar'
        self.transport._subject = topic
        future = asyncio.Future()
        future.set_result(235)
        self.mock_nats_client.subscribe.return_value = future

        await self.transport.open()
        self.assertEqual(235, self.transport._sub_id)
        self.assertTrue(self.transport._is_open)
        self.mock_nats_client.subscribe.assert_called_once_with(
            'frugal.bar',
            queue=self.queue,
            cb=mock.ANY
        )

    @utils.async_runner
    async def test_close_not_open(self):
        self.transport._is_open = False
        await self.transport.close()
        self.assertFalse(self.mock_nats_client.unsubscribe.called)

    @utils.async_runner
    async def test_close_publisher(self):
        self.transport._is_open = True
        self.transport._pull = False
        future = asyncio.Future()
        future.set_result(None)
        flush_mock = mock.Mock()
        flush_mock.return_value = future
        self.mock_nats_client.flush = flush_mock

        await self.transport.close()
        self.assertFalse(self.transport._is_open)
        self.assertFalse(self.mock_nats_client.unsubscribe.called)
        self.assertTrue(flush_mock.called)

    @utils.async_runner
    async def test_close_subscriber(self):
        self.transport._is_open = True
        self.transport._pull = True
        self.transport._sub_id = 235
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = future
        await self.transport.close()

        self.assertIsNone(self.transport._sub_id)
        self.assertFalse(self.transport._is_open)
        self.mock_nats_client.unsubscribe.assert_called_once_with(235)

    @utils.async_runner
    async def test_flush_not_open(self):
        self.transport._is_open = False
        with self.assertRaises(TTransportException) as cm:
            await self.transport.flush()
        self.assertEqual(TTransportException.NOT_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_flush(self):
        self.transport._is_open = True
        data = bytearray([2, 3, 4, 5, 6])
        self.transport.write(data)
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.publish.return_value = future
        self.transport._subject = 'bar'
        await self.transport.flush()

        self.assertEqual(0, len(self.transport._wbuf.getvalue()))
        self.mock_nats_client.publish.assert_called_once_with(
            'frugal.bar',
            bytearray([0, 0, 0, 5]) + data
        )
