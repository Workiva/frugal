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

import mock
from threading import Thread
import time
import unittest

from thrift.protocol.TJSONProtocol import TJSONProtocolFactory
from thrift.transport.TSocket import TSocket
from thrift.transport.TSocket import TServerSocket

from frugal.protocol import FProtocolFactory
from frugal.server import FSimpleServer


class TestSimpleServer(unittest.TestCase):
    def test_it_works(self):
        processor_factory = mock.Mock()
        mock_processor = mock.Mock()
        processor_factory.get_processor.return_value = mock_processor
        proto_factory = FProtocolFactory(TJSONProtocolFactory())
        server_trans = TServerSocket(host='localhost', port=5536)
        server = FSimpleServer(processor_factory, server_trans, proto_factory)

        thread = Thread(target=lambda: server.serve())
        thread.start()
        time.sleep(0.1)

        transport = TSocket(host='localhost', port=5536)
        transport.open()
        transport.write(bytearray([0, 0, 0, 3, 1, 2, 3]))
        transport.flush()
        time.sleep(0.1)

        server.stop()
        processor_factory.get_processor.assert_called_once_with(mock.ANY)
        mock_processor.process.assert_called_with(mock.ANY, mock.ANY)
