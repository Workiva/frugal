from io import BytesIO
import struct

from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FMessageSizeException
from frugal.transport import FTransport


class FTransportBase(FTransport):
    """
    FTransportBase serves as a base class for FTransports,
    implementing a potentially size limited buffered write method.
    """
    def __init__(self, max_message_size: int):
        """
        Args:
            max_message_size: The maximum amount of data allowed to be buffered,
                              0 indicates unlimited size.
        """
        super().__init__()
        self._max_message_size = max_message_size
        self._wbuf = BytesIO()

    def isOpen(self):
        raise NotImplementedError('You must override this')

    async def open(self):
        raise NotImplementedError('You must override this')

    async def close(self):
        raise NotImplementedError('You must override this')

    def read(self, size):
        raise Exception('Do not call this')

    def write(self, buf):
        """
        Writes the given data to a buffer. Throws FMessageSizeException if
        this will cause the buffer to exceed the message size limit.

        Args:
            buf: The data to write.
        """
        size = len(buf) + len(self._wbuf.getvalue())

        if size > self._max_message_size > 0:
            self._wbuf = BytesIO()
            raise FMessageSizeException('Message exceeds max message size')
        self._wbuf.write(buf)

    async def flush(self):
        raise NotImplementedError('You must override this')

    def get_write_bytes(self):
        """Get the framed bytes from the write buffer."""
        data = self._wbuf.getvalue()
        if len(data) == 0:
            return None

        data_length = struct.pack('!I', len(data))
        return data_length + data

    def reset_write_buffer(self):
        """Reset the write buffer."""
        self._wbuf = BytesIO()
