
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
