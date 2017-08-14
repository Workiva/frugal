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

from base64 import b64decode
from io import BytesIO
from struct import unpack

from thrift.transport.TTransport import TTransportException

from frugal.exceptions import TTransportExceptionType
from frugal.transport.base_http_transport import TBaseHttpTransport


class TUrlfetchTransport(TBaseHttpTransport):
    """
    Synchronous transport implemented with urlfetch.
    """

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
        """
        Set the request timeout.

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

        code = resp.status_code
        if code >= 400:
            msg = 'urlfetch request failed, returned {0}'.format(code)
            raise TTransportException(TTransportExceptionType.UNKNOWN, msg)

        resp_body = b64decode(resp.content)
        # All responses should be framed with 4 bytes (uint32).
        if len(resp_body) < 4:
            msg = 'Invalid frame size.'
            raise TTransportException(TTransportExceptionType.UNKNOWN, msg)

        # If there are only 4 bytes, this needs to be a one-way (i.e. frame
        # size 0)
        if len(resp_body) == 4:
            if unpack('!I', resp_body)[0] != 0:
                msg = 'invalid frame'
                raise TTransportException(TTransportExceptionType.UNKNOWN, msg)

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
        raise TTransportException(type=TTransportExceptionType.TIMED_OUT)
