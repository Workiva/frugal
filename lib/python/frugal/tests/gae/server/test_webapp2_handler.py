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
import unittest

import mock
from thrift.protocol import TBinaryProtocol
import webapp2
import webtest

from frugal.protocol import FProtocolFactory
from frugal.gae.server.webapp2_handler import new_webapp2_handler


class FWebapp2HttpHandlerTest(unittest.TestCase):
    def setUp(self):
        self.mock_processor = mock.Mock()
        prot_factory = FProtocolFactory(
            TBinaryProtocol.TBinaryProtocolFactory())
        app = webapp2.WSGIApplication([('/frugal', new_webapp2_handler(
            self.mock_processor, prot_factory))])
        self.test_app = webtest.TestApp(app)

        self.request_data = bytearray([2, 3, 4])
        self.request_frame = bytearray([0, 0, 0, 3]) + self.request_data
        self.request_payload = base64.b64encode(self.request_frame)
        self.response_data = bytearray([6, 7, 8, 9, 10, 11])
        self.response_frame = bytearray([0, 0, 0, 6]) + self.response_data

        def process_data(_, oprot):
            oprot.get_transport().write(self.response_data)
        self.mock_processor.process.side_effect = process_data

    def test_basic(self):
        response = self.test_app.post('/frugal', params=self.request_payload)
        self.assertEqual(200, response.status_int)

        self.assertTrue(self.mock_processor.process.called)
        iprot, _ = self.mock_processor.process.call_args[0]
        self.assertEqual(self.request_data, iprot.get_transport().getvalue())

        expected_response_payload = base64.b64encode(self.response_frame)

        self.assertEqual(expected_response_payload, response.normal_body)
        self.assertEqual('application/x-frugal',
                         response.headers['content-type'])
        self.assertEqual('base64',
                         response.headers['content-transfer-encoding'])

    def test_response_too_large(self):
        headers = {
            'x-frugal-payload-limit': '5',
        }
        response = self.test_app.post('/frugal', params=self.request_payload,
                                      headers=headers, status='*')
        self.assertEqual(413, response.status_int)
        self.assertTrue(self.mock_processor.process.called)
        iprot, _ = self.mock_processor.process.call_args[0]
        self.assertEqual(self.request_data, iprot.get_transport().getvalue())

    def test_request_too_short(self):
        request_frame = base64.b64encode(bytearray([0]))
        response = self.test_app.post('/frugal', params=request_frame,
                                      status='*')

        self.assertEqual(400, response.status_int)

    def test_frame_size_mismatch(self):
        request_frame = base64.b64encode(bytearray([0, 0, 0, 10, 1, 1, 2]))

        response = self.test_app.post('/frugal', params=request_frame,
                                      status='*')
        self.assertEqual(400, response.status_int)

