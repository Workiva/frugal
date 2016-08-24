from asyncio import Future
import base64

import mock
from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FHttpTransport
from frugal.exceptions import FMessageSizeException
from frugal.tests.aio import utils


class TestFHttpTransport(utils.AsyncIOTestCase):
    def setUp(self):
        super().setUp()

        self.url = 'http://localhost/testing'
        self.request_capacity = 100
        self.response_capacity = 200
        self.transport = FHttpTransport(
            self.url,
            request_capacity=self.request_capacity,
            response_capacity=self.response_capacity
        )
        self.make_request_mock = mock.Mock()
        self.transport._make_request = self.make_request_mock

        self.headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
            'accept': 'application/x-frugal',
            'x-frugal-payload-limit': '200',
        }

    def test_blah(self):
        pass

    @utils.async_runner
    async def test_open_close(self):
        self.assertTrue(self.transport.isOpen())
        await self.transport.open()
        self.assertTrue(self.transport.isOpen())
        await self.transport.close()
        self.assertTrue(self.transport.isOpen())

    @utils.async_runner
    async def test_write_too_much_data(self):
        with self.assertRaises(FMessageSizeException):
            self.transport.write(b'0' * 101)

    @utils.async_runner
    async def test_flush_success(self):
        registry_mock = mock.Mock()
        registry_future = Future()
        registry_future.set_result(None)
        registry_mock.execute.return_value = registry_future
        self.transport.set_registry(registry_mock)

        request_data = bytearray([4, 5, 6, 7, 8, 9, 10, 11, 13, 12, 3])
        expected_payload = bytearray([0, 0, 0, 11]) + request_data

        response_data = bytearray([23, 24, 25, 26, 27, 28, 29])
        response_frame = bytearray([0, 0, 0, 7]) + response_data
        response_encoded = base64.b64encode(response_frame)
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        self.transport.write(request_data[:3])
        self.transport.write(request_data[3:7])
        self.transport.write(request_data[7:])
        await self.transport.flush()

        self.assertTrue(self.make_request_mock.called)
        request = self.make_request_mock.call_args[0][0]
        self.assertEqual(request, base64.b64encode(expected_payload))

        registry_mock.execute.assert_called_once_with(response_data)

    @utils.async_runner
    async def test_flush_invalid_response_frame(self):
        response_encoded = base64.b64encode(bytearray([4, 5]))
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        self.transport.write(bytearray([1, 2, 3, 4]))
        with self.assertRaises(TTransportException):
            await self.transport.flush()

        self.assertTrue(self.make_request_mock.called)

    @utils.async_runner
    async def test_flush_oneway(self):
        registry_mock = mock.Mock()
        self.transport.set_registry(registry_mock)

        response_encoded = base64.b64encode(bytearray([0, 0, 0, 0]))
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        self.transport.write(bytearray([1, 2, 3]))
        await self.transport.flush()

        self.assertTrue(self.make_request_mock.called)
        self.assertFalse(registry_mock.execute.called)

    @utils.async_runner
    async def test_flush_missing_data(self):
        response_encoded = base64.b64encode(bytearray([0, 0, 0, 1]))
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        self.transport.write(bytearray([1, 2, 3]))
        with self.assertRaises(TTransportException) as e:
            await self.transport.flush()

        self.assertEqual(str(e.exception), 'missing data')

    @utils.async_runner
    async def test_flush_response_too_large(self):
        message = b'something went wrong'
        encoded_message = base64.b64encode(message)
        response_future = Future()
        response_future.set_result((413, encoded_message))
        self.make_request_mock.return_value = response_future

        self.transport.write(bytearray([1, 2, 3]))
        with self.assertRaises(FMessageSizeException) as e:
            await self.transport.flush()

        self.assertEqual(str(e.exception),
                         'response was too large for the transport')

    @utils.async_runner
    async def test_flush_response_error(self):
        message = b'something went wrong'
        encoded_message = base64.b64encode(message)
        response_future = Future()
        response_future.set_result((404, encoded_message))
        self.make_request_mock.return_value = response_future

        self.transport.write(bytearray([1, 2, 3]))
        with self.assertRaises(TTransportException) as e:
            await self.transport.flush()

        self.assertEqual(str(e.exception),
                         'request errored with code {0} and message {1}'.format(
                             404, str(message)
                         ))
