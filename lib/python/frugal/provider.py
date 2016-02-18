
class FScopeProvider(object):
    """
    FScopeProviders produce FScopeTransports and FProtocols for use
    with Frugal Publishers and Subscribers.
    """

    def __init__(self, transport_factory, protocol_factory):
        """Initialize FScopeProvider.

        Args:
            transport_factory: FScopeTransportFactory.
            protocol_factory: FProtocolFactory.
        """
        self._transport_factory = transport_factory
        self._protocol_factory = protocol_factory

    def new(self):
        """Return a new FScopeTransport and FProtocol."""
        transport = self._transport_factory.get_transport()
        protocol = self._protocol_factory.get_protocol(transport)
        return transport, protocol
