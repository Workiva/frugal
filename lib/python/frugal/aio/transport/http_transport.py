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

import asyncio
import base64

import async_timeout
from aiohttp.client import ClientSession
from thrift.transport.TTransport import TTransportBase
from thrift.transport.TTransport import TMemoryBuffer
from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FTransportBase
from frugal.context import FContext
from frugal.exceptions import TTransportExceptionType


class FHttpTransport(FTransportBase):
    """
    FHttpTransport is an FTransport that uses http as the underlying transport.
    This allows messages of arbitrary sizes to be sent and received.
    """
    def __init__(self, url, request_capacity=0, response_capacity=0):
        """
        Create an HTTP transport.

        Args:
            url: The url to send requests to.
            request_capacity: The maximum size allowed to be written in a
                              request. Set to 0 for no size restrictions.
            response_capacity: The maximum size allowed to be read in a
                               response. Set to 0 for no size restrictions
        """
        super().__init__(request_capacity)
        self._url = url

        self._headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
            'accept': 'application/x-frugal',
        }
        if response_capacity > 0:
            self._headers['x-frugal-payload-limit'] = str(response_capacity)

    def is_open(self):
        """Always returns True"""
        return True

    async def open(self):
        """No-op"""
        pass

    async def close(self):
        """No-op"""
        pass

    async def oneway(self, context: FContext, payload):
        """
        Write the current buffer. This transport detects oneway requests via
        via the payload size on the server response. Therefore, just call
        through to request.
        """
        await  self.request(context, payload)

    async def request(self, context: FContext, payload) -> TTransportBase:
        """
        Write the current buffer payload over the network and return the
        response.
        """
        self._preflight_request_check(payload)
        encoded = base64.b64encode(payload)

        status, text = await self._make_request(context, encoded)
        if status == 413:
            raise TTransportException(
                type=TTransportExceptionType.RESPONSE_TOO_LARGE,
                message='response was too large for the transport'
            )

        if status >= 300:
            raise TTransportException(
                type=TTransportExceptionType.UNKNOWN,
                message='request errored with code {0} and message {1}'.format(
                    status, str(text)
                )
            )

        decoded = base64.b64decode(text)
        if len(decoded) < 4:
            raise TTransportException(type=TTransportExceptionType.UNKNOWN,
                                      message='invalid frame size')

        if len(decoded) == 4:
            if any(decoded):
                raise TTransportException(type=TTransportExceptionType.UNKNOWN,
                                          message='missing data')
            # One-way method, drop response
            return

        return TMemoryBuffer(decoded[4:])

    async def _make_request(self, context: FContext, payload):
        """
        Helper method to make a request over the network.

        Args:
            payload: The data to be sent over the network.
        Return:
            The status code and body of the response.
        Throws:
            TTransportException if the request timed out.
        """
        with ClientSession() as session:
            try:
                with async_timeout.timeout(context.timeout / 1000):
                    async with session.post(self._url,
                                            data=payload,
                                            headers=self._headers) as response:
                        return response.status, await response.content.read()
            except asyncio.TimeoutError:
                raise TTransportException(
                    type=TTransportExceptionType.TIMED_OUT,
                    message='request timed out'
                )
