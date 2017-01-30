from base64 import b64decode
from io import BytesIO
from struct import unpack

import requests
from requests.exceptions import ReadTimeout
from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FrugalTTransportExceptionType
from frugal.transport.base_http_transport import TBaseHttpTransport


class THttpTransport(TBaseHttpTransport):
    """Synchronous transport implemented with Requests."""

    def __init__(self, url, request_capacity=0, response_capacity=0,
                 headers=None, get_headers=None):
        """Initialize a new THttpTransport.

        Args:
            url: url of the Frugal server.
            request_capacity: max size allowed to be written in a request. Set
                              0 for no restriction.
            response_capacity: max size allowed to be read in a response. Set
                               0 for no restriction.
            headers: dict containing static headers.
            get_headers: func which returns dynamic headers per request.
        """

        super(THttpTransport, self).__init__(
            url, request_capacity=request_capacity,
            response_capacity=response_capacity, headers=headers,
            get_headers=get_headers)

    def flush(self):
        headers, body = self._get_headers_and_body()

        if not body:
            return

        timeout = None
        if self._timeout:
            # Requests uses timeout in seconds.
            timeout = self._timeout / 1000.0
            if timeout <= 0:
                timeout = None

        try:
            resp = requests.post(self._url, data=body, headers=headers,
                                 timeout=timeout)
        except ReadTimeout:
            raise TTransportException(
                type=FrugalTTransportExceptionType.TIMED_OUT,
                message='Request timed out')
        if resp.status_code >= 400:
            raise TTransportException(
                type=FrugalTTransportExceptionType.UNKNOWN,
                message='HTTP request failed, returned {0}: {1}'.format(
                    resp.status_code, resp.reason))

        resp_body = b64decode(resp.content)
        # All responses should be framed with 4 bytes (uint32).
        if len(resp_body) < 4:
            raise TTransportException(
                type=FrugalTTransportExceptionType.UNKNOWN,
                message='invalid frame size')

        # If there are only 4 bytes, this needs to be a one-way (i.e. frame
        # size 0)
        if len(resp_body) == 4:
            if unpack('!I', resp_body)[0] != 0:
                raise TTransportException(
                    type=FrugalTTransportExceptionType.UNKNOWN,
                    message='invalid frame')

            # It's a oneway, drop it.
            return

        self._rbuff = BytesIO(resp_body[4:])

