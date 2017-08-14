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

import unittest

from thrift.transport.TTransport import TTransportException

from frugal.exceptions import TTransportExceptionType
from frugal.transport import TMemoryOutputBuffer


class TestTOutputMemoryBuffer(unittest.TestCase):

    def setUp(self):
        super(TestTOutputMemoryBuffer, self).setUp()

        self.buffer = TMemoryOutputBuffer(10)

    def test_write(self):
        self.buffer.write(b'foo')
        self.assertSequenceEqual(b'\x00\x00\x00\x03foo', self.buffer.getvalue())

    def test_write_size_exception(self):
        self.assertEqual(4, len(self.buffer))
        self.buffer.write(bytearray(6))
        self.assertEqual(10, len(self.buffer))
        with self.assertRaises(TTransportException) as cm:
            self.buffer.write(bytearray(1))
        self.assertEqual(TTransportExceptionType.REQUEST_TOO_LARGE,
                         cm.exception.type)
        self.assertEqual(4, len(self.buffer))

    def test_len(self):
        self.buffer.write(b'12345')
        self.assertEqual(len(self.buffer), 9)
