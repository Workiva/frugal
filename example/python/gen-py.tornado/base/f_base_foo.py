

class IFace(object):

    def base_ping(context):
        pass


class Client(Iface):

    def __init__(self, transport, protocol_factory):
        self._transport = transport
        self._transport.set_registry(FClientRegistry())
        self._protocol_factory = protocol_factory
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._oport = self._protocol_factory.get_protocol(self._transport)

