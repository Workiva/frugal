from frugal.exceptions import FException
from threading import Lock

NEXT_OP_ID = 0


def increment_and_get_next_op_id():
    lock = Lock()
    lock.acquire()
    op_id = NEXT_OP_ID + 1
    lock.release()
    return op_id


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

    def close(self):
        pass


class FServerRegistry(FRegistry):
    """
    FServerRegistry is intended for use only by Frugal servers.
    This is only to be used by generated code.
    """

    def __init__(self, processor, inputProtocolFactory, outputProtocol):
        self._processor = processor
        self._inputProtocolFactory = inputProtocolFactory
        self._outputProtocol = outputProtocol

    def register(self, context, callback):
        pass

    def unregister(self, context):
        pass

    def execute(self, frame):
        self._processor.process(
            #TODO add the TMemoryProtocol
            self._inputProtocolFactory.get_protocol(),
            self._outputProtocol
        )

    def close(self):
        pass


class FClientRegistry(FRegistry):
    """
    FClientRegistry is intended for use only by Frugal clients.
    This is only to be used by generated code.
    """

    def __init__(self):
        self._handlers = {}

    def register(self, context, callback):
        if context.get_op_id() in self._handlers:
            raise FException("context already registered")

        op_id = increment_and_get_next_op_id()
        self._handlers[op_id] = callback

    def unregister(self, context):
        self._handlers.pop(context.get_op_id(), None)

    def execute(self, frame):
        pass

    def close(self):
        pass
