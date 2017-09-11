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

import base64
import httplib
import logging

from thrift.transport.TTransport import TMemoryBuffer
from thrift.transport.TTransport import TTransportException
from tornado import gen
from tornado.httpclient import AsyncHTTPClient
from tornado.httpclient import HTTPError
from tornado.httpclient import HTTPRequest

from frugal.exceptions import TTransportExceptionType
from frugal.tornado.transport.transport import FTransportBase

logger = logging.getLogger(__name__)


class FHttpTransport(FTransportBase):
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
        super(FHttpTransport, self).__init__(
            request_size_limit=request_capacity)
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
    def is_open(self):
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
    def oneway(self, context, payload):
        """
        Write the current buffer. This transport detects oneway requests via
        via the payload size on the server response. Therefore, just call
        through to request.
        """
        yield self.request(context, payload)

    @gen.coroutine
    def request(self, context, payload):
        """
        Write the current buffer and return the response.
        """
        self._preflight_request_check(payload)
        encoded = base64.b64encode(payload)
        request = HTTPRequest(self._url,
                              method='POST',
                              body=encoded,
                              headers=self._headers,
                              request_timeout=context.timeout / 1000.0
                              )

        try:
            response = yield self._http.fetch(request)
        except HTTPError as e:
            if e.code == httplib.REQUEST_ENTITY_TOO_LARGE:
                raise TTransportException(
                    type=TTransportExceptionType.RESPONSE_TOO_LARGE,
                    message='response was too large')

            # Tornado HttpClient uses 599 as the HTTP code to indicate a
            # request timeout
            if e.code == 599:
                raise TTransportException(
                    type=TTransportExceptionType.TIMED_OUT,
                    message='request timed out')

            message = 'response errored with code {0} and body {1}'.format(
                e.code, e.message
            )
            raise TTransportException(
                type=TTransportExceptionType.UNKNOWN,
                message=message)

        decoded = base64.b64decode(response.body)

        if len(decoded) < 4:
            raise TTransportException(
                type=TTransportExceptionType.UNKNOWN,
                message='invalid frame size')

        if len(decoded) == 4:
            # One-way method, drop response
            return

        raise gen.Return(TMemoryBuffer(decoded[4:]))
