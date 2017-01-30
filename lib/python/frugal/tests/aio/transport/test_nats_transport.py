import asyncio

import mock
from nats.aio.client import Client
from thrift.transport.TTransport import TTransportException

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.aio.transport import FNatsTransport
from frugal.exceptions import FrugalTTransportExceptionType
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

    @mock.patch('frugal.aio.transport.nats_transport.new_inbox')
    def test_init(self, mock_new_inbox):
        self.assertEqual(self.mock_nats_client, self.transport._nats_client)
        self.assertEqual(self.subject, self.transport._subject)
        self.assertEqual(self.inbox, self.transport._inbox)

        mock_new_inbox.return_value = 'a new inbox'
        transport = FNatsTransport(self.mock_nats_client,
                                   self.subject)
        mock_new_inbox.assert_called_once_with()
        self.assertEqual('a new inbox', transport._inbox)

    @utils.async_runner
    async def test_open_nats_not_connected(self):
        self.mock_nats_client.is_connected = False

        with self.assertRaises(TTransportException) as cm:
            await self.transport.open()
        self.assertEqual(FrugalTTransportExceptionType.NOT_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_open_already_open(self):
        self.mock_nats_client.is_connected = True
        self.transport._is_open = True

        with self.assertRaises(TTransportException) as cm:
            await self.transport.open()
        self.assertEqual(FrugalTTransportExceptionType.ALREADY_OPEN, cm.exception.type)

    @utils.async_runner
    async def test_open_subscribes(self):
        future = asyncio.Future()
        future.set_result(235)
        self.mock_nats_client.subscribe.return_value = future
        await self.transport.open()

        self.assertEqual(235, self.transport._sub_id)
        self.mock_nats_client.subscribe.assert_called_once_with(
            self.inbox,
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
        callback.assert_called_once_with(message.data[4:])

    @utils.async_runner
    async def test_close_not_subscribed(self):
        self.transport._sub_id = None
        await self.transport.close()
        self.assertFalse(self.mock_nats_client.unsubscribe.called)

    @utils.async_runner
    async def test_close_unsubscribes(self):
        self.transport._is_open = True
        self.transport._sub_id = 235
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = future
        await self.transport.close()

        self.assertIsNone(self.transport._sub_id)
        self.assertFalse(self.transport._is_open)
        self.mock_nats_client.unsubscribe.assert_called_once_with(235)

    @utils.async_runner
    async def test_flush(self):
        self.transport._is_open = True
        data = bytearray([2, 3, 4, 5, 6, 7])
        data_len = bytearray([0, 0, 0, 6])
        frame = data_len + data
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.publish_request.return_value = future
        await self.transport.flush(frame)

        self.mock_nats_client.publish_request.assert_called_once_with(
            self.subject,
            self.inbox,
            frame
        )

    def test_request_size_limit(self):
        self.assertEqual(_NATS_MAX_MESSAGE_SIZE,
                         self.transport.get_request_size_limit())
