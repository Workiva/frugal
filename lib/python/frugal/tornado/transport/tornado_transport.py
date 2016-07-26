from io import BytesIO
import logging
from struct import pack

from thrift.transport.TTransport import TTransportException
from tornado import gen, locks

from frugal.exceptions import FMessageSizeException
from frugal.transport import FTransport
from frugal.util.deprecate import deprecated

logger = logging.getLogger(__name__)


class FTornadoTransport(FTransport):
    """ FTornadoTransport implements the buffered write data and registry
    interactions shared by all FTransports.
    """

    def __init__(self, max_message_size=1024 * 1024):
        super(FTornadoTransport, self).__init__()
        self._max_message_size = max_message_size
        self._wbuf = BytesIO()

        # TODO: Remove this with 2.0
        self._execute = None
        self._open_lock = locks.Lock()

    # TODO: Remove with 2.0
    @deprecated
    def set_execute_callback(self, execute):
        """Set the message callback execute function

        Args:
            execute: message callback execute function

        @deprecated Construct transports which call execute_frame,
                    triggering the registry directly.
        """
        self._execute = execute

    # TODO: With 2.0, make this a gen.coroutine and protect registry access
    # with a tornado lock
    def set_registry(self, registry):
        """Set FRegistry on the transport.  No-Op if already set.
        args:
            registry: FRegistry to set on the transport
        """
        if not registry:
            ex = ValueError("registry cannot be null.")
            logger.exception(ex)
            raise ex

        # TODO: With 2.0, consider throwing a StandardError.
        # Currently, the generated code sets the registry for each extending
        # service for a particular base service.
        if self._registry:
            return

        self._registry = registry

    # TODO: With 2.0, make this a gen.coroutine and protect registry access
    # with a tornado lock
    def register(self, context, callback):
        if not self._registry:
            ex = StandardError("registry cannot be null.")
            logger.exception(ex)
            raise ex

        self._registry.register(context, callback)

    # TODO: With 2.0, make this a gen.coroutine and protect registry access
    # with a tornado lock
    def unregister(self, context):
        if not self._registry:
            ex = StandardError("registry cannot be null.")
            logger.exception(ex)
            raise ex

        self._registry.unregister(context)

    @gen.coroutine
    def isOpen(self):
        raise gen.Return(NotImplementedError("You must override this."))

    @gen.coroutine
    def open(self):
        raise gen.Return(NotImplementedError("You must override this."))

    @gen.coroutine
    def close(self):
        raise gen.Return(NotImplementedError("You must override this."))

    def read(self, size):
        raise NotImplementedError("Don't call this.")

    @gen.coroutine
    def write(self, buff):
        """Writes the bytes to a buffer. Throws FMessageSizeException if the
        buffer exceeds limit.

        Args:
            buff: buffer to append to the write buffer.
        """
        if not (yield self.isOpen()):
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Transport not open!")
            logger.exception(ex)
            raise ex

        size = len(buff) + len(self._wbuf.getvalue())

        if size > self._max_message_size > 0:
            ex = FMessageSizeException("Message exceeds max message size")
            logger.exception(ex)
            raise ex

        self._wbuf.write(buff)

    @gen.coroutine
    def flush(self):
        raise gen.Return(NotImplementedError("You must override this."))

    def get_write_bytes(self):
        """Get the framed bytes from the write buffer."""
        frame = self._wbuf.getvalue()
        if len(frame) == 0:
            return None

        frame_length = pack('!I', len(frame))
        return '{0}{1}'.format(frame_length, frame)

    def reset_write_buffer(self):
        """Reset the write buffer."""
        self._wbuf = BytesIO()

    def execute_frame(self, frame):
        """Execute a frugal frame.
        NOTE: this frame must include the frame size.
        """
        self._registry.execute(frame[4:])

