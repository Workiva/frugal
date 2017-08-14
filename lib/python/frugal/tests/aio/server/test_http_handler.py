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
import mock

from aiohttp import test_utils, CIMultiDict
from aiohttp.streams import StreamReader
from thrift.protocol import TBinaryProtocol

from frugal.protocol import FProtocolFactory
from frugal.aio.server import new_http_handler
from frugal.tests.aio import utils


class TestFHttpHandler(utils.AsyncIOTestCase):
    def setUp(self):
        self.processor = mock.Mock()

        prot_factory = FProtocolFactory(
                TBinaryProtocol.TBinaryProtocolFactory())
        self.handler = new_http_handler(self.processor, prot_factory)
        super().setUp()

    @utils.async_runner
    async def test_basic(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9])
        response_frame = bytearray([0, 0, 0, 4]) + response_data
        request_payload_reader = StreamReader()
        request_payload_reader.feed_data(request_payload)
        request_payload_reader.feed_eof()

        request = test_utils.make_mocked_request('POST', '/frugal',
                                                 payload=request_payload_reader)

        async def process_data(_, oprot):
            oprot.get_transport().write(response_data)
        self.processor.process.side_effect = process_data

        response = await self.handler(request)
        self.assertEqual(200, response.status)
        self.assertTrue(self.processor.process.called)
        iprot, _ = self.processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())
        received_payload = base64.b64decode(response.text)

        self.assertEqual(response_frame, received_payload)
        self.assertEqual('application/x-frugal',
                         response.headers['content-type'])
        self.assertEqual('base64',
                         response.headers['content-transfer-encoding'])

    @utils.async_runner
    async def test_response_too_large(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9, 10, 11])
        request_payload_reader = StreamReader()
        request_payload_reader.feed_data(request_payload)
        request_payload_reader.feed_eof()

        headers = CIMultiDict({
            'x-frugal-payload-limit': '5',
        })
        request = test_utils.make_mocked_request('POST', '/frugal',
                                                 payload=request_payload_reader,
                                                 headers=headers)

        async def process_data(_, oprot):
            oprot.get_transport().write(response_data)
        self.processor.process.side_effect = process_data

        response = await self.handler(request)
        self.assertEqual(413, response.status)
        self.assertTrue(self.processor.process.called)
        iprot, _ = self.processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())
