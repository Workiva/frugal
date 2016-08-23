from base64 import b64encode
from struct import pack_into
import unittest

from mock import Mock
from mock import patch
from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FMessageSizeException
from frugal.transport.http_transport import FHttpTransport


@patch('frugal.transport.http_transport.Http')
class TestFHttpTransport(unittest.TestCase):

    def test_request(self, mock_http):
        url = 'http://localhost:8080/frugal'
        headers = {'foo': 'bar'}
        mock_http_client = mock_http.return_value
        resp = Mock(status=200)
        response = b'response'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(response))
        resp_body = b64encode(buff + response)
        mock_http_client.request.return_value = (resp, resp_body)

        def get_headers():
            return {'baz': 'qux'}

        tr = FHttpTransport(url, headers=headers, get_headers=get_headers,
                            response_capacity=500)

        tr.open()
        self.assertTrue(tr.isOpen())

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_http_client.request.assert_called_once_with(
            url, method='POST', body=encoded_frame,
            headers={'foo': 'bar', 'baz': 'qux', 'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport',
                     'x-frugal-payload-limit': '500'})

        resp = tr.read(len(response))
        self.assertEqual(response, resp)

        tr.close()
        self.assertTrue(tr.isOpen())  # open/close are no-ops

    def test_flush_no_body(self, mock_http):
        url = 'http://localhost:8080/frugal'
        mock_http_client = mock_http.return_value

        tr = FHttpTransport(url)
        tr.flush()

        self.assertFalse(mock_http_client.request.called)

    def test_flush_bad_response(self, mock_http):
        url = 'http://localhost:8080/frugal'
        mock_http_client = mock_http.return_value
        resp = Mock(status=500)
        mock_http_client.request.return_value = (resp, None)

        tr = FHttpTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)

        with self.assertRaises(TTransportException):
            tr.flush()

        mock_http_client.request.assert_called_once_with(
            url, method='POST', body=encoded_frame,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport'})

    def test_flush_bad_oneway_response(self, mock_http):
        url = 'http://localhost:8080/frugal'
        mock_http_client = mock_http.return_value
        resp = Mock(status=200)
        buff = bytearray(4)
        pack_into('!I', buff, 0, 10)
        resp_body = b64encode(buff)
        mock_http_client.request.return_value = (resp, resp_body)

        tr = FHttpTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)

        with self.assertRaises(TTransportException):
            tr.flush()

        mock_http_client.request.assert_called_once_with(
            url, method='POST', body=encoded_frame,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport'})

    def test_flush_oneway(self, mock_http):
        url = 'http://localhost:8080/frugal'
        mock_http_client = mock_http.return_value
        resp = Mock(status=200)
        buff = bytearray(4)
        pack_into('!I', buff, 0, 0)
        resp_body = b64encode(buff)
        mock_http_client.request.return_value = (resp, resp_body)

        tr = FHttpTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_http_client.request.assert_called_once_with(
            url, method='POST', body=encoded_frame,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport'})

        resp = tr.read(10)
        self.assertEqual(b'', resp)

    def test_write_limit_exceeded(self, mock_http):
        url = 'http://localhost:8080/frugal'
        mock_http_client = mock_http.return_value
        resp = Mock(status=200)
        buff = bytearray(4)
        pack_into('!I', buff, 0, 0)
        resp_body = b64encode(buff)
        mock_http_client.request.return_value = (resp, resp_body)

        tr = FHttpTransport(url, request_capacity=5)

        data = b'helloworld'

        with self.assertRaises(FMessageSizeException):
            tr.write(data)

        self.assertFalse(mock_http_client.request.called)
