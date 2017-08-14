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

from thrift.protocol import TBinaryProtocol
from tornado import gen
from tornado.testing import AsyncHTTPTestCase
from tornado.web import Application

from frugal.protocol import FProtocolFactory
from frugal.tornado.server import FHttpHandler


class TestFTornadoHTTPHandler(AsyncHTTPTestCase):

    def setUp(self):
        self.processor = mock.Mock()
        super(TestFTornadoHTTPHandler, self).setUp()

    def get_app(self):
        prot_factory = FProtocolFactory(
                TBinaryProtocol.TBinaryProtocolFactory())
        return Application([
            ('/frugal', FHttpHandler, {
                'processor': self.processor,
                'protocol_factory': prot_factory,
            })
        ])

    def test_basic(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9])
        response_frame = bytearray([0, 0, 0, 4]) + response_data

        @gen.coroutine
        def process_data(_, oprot):
            oprot.get_transport().write(response_data)
        self.processor.process.side_effect = process_data

        response = self.fetch('/frugal', method='POST', body=request_payload)
        self.assertEqual(200, response.code)
        self.assertTrue(self.processor.process.called)
        iprot, _ = self.processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())

        expected_response_payload = base64.b64encode(response_frame)

        self.assertEqual(expected_response_payload, response.body)
        self.assertEqual('application/x-frugal',
                         response.headers['content-type'])
        self.assertEqual('base64',
                         response.headers['content-transfer-encoding'])

    def test_async(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9])
        response_frame = bytearray([0, 0, 0, 4]) + response_data

        @gen.coroutine
        def process_data(_, oprot):
            yield gen.moment
            oprot.get_transport().write(response_data)
        self.processor.process.side_effect = process_data

        response = self.fetch('/frugal', method='POST', body=request_payload)
        self.assertEqual(200, response.code)
        self.assertTrue(self.processor.process.called)
        iprot, _ = self.processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())

        expected_response_payload = base64.b64encode(response_frame)

        self.assertEqual(expected_response_payload, response.body)
        self.assertEqual('application/x-frugal',
                         response.headers['content-type'])
        self.assertEqual('base64',
                         response.headers['content-transfer-encoding'])

    def test_response_too_large(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9, 10, 11])

        @gen.coroutine
        def process_data(_, oprot):
            oprot.get_transport().write(response_data)
        self.processor.process.side_effect = process_data

        headers = {
            'x-frugal-payload-limit': '5',
        }
        response = self.fetch('/frugal', method='POST', body=request_payload,
                              headers=headers)
        self.assertEqual(413, response.code)
        self.assertTrue(self.processor.process.called)
        iprot, _ = self.processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())
