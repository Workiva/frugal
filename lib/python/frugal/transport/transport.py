from threading import Lock

from thrift.transport.TTransport import TTransportBase, TFramedTransport


class FTransport(TTransportBase, object):
    """FTranpsort is a Thrift TTransport for services."""

    def __init__(self):
        self._registry = None

    def set_registry(self, registry):
        """Set the FRegistry for the transport

        Args:
            registry: FRegistry
        """
        pass

    def register(self, context, callback):
        pass

    def unregister(self, context):
        pass

    def set_monitor(self, monitor):
        pass

    def closed(self):
        pass


class FMuxTransport(FTransport):
    """FMuxTrasport is a Multiplexed FTranpsort"""

    def __init__(self, thrift_transport):
        """Construct a new FMuxTransport

        Args:
            thrift_transport: TTransport to wrap
        """
        super(FTransport, self).__init__()
        self._registry = None
        self._framed_transport = TFramedTransport(thrift_transport)
        self._lock = Lock()

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

    def is_open(self):
        with self._lock:
            return (self._framed_transport.isOpen() and
                    self._registry is not None)

    def open(self):
        with self._lock:
            self._framed_transport.open()

    def close(self):
        with self._lock:
            if self._registry is None:
                return

            self._framed_transport.close()

    def read(self, size):
        self._framed_transport.read(size)

    def write(self, buff):
        self._framed_transport.write(buff)

    def flush(self):
        with self._lock:
            self._framed_transport.flush()
