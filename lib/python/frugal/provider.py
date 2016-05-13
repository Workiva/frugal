
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
        """Return a tupled FScopeTransport and FProtocol.
        Returns:
            (FScopeTransport, FProtocolFactory)
        """
        transport = self._transport_factory.get_transport()
        return transport, self._protocol_factory


class FServiceProvider(object):
    """FServiceProvider is the service equivalent of FScopeProvider. It produces
     FTransports and FProtocols for use by RPC service clients.
     """

    def __init__(self, transport, protocol_factory):
        self._transport = transport
        self._protocol_factory = protocol_factory

    def get_transport(self):
        """Get the FTransport from the provider."""
        return self._transport

    def get_protocol_factory(self):
        """Get the FProtocolFactory from the provider."""
        return self._protocol_factory

