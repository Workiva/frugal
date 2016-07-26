import base64
import httplib
import logging

from thrift.transport.TTransport import TTransportException
from tornado import gen
from tornado.httpclient import AsyncHTTPClient
from tornado.httpclient import HTTPError
from tornado.httpclient import HTTPRequest

from frugal.tornado.transport import FTornadoTransport

logger = logging.getLogger(__name__)


class FHttpTransport(FTornadoTransport):
    def __init__(self, url, request_capacity=0, response_capacity=0):
        """
        Create an HTTP transport.

        Args:
            url: The url to send requests to.
            request_capacity: The maximum size allowed to be written in a
                              request. Set to 0 for no size restrictions.
            response_capacity: The maximum size allowed to be read in a
                               response. Set to 0 for no size restrictions.
        """
        super(FHttpTransport, self).__init__(max_message_size=request_capacity)
        self._url = url
        self._http = AsyncHTTPClient()

        # create headers
        self._headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
            'accept': 'application/x-frugal',
        }
        if response_capacity > 0:
            self._headers['x-frugal-payload-limit'] = str(response_capacity)

        self._execute = None

    @gen.coroutine
    def isOpen(self):
        """Always returns True"""
        # Tornado requires we raise a special exception to return a value.
        raise gen.Return(True)

    @gen.coroutine
    def open(self):
        """No-op"""
        pass

    @gen.coroutine
    def close(self):
        """no-op"""
        pass

    @gen.coroutine
    def flush(self):
        """
        Write the current buffer and execute the set callback with the
        response.
        """
        frame = self.get_write_bytes()
        if not frame:
            return

        self.reset_write_buffer()
        encoded = base64.b64encode(frame)
        request = HTTPRequest(self._url, method='POST', body=encoded,
                              headers=self._headers)

        try:
            response = yield self._http.fetch(request)
        except HTTPError as e:
            if e.code == httplib.REQUEST_ENTITY_TOO_LARGE:
                raise TTransportException(type=TTransportException.UNKNOWN,
                                          message='response was too large')

            message = 'response errored with code {0} and body {1}'.format(
                e.code, e.message
            )
            raise TTransportException(type=TTransportException.UNKNOWN,
                                      message=message)

        decoded = base64.b64decode(response.body)

        if len(decoded) < 4:
            raise TTransportException(type=TTransportException.UNKNOWN,
                                      message='invalid frame size')

        if len(decoded) == 4:
            # One-way method, drop response
            return

        self.execute_frame(decoded)
