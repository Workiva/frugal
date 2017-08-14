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
import logging
import struct

from thrift.transport.TTransport import TMemoryBuffer

logger = logging.getLogger(__name__)


class _FHttpException(Exception):
    def __init__(self, code, message=None):
        super(_FHttpException, self).__init__(message)
        self.code = code


class _FHttpRequest(object):
    """
    _FHttpRequest stores data from an http request in a generic format.
    """
    def __init__(self, headers=None, body=b''):
        if headers is None:
            headers = {}
        self.headers = headers
        self.body = body


class _FHttpResponse(object):
    """
    _FHttpResponse returns data to be sent in an http response in a generic
    format.
    """
    def __init__(self, status_code=200, headers=None, body=b''):
        self._status_code = status_code
        if headers is None:
            headers = {}
        self._headers = headers
        self._body = body

    @property
    def status_code(self):
        """
        Status code of the response as an integer.
        """
        return self._status_code

    @property
    def headers(self):
        """
        dict of the HTTP response headers.
        """
        return self._headers

    @property
    def body(self):
        """
        The http response body.
        """
        return self._body


class _FHttpRequestHandler(object):
    """
    _FHttpRequestHandler provides functionality to process rpcs from http.

    Reading from/writing to the network of a library of choice is left up to
    the caller.
    """
    def __init__(self, processor, protocol_factory):
        """
        Initializes a _FHttpRequestHandler.

        Args:
            processor: The processor to use to handle requests.
            protocol_factory: A protocol factory to serialize/deserialize
                              frugal requests.
        """
        self._processor = processor
        self._protocol_factory = protocol_factory

    def _preprocess_http_request(self, request):
        """
        Performs some common preprocessing on an http request.

        Args:
            request: A _FHttpRequest object.
        Returns:
            otrans: The output transport, used for getting the output data
            iprot: The input protocol, given to a processor.
            oprot: The output protocol, given to a processor.
            response_limit: The maximum allowed size of the response,
                            zero if unlimited.
        """
        response_limit = int(request.headers.get('x-frugal-payload-limit', 0))
        payload = base64.b64decode(request.body)

        # Need 4 bytes for the frame size, at a minimum.
        if len(payload) < 4:
            logger.exception("invalid request size %s", len(payload))
            raise _FHttpException(400)

        # Ensure expected frame size equals actual size.
        size = struct.unpack('!I', payload[:4])[0]
        length = len(payload) - 4
        if size != length:
            raise _FHttpException(
                400, message='Mismatch between expected frame ' +
                'size ({}) and actual size ({})'.format(size, length))

        itrans = TMemoryBuffer(payload[4:])
        otrans = TMemoryBuffer()
        iprot = self._protocol_factory.get_protocol(itrans)
        oprot = self._protocol_factory.get_protocol(otrans)
        return otrans, iprot, oprot, response_limit

    def _postprocess_http_request(self, otrans, response_limit):
        """
        Performs some common postprocessing to produce an http response.

        Args:
            otrans: The output transport.
            response_limit: The maximum allowed size of the response,
                            zero if unlimited.
        Returns:
            A _FHttpResponse to write back to the client.
        """
        output_data = otrans.getvalue()
        if len(output_data) > response_limit > 0:
            logger.exception('response limit exceeded')
            return _FHttpResponse(status_code=413)

        frame_len = struct.pack('!I', len(output_data))
        frame = base64.b64encode(frame_len + output_data)

        headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
        }
        return _FHttpResponse(headers=headers, body=frame)

    def _handle_processor_exception(self, ex):
        """
        Handles an unexpected exception from a processor.

        Args:
            ex: The exception.
        Returns:
            A _FHttpResponse.
        """
        return _FHttpResponse(status_code=500, body=ex.message)

    def handle_http_request(self, request):
        """
        Handles an http rpc. This must be implemented because of differences
        in processors (synchronous v tornado coroutine v asyncio etc.). It
        is recommended to use _preprocess_http_request and
        _postprocess_http_request to make this easier. If those function are
        used correctly, only calling the processor and handling processor
        errors should be necessary for implementors.

        Args:
            request: A _FHttpRequest object.
        Returns:
            a _FHttpResponse object.
        """
        raise NotImplementedError()


class _FSynchronousHttpRequestHandler(_FHttpRequestHandler):
    """
    An http request handler for synchronous processors.
    """

    def handle_http_request(self, request):
        """
        Handle a given http request

        Args:
            request - http request to handle
        Returns:
            _FHttpResponse
        """
        try:
            otrans, iprot, oprot, response_limit = \
                self._preprocess_http_request(request)
        except _FHttpException as e:
            return _FHttpResponse(status_code=e.code)

        try:
            self._processor.process(iprot, oprot)
        except Exception as e:
            return self._handle_processor_exception(e)

        return self._postprocess_http_request(otrans, response_limit)
