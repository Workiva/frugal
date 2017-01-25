from base64 import b64decode
from io import BytesIO
from struct import unpack

from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FTimeoutException
from frugal.transport.base_http_transport import TBaseHttpTransport


class TUrlfetchTransport(TBaseHttpTransport):
    """Synchronous transport implemented with urlfetch."""

    def __init__(self, url, headers=None, get_headers=None):
        """Initialize a new FUrlfetchTransport.

        Args:
            url: url of the Frugal server.
            headers: dict containing static headers.
            get_headers: func which returns dynamic headers per request.
        """

        self._timeout = None
        super(TUrlfetchTransport, self).__init__(url, headers=headers,
                                                 get_headers=get_headers)

    def set_timeout(self, timeout):
        """Set the request timeout.

        Args:
            timeout: request timeout in milliseconds.
        """

        self._timeout = timeout / 1000

    def flush(self):
        headers, body = self._get_headers_and_body()

        if not body:
            return

        resp = _urlfetch(self._url, body, self._url.startswith('https://'),
                         self._timeout, headers)

        if resp.status_code >= 400:
            raise TTransportException(
                TTransportException.UNKNOWN,
                'urlfetch request failed, returned {0}'.format(
                    resp.status_code))

        resp_body = b64decode(resp.content)
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


def _urlfetch(url, body, validate_certificate, timeout, headers):
    from google.appengine.api import urlfetch
    from google.appengine.api.urlfetch_errors import DeadlineExceededError
    try:
        return urlfetch.fetch(
            url, method=urlfetch.POST, payload=body, headers=headers,
            validate_certificate=url.startswith('https://'),
            deadline=timeout
        )
    except DeadlineExceededError:
        raise FTimeoutException()

