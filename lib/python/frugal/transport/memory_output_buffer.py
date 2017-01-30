import struct

from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FrugalTTransportExceptionType
from thrift.transport.TTransport import TMemoryBuffer


class TMemoryOutputBuffer(TMemoryBuffer, object):
    """
    An implementation of TMemoryBuffer using a bounded memory buffer. Writes
    that cause the buffer to exceed its size throw an FMessageSizeException.
    This implementation handles framing data.
    """

    def __init__(self, limit, value=None):
        """Create an instance of FBoundedMemoryBuffer where size is the
        maximum writable length of the buffer.

           Args:
               limit: integer size limit of the buffer
               value: optional data value to initialize the buffer with.
        """
        super(TMemoryOutputBuffer, self).__init__(value)
        self._limit = limit

    def write(self, buf):
        """Bounded write to buffer"""
        if len(self) + len(buf) > self._limit > 0:
            self._buffer = TMemoryBuffer()
            raise TTransportException(
                type=FrugalTTransportExceptionType.REQUEST_TOO_LARGE,
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
