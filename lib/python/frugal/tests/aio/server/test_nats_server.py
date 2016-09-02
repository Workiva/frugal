import asyncio
from io import BytesIO

import mock

from frugal.tests.aio import utils
from frugal.aio.server import FNatsServer


class TestFNatsServer(utils.AsyncIOTestCase):
    def setUp(self):
        super().setUp()
        self.subjects = ['foo', 'bar']
        self.mock_nats_client = mock.Mock()
        self.mock_processor = mock.Mock()
        self.mock_transport_factory = mock.Mock()
        self.mock_prot_factory = mock.Mock()

        self.server = FNatsServer(
            self.mock_nats_client,
            self.subjects,
            self.mock_processor,
            self.mock_prot_factory
        )

    @utils.async_runner
    async def test_serve(self):
        future1 = asyncio.Future()
        future1.set_result(235)
        future2 = asyncio.Future()
        future2.set_result(694)
        self.mock_nats_client.subscribe.side_effect = [future1, future2]
        await self.server.serve()
        self.assertEqual([235, 694], self.server._sub_ids)
        self.mock_nats_client.subscribe.assert_has_calls([
            mock.call(
                self.subjects[0],
                queue='',
                cb=self.server._on_message_callback
            ),
            mock.call(
                self.subjects[1],
                queue='',
                cb=self.server._on_message_callback
            )
        ])

    @utils.async_runner
    async def test_stop(self):
        self.server._sub_ids = [235, 694]
        future1 = asyncio.Future()
        future1.set_result(None)
        future2 = asyncio.Future()
        future2.set_result(None)
        self.mock_nats_client.unsubscribe.side_effect = [future1, future2]
        await self.server.stop()
        self.mock_nats_client.unsubscribe.assert_has_calls([
            mock.call(235),
            mock.call(694),
        ])

    @utils.async_runner
    async def test_on_message_callback_no_reply(self):
        data = bytearray([1, 2, 3, 4, 5])
        data_size = bytearray([0, 0, 0, 5])
        message = mock.Mock(subject='foo', reply='', data=data_size + data)
        await self.server._on_message_callback(message)
        self.assertFalse(self.mock_processor.process.called)
        self.assertFalse(self.mock_prot_factory.get_protocol.called)

    @mock.patch('frugal.aio.server.nats_server._NATS_MAX_MESSAGE_SIZE', 6)
    @utils.async_runner
    async def test_on_message_callback_large_frame_size(self):
        data = bytearray([1, 2, 3, 4, 5])
        data_size = bytearray([0, 0, 0, 5])
        message = mock.Mock(subject='foo', reply='bar', data=data_size + data)
        await self.server._on_message_callback(message)
        self.assertFalse(self.mock_processor.process.called)
        self.assertFalse(self.mock_prot_factory.get_protocol.called)

    @utils.async_runner
    async def test_on_message_callback(self):
        data = bytearray([1, 2, 3, 4, 5])
        data_size = bytearray([0, 0, 0, 5])
        message = mock.Mock(subject='foo', reply='bar', data=data_size + data)
        iprot = BytesIO()
        oprot = BytesIO()
        self.server._protocol_factory.get_protocol.side_effect = [
            iprot,
            oprot,
        ]

        future = asyncio.Future()
        future.set_result(None)
        self.mock_processor.process.return_value = future

        await self.server._on_message_callback(message)
        self.mock_processor.process.assert_called_once_with(iprot, oprot)
