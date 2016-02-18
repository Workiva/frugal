from threading import Lock

from thrift.transport.TTransport import TMemoryBuffer

from frugal.exceptions import FException


class FRegistry(object):
    """
    Registry is responsible for multiplexing received
    messages to the appropriate callback.
    """

    def register(self, context):
        pass

    def unregister(self, context):
        pass

    def execute(self, frame):
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
        with self._handlers_lock:
            if context.get_op_id() in self._handlers:
                raise FException("context already registered")

        op_id = self._increment_and_get_next_op_id()
        with self._handlers_lock:
            self._handlers[op_id] = callback

    def unregister(self, context):
        with self._handlers_lock:
            self._handlers.pop(context.get_op_id(), None)

    def execute(self, frame):
        # TODO
        pass

    def _increment_and_get_next_op_id(self):
        with self._opid_lock:
            self._next_opid += 1
            op_id = self._next_opid
        return op_id

