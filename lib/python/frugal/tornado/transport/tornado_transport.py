from io import BytesIO
from struct import pack

from tornado import gen, locks

from frugal.exceptions import FException, FMessageSizeException
from frugal.transport import FTransport


class FTornadoTransport(FTransport):
    """ FTornadoTransport implements the buffered write data and registry
    interactions shared by all FTransports.
    """

    def __init__(self, max_message_size=1024 * 1024):
        super(FTornadoTransport, self).__init__()
        self._max_message_size = max_message_size
        self._wbuf = BytesIO()

        self._registry_lock = locks.Lock()

    @gen.coroutine
    def set_registry(self, registry):
        """Set FRegistry on the transport.  No-Op if already set.

        Args:
            registry: FRegistry to set on the transport
        """
        if not registry:
            raise ValueError("registry cannot be null.")

        with (yield self._registry_lock.acquire()):
            if self._registry:
                raise FException("registry already set.")

            self._registry = registry

    @gen.coroutine
    def register(self, context, callback):
        """Register a provided FContext and callback function with the
        transport's internal FRegistry.

        Args:
            context: FContext to register.
            callback: function to register.

        Raises:
            StandardError: if registry has not been set.
        """
        with (yield self._registry_lock.acquire()):
            if not self._registry:
                raise FException("registry cannot be null.")

            self._registry.register(context, callback)

    @gen.coroutine
    def unregister(self, context):
        """Unregsiter the given context from the transports internal registry.

        Args:
            context: FContext to remove from the registry.

        Raises:
            StandardError: if registry has not been set.
        """
        with (yield self._registry_lock.acquire()):
            if not self._registry:
                raise FException("registry cannot be null.")

            self._registry.unregister(context)

    def isOpen(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def open(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def close(self):
        raise NotImplementedError("You must override this.")

    def read(self, size):
        raise NotImplementedError("Don't call this.")

    def write(self, buff):
        """Writes the bytes to a buffer. Throws FMessageSizeException if the
        buffer exceeds limit.

        Args:
            buff: buffer to append to the write buffer.
        """
        size = len(buff) + len(self._wbuf.getvalue())

        if size > self._max_message_size > 0:
            raise FMessageSizeException("Message exceeds max message size")

        self._wbuf.write(buff)

    @gen.coroutine
    def flush(self):
        raise NotImplementedError("You must override this.")

    def get_write_bytes(self):
        """Get the framed bytes from the write buffer."""
        frame = self._wbuf.getvalue()
        if len(frame) == 0:
            return None

        frame_length = pack('!I', len(frame))
        return b'{0}{1}'.format(frame_length, frame)

    def reset_write_buffer(self):
        """Reset the write buffer."""
        self._wbuf = BytesIO()

    def execute_frame(self, frame):
        """Execute a frugal frame.
        NOTE: this frame must include the frame size.
        """
        self._registry.execute(frame[4:])

