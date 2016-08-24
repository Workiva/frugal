from base64 import b64encode
from base64 import b64decode
from io import BytesIO
from struct import pack_into
from struct import unpack

from httplib2 import Http
from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FMessageSizeException
from frugal.transport import FSynchronousTransport


class FBaseHttpTransport(FSynchronousTransport):
    """Base synchronous transport implemented with HTTP."""

    def __init__(self, url, request_capacity=0, response_capacity=0,
                 headers=None, get_headers=None):
        """Initialize a new FBaseHttpTransport.

        Args:
            url: url of the Frugal server.
            request_capacity: max size allowed to be written in a request. Set
                              0 for no restriction.
            response_capacity: max size allowed to be read in a response. Set
                               0 for no restriction.
            headers: dict containing static headers.
            get_headers: func which returns dynamic headers per request.
        """

        self._url = url
        self._wbuff = BytesIO()
        self._rbuff = BytesIO()
        self._request_capacity = request_capacity
        self._response_capacity = response_capacity
        self._custom_headers = headers
        self._get_headers = get_headers

    def isOpen(self):
        return True

    def open(self):
        pass

    def close(self):
        pass

    def read(self, sz):
        return self._rbuff.read(sz)

    def write(self, buf):
        size = len(buf) + len(self._wbuff.getvalue())
        if size + 4 > self._request_capacity > 0:
            self._wbuff = BytesIO()
            raise FMessageSizeException('Message exceeds max message size')

        self._wbuff.write(buf)

    def _get_headers_and_body(self):
        """Return the request headers and body."""

        data = self._wbuff.getvalue()
        self._wbuff = BytesIO()
        frame_size = len(data)
        if frame_size == 0:
            return None, None

        # Prepend the frame size to the message.
        buff = bytearray(4)
        pack_into('!I', buff, 0, frame_size)

        body = b64encode(buff + data)

        headers = {
            'Content-Type': 'application/x-frugal',
            'Content-Length': str(len(body)),
            'Content-Transfer-Encoding': 'base64',
        }

        if self._response_capacity:
            headers['x-frugal-payload-limit'] = str(self._response_capacity)

        if self._get_headers:
            headers.update(self._get_headers())

        if self._custom_headers:
            for name, value in self._custom_headers.items():
                headers[name] = value

        if 'User-Agent' not in headers:
            headers['User-Agent'] = 'Python/FHttpTransport'

        return headers, body


class FHttpTransport(FBaseHttpTransport):
    """Synchronous transport implemented with httplib2."""

    def __init__(self, url, request_capacity=0, response_capacity=0,
                 headers=None, get_headers=None):
        """Initialize a new FHttpTransport.

        Args:
            url: url of the Frugal server.
            request_capacity: max size allowed to be written in a request. Set
                              0 for no restriction.
            response_capacity: max size allowed to be read in a response. Set
                               0 for no restriction.
            headers: dict containing static headers.
            get_headers: func which returns dynamic headers per request.
        """

        self._http = Http()
        super(FHttpTransport, self).__init__(
            url, request_capacity=request_capacity,
            response_capacity=response_capacity, headers=headers,
            get_headers=get_headers)

    def flush(self):
        headers, body = self._get_headers_and_body()

        if not body:
            return

        resp, resp_body = self._http.request(self._url, method='POST',
                                             body=body, headers=headers)
        if resp.status >= 400:
            raise TTransportException(
                TTransportException.UNKNOWN,
                'HTTP request failed, returned{0}: {1}'.format(resp.status,
                                                               resp.reason))

        resp_body = b64decode(resp_body)
        # All responses should be framed with 4 bytes (uint32).
        if len(resp_body) < 4:
            raise TTransportException(TTransportException.UNKNOWN,
                                      'invalid frame size')

        # If there are only 4 bytes, this needs to be a one-way (i.e. frame
        # size 0)
        if len(resp_body) == 4:
            if unpack('!I', resp_body)[0] != 0:
                raise TTransportException(TTransportException.UNKNOWN,
                                          'invalid frame')

            # It's a oneway, drop it.
            return

        self._rbuff = BytesIO(resp_body[4:])

