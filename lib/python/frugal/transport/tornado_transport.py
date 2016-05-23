from threading import Lock

from thrift.transport.TTransport import TTransportException
from tornado import ioloop, gen

from .transport import FTransport
from .transport_factory import FTransportFactory


class FMuxTornadoTransport(FTransport):
    """FMuxTornadoTransport is a wrapper around a TFramedTransport"""

    def __init__(self, framed_transport, io_loop=None):
        super(FTransport, self).__init__()
        self._registry = None
        self._transport = framed_transport
        self.io_loop = io_loop or ioloop.IOLoop.current()
        self._lock = Lock()

    def isOpen(self):
        return self._transport.isOpen()

    @gen.coroutine
    def open(self):
        try:
            yield self._transport.open()
        except TTransportException as ex:
            if ex.type != TTransportException.ALREADY_OPEN:
                # It's okay if transport is already open
                raise ex

    @gen.coroutine
    def close(self):
        yield self._transport.close()

    def set_registry(self, registry):
        """Set FRegistry on the transport.  No-Op if already set.
        args:
            registry: FRegistry to set on the transport
        """
        with self._lock:
            if not registry:
                raise ValueError("registry cannot be null.")

            if self._registry:
                return

            self._registry = registry
            self._transport.set_execute_callback(registry.execute)

    def register(self, context, callback):
        with self._lock:
            if not self._registry:
                raise StandardError("registry cannot be null.")

            self._registry.register(context, callback)

    def unregister(self, context):
        with self._lock:
            if not self._registry:
                raise StandardError("registry cannot be null.")

            self._registry.unregister(context)

    def read(self):
        raise StandardError("you're doing it wrong")

    def write(self, buff):
        self._transport.write(buff)

    def flush(self):
        self._transport.flush()


class FMuxTornadoTransportFactory(FTransportFactory):
    """Factory for creating FMuxTransports."""

    def get_transport(self, thrift_transport):
        """ Returns a new FMuxTransport wrapping the given TTransport

        Args:
            thrift_transport: TTransport to wrap
        Returns:
            new FTransport
        """

        return FMuxTornadoTransport(thrift_transport)
