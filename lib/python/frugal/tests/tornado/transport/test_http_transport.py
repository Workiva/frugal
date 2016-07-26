import base64

import mock
from thrift.transport.TTransport import TTransportException
from tornado.concurrent import Future
from tornado.httpclient import AsyncHTTPClient
from tornado.httpclient import HTTPError
from tornado.httpclient import HTTPResponse
from tornado.testing import gen_test, AsyncTestCase

from frugal.exceptions import FMessageSizeException
from frugal.tornado.transport.http_transport import FHttpTransport


class TestFHttpTransport(AsyncTestCase):
    def setUp(self):
        super(TestFHttpTransport, self).setUp()

        self.url = 'http://localhost/testing'
        self.request_capacity = 100
        self.response_capacity = 200
        self.transport = FHttpTransport(
            self.url, request_capacity=self.request_capacity,
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
        self.assertTrue((yield self.transport.isOpen()))
        yield self.transport.open()
        self.assertTrue((yield self.transport.isOpen()))
        self.assertIsNotNone(self.transport._http)
        yield self.transport.close()
        self.assertTrue((yield self.transport.isOpen()))
        self.assertIsNotNone(self.transport._http)

    @gen_test
    def test_write_too_much_data(self):
        self.transport._http = self.http_mock
        with self.assertRaises(FMessageSizeException):
            yield self.transport.write(bytearray([0] * 101))

    @gen_test
    def test_flush_success(self):
        registry_mock = mock.Mock()
        self.transport.set_registry(registry_mock)
        self.transport._http = self.http_mock

        request_data = bytearray([4, 5, 6, 8, 9, 10, 11, 13, 12, 3])
        expected_payload = bytearray([0, 0, 0, 10]) + request_data

        response_mock = mock.Mock(spec=HTTPResponse)
        response_data = bytearray([23, 24, 25, 26, 27, 28, 29])
        response_frame = bytearray([0, 0, 0, 10]) + response_data
        response_encoded = base64.b64encode(response_frame)
        response_mock.body = response_encoded
        response_future = Future()
        response_future.set_result(response_mock)
        self.http_mock.fetch.return_value = response_future

        yield self.transport.write(request_data[:3])
        yield self.transport.write(request_data[3:7])
        yield self.transport.write(request_data[7:])
        yield self.transport.flush()

        self.assertTrue(self.http_mock.fetch.called)
        request = self.http_mock.fetch.call_args[0][0]
        self.assertEqual(request.url, self.url)
        self.assertEqual(request.method, 'POST')
        self.assertEqual(request.body, base64.b64encode(expected_payload))
        self.assertEqual(request.headers, self.headers)

        registry_mock.execute.assert_called_once_with(response_data)

    @gen_test
    def test_flush_invalid_response_frame(self):
        self.transport._http = self.http_mock
        response_mock = mock.Mock(spec=HTTPResponse)
        response_mock.body = base64.b64encode(bytearray([4, 5]))
        response_future = Future()
        response_future.set_result(response_mock)
        self.http_mock.fetch.return_value = response_future

        yield self.transport.write(bytearray([1, 2, 3, 4]))

        with self.assertRaises(TTransportException):
            yield self.transport.flush()

        self.assertTrue(self.http_mock.fetch.called)

    @gen_test
    def test_flush_oneway(self):
        callback_mock = mock.Mock()
        self.transport.set_execute_callback(callback_mock)
        self.transport._http = self.http_mock

        response_encoded = base64.b64encode(bytearray([0, 0, 0, 0]))
        response_mock = mock.Mock(spec=HTTPResponse)
        response_mock.body = response_encoded
        response_future = Future()
        response_future.set_result(response_mock)
        self.http_mock.fetch.return_value = response_future

        yield self.transport.write(bytearray([1, 2, 3]))
        yield self.transport.flush()

        self.assertTrue(self.http_mock.fetch.called)
        self.assertFalse(callback_mock.called)

    @gen_test
    def test_flush_response_too_large(self):
        self.transport._http = self.http_mock

        self.http_mock.fetch.side_effect = HTTPError(code=413)
        yield self.transport.write(bytearray([0]))

        with self.assertRaises(TTransportException) as e:
            yield self.transport.flush()
            self.assertEqual(e.message, 'response was too large')

    @gen_test
    def test_flush_response_error(self):
        self.transport._http = self.http_mock

        self.http_mock.fetch.side_effect = HTTPError(code=404)
        yield self.transport.write(bytearray([0]))

        with self.assertRaises(TTransportException):
            yield self.transport.flush()
