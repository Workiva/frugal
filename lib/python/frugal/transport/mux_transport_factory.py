from .transport_factory import FTransportFactory
from .mux_transport import FMuxTransport


class FMuxTransportFactory(FTransportFactory):

    def __init__(self, num_workers):
        """ Construct new FMuxTransportFactory

        Args:
            num_workers: number of worker threads for the FTransport
        """

        self._num_workers = num_workers

    def get_transport(self, thrift_transport):
        """ Returns a new FMuxTransport wrapping the given TTransport

        Args:
            thrift_transport: TTransport to wrap
        Returns:
            new FTransport
        """

        return FMuxTransport(thrift_transport, self._num_workers)
