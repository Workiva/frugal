from threading import Lock

from thrift.transport.TTransport import TMemoryBuffer

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

    def __init__(self, processor, input_protocol_factory, output_protocol):
        """Initialize FServerRegistry.

        Args:
            processor: FProcessor is the server request processor.
            input_protocol_factory: FProtocolFactory used for creating input
                                    protocols.
            output_protocol: output FProtocol.
        """
        self._processor = processor
        self._input_protocol_factory = input_protocol_factory
        self._output_protocol = output_protocol

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
        tr = TMemoryBuffer(frame)
        self._processor.process(
            self._inputProtocolFactory.get_protocol(tr),
            self._outputProtocol
        )


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
            if str(context._get_op_id()) in self._handlers:
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
        print(frame)
        buff_without_frame = frame[4:]
        headers = _Headers._read(buff_without_frame)
        print("headers: {}".format(headers))
        op_id = headers["_opid"]

        self._handlers[op_id](TMemoryBuffer(buff_without_frame))

    def _increment_and_get_next_op_id(self):
        with self._opid_lock:
            self._next_opid += 1
            op_id = self._next_opid
        return op_id

