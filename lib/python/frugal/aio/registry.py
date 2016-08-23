from asyncio import Lock
import logging

from thrift.transport.TTransport import TMemoryBuffer

from frugal.context import _OP_ID
from frugal.exceptions import FException
from frugal.util.headers import _Headers

logger = logging.getLogger(__name__)


class FRegistry(object):
    """
    Registry is responsible for multiplexing received
    messages to the appropriate callback.
    """

    async def register(self, context, callback):
        """Register a callback for a given FContext.

        Args:
            context: FContext to register.
            callback: function to register.
        """
        pass

    async def unregister(self, context):
        """Unregister the callback for a given FContext.

        Args:
            context: FContext to unregister.
        """
        pass

    async def execute(self, frame):
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

    async def execute(self, frame):
        """Dispatch a single Frugal message frame.

        Args:
            frame: an entire Frugal message frame.
        """
        wrapped_transport = TMemoryBuffer(frame)
        iprot = self._iprot_factory.get_protocol(wrapped_transport)
        await self._processor.process(iprot, self._oprot)


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

    async def register(self, context, callback):
        """Register a callback for a given FContext.

        Args:
            context: FContext to register.
            callback: function to register.
        """
        # An FContext can be reused for multiple requests. Because of this,
        # every time an FContext is registered, it must be assigned a new op id
        # to ensure we can properly correlate responses. We use a monotonically
        # increasing atomic uint64 for this purpose. If the FContext already
        # has an op id, it has been used for a request. We check the handlers
        # map to ensure that request is not still in-flight.
        async with self._handlers_lock:
            if str(context._get_op_id()) in self._handlers:
                raise FException("context already registered")

        op_id = await self._increment_and_get_next_op_id()
        context._set_op_id(op_id)

        async with self._handlers_lock:
            self._handlers[str(op_id)] = callback

    async def unregister(self, context):
        """Unregister the callback for a given FContext.

        Args:
            context: FContext to unregister.
        """
        async with self._handlers_lock:
            self._handlers.pop(str(context._get_op_id()), None)

    async def execute(self, frame):
        """Dispatch a single Frugal message frame.

        Args:
            frame: an entire Frugal message frame.
        """
        headers = _Headers.decode_from_frame(frame)
        op_id = headers.get(_OP_ID, None)

        if not op_id:
            raise FException("Frame missing op_id")

        async with self._handlers_lock:
            handler = self._handlers.get(op_id, None)
            if not handler:
                logger.warning("Got a message for unregistered context."
                               "Dropping")
                return

            handler(TMemoryBuffer(frame))

    async def _increment_and_get_next_op_id(self):
        async with self._opid_lock:
            self._next_opid += 1
            op_id = self._next_opid
        return op_id
