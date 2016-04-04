from threading import Lock

from tornado import ioloop, gen

from .transport import FTransport


class FMuxTornadoTransport(FTransport):
    """FMuxTornadoTransport is a wrapper around a TFramedTransport"""

    def __init__(self, framed_transport, io_loop=None):
        super(FTransport, self).__init__()
        self._registry = None
        self._transport = framed_transport
        self.io_loop = io_loop or ioloop.IOLoop.current()
        self._lock = Lock()

    def isOpen(self):
        return (self._transport.isOpen() and self._registry)

    @gen.coroutine
    def open(self):
        yield self._transport.open()

    @gen.coroutine
    def close(self):
        yield self._transport.close()

    def set_registry(self, registry):
        with self._lock:
            if registry is None:
                raise StandardError("registry cannot be null.")

            if self._registry is not None:
                return

            self._registry = registry

    def register(self, context, callback):
        with self._lock:
            if self._registry is None:
                raise StandardError("registry cannot be null.")
            else:
                self._registry.register(context, callback)

    def unregister(self, context):
        with self._lock:
            if self._registry is None:
                raise StandardError("registry cannot be null.")

            self._registry.unregister(context)

