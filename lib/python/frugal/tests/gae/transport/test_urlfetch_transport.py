from base64 import b64encode
from struct import pack_into
import unittest

from mock import Mock
from mock import patch
from thrift.transport.TTransport import TTransportException

from frugal.gae.transport import TUrlfetchTransport


@patch('frugal.gae.transport.urlfetch_transport._urlfetch')
class TestFUrlfetchTransport(unittest.TestCase):

    def test_request(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'
        headers = {'foo': 'bar'}
        resp = Mock(status=200)
        response = b'response'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(response))
        resp_body = b64encode(buff + response)
        resp = Mock(status_code=200, content=resp_body)
        mock_urlfetch.return_value = resp

        def get_headers():
            return {'baz': 'qux'}

        tr = TUrlfetchTransport(url, headers=headers, get_headers=get_headers)
        deadline = 5
        tr.set_timeout(deadline * 1000)

        tr.open()
        self.assertTrue(tr.isOpen())

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_urlfetch.assert_called_once_with(
            url, encoded_frame, False, deadline,
            {'foo': 'bar', 'baz': 'qux', 'Content-Length': '20',
             'Content-Type': 'application/x-frugal',
             'Content-Transfer-Encoding': 'base64', 'User-Agent':
             'Python/TBaseHttpTransport'},
        )

        resp = tr.read(len(response))
        self.assertEqual(response, resp)

        tr.close()
        self.assertTrue(tr.isOpen())  # open/close are no-ops

    def test_request_https(self, mock_urlfetch):
        url = 'https://localhost:8080/frugal'
        resp = Mock(status=200)
        response = b'response'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(response))
        resp_body = b64encode(buff + response)
        resp = Mock(status_code=200, content=resp_body)
        mock_urlfetch.return_value = resp

        tr = TUrlfetchTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_urlfetch.assert_called_once_with(
            url, encoded_frame, True, None,
            {'Content-Length': '20', 'Content-Type': 'application/x-frugal',
             'Content-Transfer-Encoding': 'base64', 'User-Agent':
             'Python/TBaseHttpTransport'},
        )

        resp = tr.read(len(response))
        self.assertEqual(response, resp)

    def test_flush_no_body(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'

        tr = TUrlfetchTransport(url)
        tr.flush()

        self.assertFalse(mock_urlfetch.called)

    def test_flush_bad_response(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'
        resp = Mock(status_code=500)
        mock_urlfetch.return_value = resp

        tr = TUrlfetchTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)

        with self.assertRaises(TTransportException):
            tr.flush()

        mock_urlfetch.assert_called_once_with(
            url, encoded_frame, False, None,
            {'Content-Length': '20', 'Content-Type': 'application/x-frugal',
             'Content-Transfer-Encoding': 'base64', 'User-Agent':
             'Python/TBaseHttpTransport'},
        )

    def test_flush_bad_oneway_response(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'
        buff = bytearray(4)
        pack_into('!I', buff, 0, 10)
        resp_body = b64encode(buff)
        resp = Mock(status_code=200, content=resp_body)
        mock_urlfetch.return_value = resp

        tr = TUrlfetchTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)

        with self.assertRaises(TTransportException):
            tr.flush()

        mock_urlfetch.assert_called_once_with(
            url, encoded_frame, False, None,
            {'Content-Length': '20', 'Content-Type': 'application/x-frugal',
             'Content-Transfer-Encoding': 'base64', 'User-Agent':
             'Python/TBaseHttpTransport'},
        )

    def test_flush_oneway(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'
        buff = bytearray(4)
        pack_into('!I', buff, 0, 0)
        resp_body = b64encode(buff)
        resp = Mock(status_code=200, content=resp_body)
        mock_urlfetch.return_value = resp

        tr = TUrlfetchTransport(url)

        data = b'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_urlfetch.assert_called_once_with(
            url, encoded_frame, False, None,
            {'Content-Length': '20', 'Content-Type': 'application/x-frugal',
             'Content-Transfer-Encoding': 'base64', 'User-Agent':
             'Python/TBaseHttpTransport'},
        )

        resp = tr.read(10)
        self.assertEqual(b'', resp)
