import logging

from frugal.registry import FServerRegistry
from frugal.server import FServer

logger = logging.getLogger(__name__)


class FSimpleServer(FServer):
    """Simple single-threaded server that just pumps around one transport."""

    def __init__(self,
                 f_processor_factory,
                 thrift_server_transport,
                 f_transport_factory,
                 f_protocol_factory):
        """Initalize an FSimpleServer

        Args:
            processor_factory: FProcessorFactory
            thrift_server_transport: TServerTranpsort
            f_transport_factory: FTransportFactory
            f_protocol_factory: FProtocolFactory
        """

        self._processor_factory = f_processor_factory
        self._transport = thrift_server_transport
        self._transport_factory = f_transport_factory
        self._protocol_factory = f_protocol_factory
        self._stopped = False

    def serve(self):
        self._accept_loop()

    def stop(self):
        self._stopped = True
        self._transport.close()

    def _accept_loop(self):
        while not self._stopped:
            self._transport.listen()
            client = self._transport.accept()
            if not client:
                continue
            self._accept(client)

    def _accept(self, client_transport):
        processor = self._processor_factory.get_processor(client_transport)
        transport = self._transport_factory.get_transport(client_transport)
        protocol = self._protocol_factory.get_protocol(transport)
        transport.set_registry(FServerRegistry(processor,
                                               self._protocol_factory,
                                               protocol))
        transport.open()


