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

from base64 import b64encode
from io import BytesIO
from struct import pack_into

from thrift.transport.TTransport import TTransportException

from frugal.exceptions import TTransportExceptionType
from frugal.transport import TSynchronousTransport


class TBaseHttpTransport(TSynchronousTransport):
    """
    Base synchronous transport implemented with HTTP.
    """

    def __init__(self, url, request_capacity=0, response_capacity=0,
                 headers=None, get_headers=None):
        """
        Initialize a new FBaseHttpTransport.

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
        self._timeout = None

    def set_timeout(self, timeout):
        """
        Set the request timeout.

        Args:
            timeout: request timeout in milliseconds.
        """
        self._timeout = timeout

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
            raise TTransportException(
                type=TTransportExceptionType.REQUEST_TOO_LARGE,
                message='Message exceeds {0} bytes, was {1} bytes'.format(
                    self._request_capacity, size + 4))

        self._wbuff.write(buf)

    def _get_headers_and_body(self):
        """
        Return the request headers and body.

        Returns:
            headers dict and base64-encoded body string.
        """

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
            headers['User-Agent'] = 'Python/TBaseHttpTransport'

        return headers, body
