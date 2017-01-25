from base64 import b64encode
from struct import pack_into
import unittest

from mock import Mock
from mock import patch
from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FMessageSizeException
from frugal.transport.http_transport import THttpTransport


@patch('frugal.transport.http_transport.requests')
class TestTHttpTransport(unittest.TestCase):

    def test_request(self, mock_requests):
        url = 'http://localhost:8080/frugal'
        headers = {'foo': 'bar'}
        resp = Mock(status_code=200)
        response = b'response'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(response))
        resp.content = b64encode(buff + response)
        mock_requests.post.return_value = resp

        def get_headers():
            return {'baz': 'qux'}

        tr = THttpTransport(url, headers=headers, get_headers=get_headers,
                            response_capacity=500)

        tr.open()
        self.assertTrue(tr.isOpen())

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_requests.post.assert_called_once_with(
            url, data=encoded_frame, timeout=None,
            headers={'foo': 'bar', 'baz': 'qux', 'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/TBaseHttpTransport',
                     'x-frugal-payload-limit': '500'})

        resp = tr.read(len(response))
        self.assertEqual(response, resp)

        tr.close()
        self.assertTrue(tr.isOpen())  # open/close are no-ops

    def test_request_timeout(self, mock_requests):
        url = 'http://localhost:8080/frugal'
        headers = {'foo': 'bar'}
        resp = Mock(status_code=200)
        response = b'response'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(response))
        resp.content = b64encode(buff + response)
        mock_requests.post.return_value = resp

        def get_headers():
            return {'baz': 'qux'}

        tr = THttpTransport(url, headers=headers, get_headers=get_headers,
                            response_capacity=500)

        tr.open()
        self.assertTrue(tr.isOpen())

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.set_timeout(5000)
        tr.write(data)
        tr.flush()

        mock_requests.post.assert_called_once_with(
            url, data=encoded_frame, timeout=5,
            headers={'foo': 'bar', 'baz': 'qux', 'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/TBaseHttpTransport',
                     'x-frugal-payload-limit': '500'})

        resp = tr.read(len(response))
        self.assertEqual(response, resp)

        tr.close()
        self.assertTrue(tr.isOpen())  # open/close are no-ops

    def test_flush_no_body(self, mock_requests):
        url = 'http://localhost:8080/frugal'

        tr = THttpTransport(url)
        tr.flush()

        self.assertFalse(mock_requests.post.called)

    def test_flush_bad_response(self, mock_requests):
        url = 'http://localhost:8080/frugal'
        resp = Mock(status_code=500)
        mock_requests.post.return_value = resp

        tr = THttpTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)

        with self.assertRaises(TTransportException):
            tr.flush()

        mock_requests.post.assert_called_once_with(
            url, data=encoded_frame, timeout=None,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/TBaseHttpTransport'})

    def test_flush_bad_oneway_response(self, mock_requests):
        url = 'http://localhost:8080/frugal'
        resp = Mock(status_code=200)
        buff = bytearray(4)
        pack_into('!I', buff, 0, 10)
        resp.content = b64encode(buff)
        mock_requests.post.return_value = resp

        tr = THttpTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)

        with self.assertRaises(TTransportException):
            tr.flush()

        mock_requests.post.assert_called_once_with(
            url, data=encoded_frame, timeout=None,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/TBaseHttpTransport'})

    def test_flush_oneway(self, mock_requests):
        url = 'http://localhost:8080/frugal'
        resp = Mock(status_code=200)
        buff = bytearray(4)
        pack_into('!I', buff, 0, 0)
        resp.content = b64encode(buff)
        mock_requests.post.return_value = resp

        tr = THttpTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_requests.post.assert_called_once_with(
            url, data=encoded_frame, timeout=None,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/TBaseHttpTransport'})

        resp = tr.read(10)
        self.assertEqual(b'', resp)

    def test_write_limit_exceeded(self, mock_requests):
        url = 'http://localhost:8080/frugal'
        resp = Mock(status_code=200)
        buff = bytearray(4)
        pack_into('!I', buff, 0, 0)
        resp.content = b64encode(buff)
        mock_requests.post.return_value = resp

        tr = THttpTransport(url, request_capacity=5)

        data = b'helloworld'

        with self.assertRaises(FMessageSizeException):
            tr.write(data)

        self.assertFalse(mock_requests.post.called)
