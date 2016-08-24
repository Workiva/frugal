from thrift.transport.TTransport import TTransportBase


class FTransport(TTransportBase, object):
    """FTransport is a Thrift TTransport for services."""

    DEFAULT_HIGH_WATERMARK = 5 * 1000

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


class FSynchronousTransport(TTransportBase, object):
    """FSynchronousTransport is a Thrift TTransport for services which makes
    synchronous requests.
    """

    def set_timeout(self, timeout):
        """Set the request timeout.

        Args:
            timeout: request timeout in milliseconds.
        """
        pass

