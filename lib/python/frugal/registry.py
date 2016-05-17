from threading import Lock

from thrift.transport.TTransport import TMemoryBuffer

from frugal.context import _OP_ID
from frugal.exceptions import FException
from frugal.util.headers import _Headers


class FRegistry(object):
    """
    Registry is responsible for multiplexing received
    messages to the appropriate callback.
    """

    def register(self, context, callback):
        """Register a callback for a given FContext.

        Args:
            context: FContext to register.
            callback: function to register.
        """
        pass

    def unregister(self, context):
        """Unregister the callback for a given FContext.

        Args:
            context: FContext to unregister.
        """
        pass

    def execute(self, frame):
        """Dispatch a single Frugal message frame.

        Args:
            frame: an entire Frugal message frame.
        """
        pass


class FServerRegistry(FRegistry):
    """
    FServerRegistry is intended for use only by Frugal servers.
    This is only to be used by generated code.
    """

    def __init__(self, processor, iprot_factory, oprot):
        """Initialize FServerRegistry.

        Args:
            processor: FProcessor is the server request processor.
            iprot_factory: FProtocolFactory used for creating input
                                    protocols.
            oprot: output FProtocol.
        """
        self._processor = processor
        self._iprot_factory = iprot_factory
        self._oprot = oprot

    def register(self, context, callback):
        # No-op in server.
        pass

    def unregister(self, context):
        # No-op in server.
        pass

    def execute(self, frame):
        """Dispatch a single Frugal message frame.

        Args:
            frame: an entire Frugal message frame.
        """
        wrapped_transport = TMemoryBuffer(frame)
        iprot = self._inputProtocolFactory.get_protocol(wrapped_transport)
        self._processor.process(iprot, self._oprot)


class FClientRegistry(FRegistry):
    """
    FClientRegistry is intended for use only by Frugal clients.
    This is only to be used by generated code.
    """

    def __init__(self):
        self._handlers = {}
        self._handlers_lock = Lock()
        self._next_opid = 0
        self._opid_lock = Lock()

    def register(self, context, callback):
        """Register a callback for a given FContext.

        Args:
            context: FContext to register.
            callback: function to register.
        """
        with self._handlers_lock:
            if context._get_op_id() in self._handlers:
                raise FException("context already registered")

        op_id = self._increment_and_get_next_op_id()
        context._set_op_id(op_id)

        with self._handlers_lock:
            self._handlers[str(op_id)] = callback

    def unregister(self, context):
        """Unregister the callback for a given FContext.

        Args:
            context: FContext to unregister.
        """
        with self._handlers_lock:
            self._handlers.pop(str(context._get_op_id()), None)

    def execute(self, frame):
        """Dispatch a single Frugal message frame.

        Args:
            frame: an entire Frugal message frame.
        """
        headers = _Headers.decode_from_frame(frame)
        op_id = headers[_OP_ID]

        self._handlers[op_id](TMemoryBuffer(frame[4:]))

    def _increment_and_get_next_op_id(self):
        with self._opid_lock:
            self._next_opid += 1
            op_id = self._next_opid
        return op_id

