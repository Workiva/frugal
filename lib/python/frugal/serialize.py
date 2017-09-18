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

from thrift.protocol import TBinaryProtocol
from thrift.transport import TTransport

from frugal.protocol import FProtocolFactory


def serialize(
        frugal_object,
        protocol_factory=TBinaryProtocol.TBinaryProtocolFactory()):
    """Serialize a frugal entity to bytes."""
    transport = TTransport.TMemoryBuffer()
    fprotocolFactory = FProtocolFactory(protocol_factory)
    protocol = fprotocolFactory.get_protocol(transport)
    frugal_object.write(protocol)
    return transport.getvalue()


def deserialize(
        base,
        buf,
        protocol_factory=TBinaryProtocol.TBinaryProtocolFactory()):
    """Deserialize a frugal object into a base instance of a frugal object."""
    transport = TTransport.TMemoryBuffer(buf)
    fprotocolFactory = FProtocolFactory(protocol_factory)
    protocol = fprotocolFactory.get_protocol(transport)
    base.read(protocol)
    return base
