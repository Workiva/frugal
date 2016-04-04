from .tornado_transport import FMuxTornadoTransport


class FTransportFactory(object):
    """FTransportFactory is responsible for creating new FTransports."""

    def get_transport(self, thrift_transport):
        """ Retuns a new FTransport wrapping the given TTransport.

        Args:
            thrift_transport: TTransport to wrap.
        Returns:
            new FTranpsort
        """
        pass


class FScopeTransportFactory(FTransportFactory):
    """Factory Interface for creating FScopeTransports"""

    def get_transport(self):
        """ Get a new FScopeTransport instance.

        Returns:
            FScopeTransport
        """

        pass


class FMuxTransportFactory(FTransportFactory):
    """Factory for creating FMuxTransports."""

    def get_transport(self, thrift_transport):
        """ Returns a new FMuxTransport wrapping the given TTransport

        Args:
            thrift_transport: TTransport to wrap
        Returns:
            new FTransport
        """

        return FMuxTornadoTransport(thrift_transport)
