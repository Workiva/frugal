from thrift.transport.TTransport import TTransportBase


class FTransport(TTransportBase):

    # TODO: implement.

    def __init__(self, registry=None):
        self._registry = registry

    def set_registry(registry):
        pass

    def register(context, callback):
        pass

    def unregister(context):
        pass

    def set_monitor(monitor):
        pass

    def closed():
        pass
