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

import struct

from thrift.transport.TTransport import TTransportException

from frugal.exceptions import TTransportExceptionType
from thrift.transport.TTransport import TMemoryBuffer


class TMemoryOutputBuffer(TMemoryBuffer, object):
    """
    An implementation of TMemoryBuffer using a bounded memory buffer. Writes
    that cause the buffer to exceed its size throw an FMessageSizeException.
    This implementation handles framing data.
    """

    def __init__(self, limit, value=None):
        """
        Create an instance of FBoundedMemoryBuffer where size is the
        maximum writable length of the buffer.

           Args:
               limit: integer size limit of the buffer
               value: optional data value to initialize the buffer with.
        """
        super(TMemoryOutputBuffer, self).__init__(value)
        self._limit = limit

    def write(self, buf):
        """
        Bounded write to buffer
        """
        if len(self) + len(buf) > self._limit > 0:
            self._buffer = TMemoryBuffer()
            raise TTransportException(
                type=TTransportExceptionType.REQUEST_TOO_LARGE,
                message="Buffer size reached {}".format(self._limit))
        self._buffer.write(buf)

    def getvalue(self):
        # TODO make more efficient?
        data = self._buffer.getvalue()
        return struct.pack('!I', len(data)) + data

    def read(self, sz):
        raise Exception("don't call this")

    def __len__(self):
        return len(self.getvalue())
