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

from asyncio import Future
import base64
import mock

from aiohttp.client import ClientConnectorError
from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FHttpTransport
from frugal.exceptions import TTransportExceptionType
from frugal.context import FContext
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
        self.assertTrue(self.transport.is_open())
        await self.transport.open()
        self.assertTrue(self.transport.is_open())
        await self.transport.close()
        self.assertTrue(self.transport.is_open())

    @utils.async_runner
    async def test_oneway(self):
        response_encoded = base64.b64encode(bytearray([0, 0, 0, 0]))
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        self.assertIsNone(await self.transport.oneway(
            FContext(), bytearray([0, 0, 0, 3, 1, 2, 3])
        ))

        self.assertTrue(self.make_request_mock.called)

    @utils.async_runner
    async def test_request(self):
        request_data = bytearray([4, 5, 6, 7, 8, 9, 10, 11, 13, 12, 3])
        request_frame = bytearray([0, 0, 0, 11]) + request_data

        response_data = bytearray([23, 24, 25, 26, 27, 28, 29])
        response_frame = bytearray([0, 0, 0, 7]) + response_data
        response_encoded = base64.b64encode(response_frame)
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        ctx = FContext()
        response_transport = await self.transport.request(
            ctx, request_frame)

        self.assertEqual(response_data, response_transport.getvalue())
        self.assertTrue(self.make_request_mock.called)
        request_args = self.make_request_mock.call_args[0]
        self.assertEqual(request_args[0], ctx)
        self.assertEqual(request_args[1], base64.b64encode(request_frame))

    @utils.async_runner
    async def test_request_extra_headers_with_context(self):

        def generate_test_header(fcontext):
            return {
                'first-header': fcontext.correlation_id,
                'second-header': 'test'
            }

        transport_with_headers = FHttpTransport(
            self.url,
            request_capacity=self.request_capacity,
            response_capacity=self.response_capacity,
            get_request_headers=generate_test_header
        )

        transport_with_headers._make_request = self.make_request_mock

        request_data = bytearray([4, 5, 6, 7, 8, 9, 10, 11, 13, 12, 3])
        request_frame = bytearray([0, 0, 0, 11]) + request_data

        response_data = bytearray([23, 24, 25, 26, 27, 28, 29])
        response_frame = bytearray([0, 0, 0, 7]) + response_data
        response_encoded = base64.b64encode(response_frame)
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        ctx = FContext()
        response_transport = await transport_with_headers.request(
            ctx, request_frame)

        self.assertEqual(response_data, response_transport.getvalue())
        self.assertTrue(self.make_request_mock.called)
        request_args = self.make_request_mock.call_args[0]
        self.assertEqual(request_args[0], ctx)
        self.assertEqual(request_args[1], base64.b64encode(request_frame))

    @utils.async_runner
    async def test_request_too_much_data(self):
        with self.assertRaises(TTransportException) as cm:
            await self.transport.request(FContext(), b'0' * 101)
        self.assertEqual(TTransportExceptionType.REQUEST_TOO_LARGE,
                         cm.exception.type)

    @utils.async_runner
    async def test_request_invalid_response_frame(self):
        response_encoded = base64.b64encode(bytearray([4, 5]))
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        with self.assertRaises(TTransportException):
            await self.transport.request(
                FContext(), bytearray([0, 0, 0, 4, 1, 2, 3, 4])
            )

        self.assertTrue(self.make_request_mock.called)

    @utils.async_runner
    async def test_request_missing_data(self):
        response_encoded = base64.b64encode(bytearray([0, 0, 0, 1]))
        response_future = Future()
        response_future.set_result((200, response_encoded))
        self.make_request_mock.return_value = response_future

        with self.assertRaises(TTransportException) as e:
            await self.transport.request(
                FContext(), bytearray([0, 0, 0, 31, 2, 3]))

        self.assertEqual(str(e.exception), 'missing data')

    @utils.async_runner
    async def test_request_response_too_large(self):
        message = b'something went wrong'
        encoded_message = base64.b64encode(message)
        response_future = Future()
        response_future.set_result((413, encoded_message))
        self.make_request_mock.return_value = response_future

        with self.assertRaises(TTransportException) as e:
            await self.transport.request(
                FContext(), bytearray([0, 0, 0, 3, 1, 2, 3]))

        self.assertEqual(TTransportExceptionType.RESPONSE_TOO_LARGE,
                         e.exception.type)
        self.assertEqual(str(e.exception),
                         'response was too large for the transport')

    @utils.async_runner
    async def test_request_response_error(self):
        message = b'something went wrong'
        response_future = Future()
        response_future.set_result((404, message))
        self.make_request_mock.return_value = response_future

        with self.assertRaises(TTransportException) as e:
            await self.transport.request(
                FContext(), bytearray([0, 0, 0, 3, 1, 2, 3]))

        self.assertEqual(
            str(e.exception),
            'request errored with code {0} and message {1}'.format(
                404, str(message)
            )
        )

    @utils.async_runner
    async def test_request_service_unavailable(self):
        transport = FHttpTransport(
            self.url,
            request_capacity=self.request_capacity,
            response_capacity=self.response_capacity
        )

        with self.assertRaises(TTransportException) as e:
            await transport.request(
                FContext(), bytearray([0, 0, 0, 3, 1, 2, 3])
            )

        self.assertTrue(
            "service not available: " in str(e.exception)
        )
