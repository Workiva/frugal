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

import base64
import mock
import socket

from thrift.transport.TTransport import TTransportException
from tornado.concurrent import Future
from tornado.httpclient import AsyncHTTPClient
from tornado.httpclient import HTTPError
from tornado.httpclient import HTTPResponse
from tornado.testing import AsyncTestCase
from tornado.testing import gen_test

from frugal.context import FContext
from frugal.exceptions import TTransportExceptionType
from frugal.tornado.transport.http_transport import FHttpTransport


class TestFHttpTransport(AsyncTestCase):
    def setUp(self):
        super(TestFHttpTransport, self).setUp()

        self.url = 'http://localhost/testing'
        self.request_capacity = 100
        self.response_capacity = 200
        self.transport = FHttpTransport(
            url=self.url,
            request_capacity=self.request_capacity,
            response_capacity=self.response_capacity
        )
        self.http_mock = mock.Mock(spec=AsyncHTTPClient)
        self.headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
            'accept': 'application/x-frugal',
            'x-frugal-payload-limit': '200',
        }

    @gen_test
    def test_open_close(self):
        self.assertTrue((yield self.transport.is_open()))
        yield self.transport.open()
        self.assertTrue((yield self.transport.is_open()))
        self.assertIsNotNone(self.transport._http)
        yield self.transport.close()
        self.assertTrue((yield self.transport.is_open()))
        self.assertIsNotNone(self.transport._http)

    @gen_test
    def test_oneway(self):
        callback_mock = mock.Mock()
        self.transport._http = self.http_mock

        response_encoded = base64.b64encode(bytearray([0, 0, 0, 0]))
        response_mock = mock.Mock(spec=HTTPResponse)
        response_mock.body = response_encoded
        response_future = Future()
        response_future.set_result(response_mock)
        self.http_mock.fetch.return_value = response_future

        response = yield self.transport.oneway(
            FContext(), bytearray([0, 0, 0, 3, 1, 2, 3]))

        self.assertIsNone(response)
        self.assertTrue(self.http_mock.fetch.called)
        self.assertFalse(callback_mock.called)

    @gen_test
    def test_request(self):
        self.transport._http = self.http_mock

        request_data = bytearray([4, 5, 6, 8, 9, 10, 11, 13, 12, 3])
        request_frame = bytearray([0, 0, 0, 10]) + request_data

        response_mock = mock.Mock(spec=HTTPResponse)
        response_data = bytearray([23, 24, 25, 26, 27, 28, 29])
        response_frame = bytearray([0, 0, 0, 10]) + response_data
        response_encoded = base64.b64encode(response_frame)
        response_mock.body = response_encoded
        response_future = Future()
        response_future.set_result(response_mock)
        self.http_mock.fetch.return_value = response_future

        ctx = FContext()
        response_transport = yield self.transport.request(ctx, request_frame)

        self.assertEqual(response_data, response_transport.getvalue())
        self.assertTrue(self.http_mock.fetch.called)
        request = self.http_mock.fetch.call_args[0][0]
        self.assertEqual(request.url, self.url)
        self.assertEqual(request.method, 'POST')
        self.assertEqual(request.body, base64.b64encode(request_frame))
        self.assertEqual(request.headers, self.headers)

    @gen_test
    def test_request_extra_headers_with_context(self):

        def generate_test_header(fcontext):
            return {
                'first-header': fcontext.correlation_id,
                'second-header': 'test'
            }

        transport_with_headers = FHttpTransport(
            url=self.url,
            request_capacity=self.request_capacity,
            response_capacity=self.response_capacity,
            get_request_headers=generate_test_header
        )

        transport_with_headers._http = self.http_mock

        request_data = bytearray([4, 5, 6, 8, 9, 10, 11, 13, 12, 3])
        request_frame = bytearray([0, 0, 0, 10]) + request_data

        response_mock = mock.Mock(spec=HTTPResponse)
        response_data = bytearray([23, 24, 25, 26, 27, 28, 29])
        response_frame = bytearray([0, 0, 0, 10]) + response_data
        response_encoded = base64.b64encode(response_frame)
        response_mock.body = response_encoded
        response_future = Future()
        response_future.set_result(response_mock)
        self.http_mock.fetch.return_value = response_future

        ctx = FContext()
        response_transport = yield transport_with_headers.request(ctx, request_frame)
        expected_headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
            'accept': 'application/x-frugal',
            'x-frugal-payload-limit': '200',
            'first-header': ctx.correlation_id,
            'second-header': 'test',
        }

        self.assertEqual(response_data, response_transport.getvalue())
        self.assertTrue(self.http_mock.fetch.called)
        request = self.http_mock.fetch.call_args[0][0]
        self.assertEqual(request.url, self.url)
        self.assertEqual(request.method, 'POST')
        self.assertEqual(request.body, base64.b64encode(request_frame))
        self.assertEqual(request.headers, expected_headers)

    @gen_test
    def test_request_extra_headers_with_context_modify_defaults(self):

        def generate_test_header(fcontext):
            return {
                'first-header': fcontext.correlation_id,
                'second-header': 'test',
                'content-type': 'these should',
                'content-transfer-encoding': 'not be',
                'accept': 'the final values'
            }

        transport_with_headers = FHttpTransport(
            url=self.url,
            request_capacity=self.request_capacity,
            response_capacity=self.response_capacity,
            get_request_headers=generate_test_header
        )

        transport_with_headers._http = self.http_mock

        request_data = bytearray([4, 5, 6, 8, 9, 10, 11, 13, 12, 3])
        request_frame = bytearray([0, 0, 0, 10]) + request_data

        response_mock = mock.Mock(spec=HTTPResponse)
        response_data = bytearray([23, 24, 25, 26, 27, 28, 29])
        response_frame = bytearray([0, 0, 0, 10]) + response_data
        response_encoded = base64.b64encode(response_frame)
        response_mock.body = response_encoded
        response_future = Future()
        response_future.set_result(response_mock)
        self.http_mock.fetch.return_value = response_future

        ctx = FContext()
        response_transport = yield transport_with_headers.request(ctx, request_frame)
        expected_headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
            'accept': 'application/x-frugal',
            'x-frugal-payload-limit': '200',
            'first-header': ctx.correlation_id,
            'second-header': 'test',
        }

        self.assertEqual(response_data, response_transport.getvalue())
        self.assertTrue(self.http_mock.fetch.called)
        request = self.http_mock.fetch.call_args[0][0]
        self.assertEqual(request.url, self.url)
        self.assertEqual(request.method, 'POST')
        self.assertEqual(request.body, base64.b64encode(request_frame))
        self.assertEqual(request.headers, expected_headers)

    @gen_test
    def test_request_too_much_data(self):
        self.transport._http = self.http_mock
        with self.assertRaises(TTransportException) as cm:
            yield self.transport.request(FContext(), bytearray([0] * 101))
        self.assertEqual(TTransportExceptionType.REQUEST_TOO_LARGE,
                         cm.exception.type)

    @gen_test
    def test_request_invalid_response_frame(self):
        self.transport._http = self.http_mock
        response_mock = mock.Mock(spec=HTTPResponse)
        response_mock.body = base64.b64encode(bytearray([4, 5]))
        response_future = Future()
        response_future.set_result(response_mock)
        self.http_mock.fetch.return_value = response_future

        with self.assertRaises(TTransportException):
            yield self.transport.request(
                FContext(), bytearray([0, 0, 0, 4, 1, 2, 3, 4]))

        self.assertTrue(self.http_mock.fetch.called)

    @gen_test
    def test_request_response_too_large(self):
        self.transport._http = self.http_mock

        self.http_mock.fetch.side_effect = HTTPError(code=413)

        with self.assertRaises(TTransportException) as cm:
            yield self.transport.request(
                FContext(), bytearray([0, 0, 0, 1, 0]))

        self.assertEqual(cm.exception.message, 'response was too large')

    @gen_test
    def test_request_response_error(self):
        self.transport._http = self.http_mock

        self.http_mock.fetch.side_effect = HTTPError(code=404)

        with self.assertRaises(TTransportException):
            yield self.transport.request(
                FContext(), bytearray([0, 0, 0, 1, 0]))

    @gen_test
    def test_request_timeout(self):
        self.transport._http = self.http_mock

        self.http_mock.fetch.side_effect = HTTPError(code=599)

        with self.assertRaises(TTransportException) as cm:
            yield self.transport.request(
                FContext(), bytearray([0, 0, 0, 1, 0]))
        self.assertEqual(
            TTransportExceptionType.TIMED_OUT, cm.exception.type)
        self.assertEqual("request timed out", cm.exception.message)

    @gen_test
    def test_request_service_unavailable(self):
        self.transport._http = self.http_mock

        self.http_mock.fetch.side_effect = socket.gaierror()

        with self.assertRaises(TTransportException) as e:
            yield self.transport.request(
                FContext(), bytearray([0, 0, 0, 3, 1, 2, 3])
            )

        self.assertEqual(
            str(e.exception),
            'service not available: None'
        )
