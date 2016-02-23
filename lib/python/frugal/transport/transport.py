from thrift.transport.TTransport import TTransportBase


class FTransport(TTransportBase):
    """FTranpsort is a THrift TTransport for services."""

    # TODO: implement.

    def __init__(self, registry=None):
        self._registry = registry

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
