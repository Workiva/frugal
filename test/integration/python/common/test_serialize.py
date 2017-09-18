# -*- coding: utf-8 -*-
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

from thrift.protocol import TBinaryProtocol
from thrift.protocol import TJSONProtocol

from frugal.serialize import deserialize
from frugal.serialize import serialize

from frugal_test.f_FrugalTest import Bonk

class TestSerialize(unittest.TestCase):
    def test_serialize_writes(self):
        """
        Ensure serialization will work with
        non-ASCII unicode strings.
        """
        bonk = Bonk(u"hello–world", 2)
        result = serialize(bonk)
        debonk = deserialize(Bonk(), result)
        self.assertEqual(debonk.hello, u"hello–world")
        self.assertEqual(debonk.type, 2)

    def test_alternative_protocol(self):
        """
        Ensure serialize will use a given thrift protocol factory.
        """
        bonk = Bonk(u"hello–world", 2)
        pf = TJSONProtocol.TJSONProtocolFactory()
        result = serialize(bonk, pf)
        self.assertContains(result, "hello")
        debonk = deserialize(Bonk(), result, pf)
        self.assertEqual(debonk.hello, u"hello–world")
        self.assertEqual(debonk.type, 2)
