import logging

from thrift.transport import TTransport

from .server import FServer

logger = logging.getLogger(__name__)


class FSimpleServer(FServer):
    """Simple single-threaded server that just pumps around one transport."""

    def __init__(self,
                 processor_factory,
                 transport,
                 transport_factory,
                 protocol_factory):
        """Initalize an FSimpleServer

        Args:
            processor_factory: FProcessorFactory
            transport: FServerTranpsort
            transport_factory: FTransportFactory
            protocol_factory: FProtocolFactory
        """

        self._processor_factory = processor_factory
        self._transport = transport
        self._transport_factory = transport_factory
        self._protocol_factory = protocol_factory
        self._stopped = False

    def serve(self):
        while not self._stopped:
            client = self._transport.accept()
            if not client:
                continue
            itrans = self.inputTransportFactory.getTransport(client)
            otrans = self.outputTransportFactory.getTransport(client)

            iprot = self.inputProtocolFactory.getProtocol(itrans)
            oprot = self.outputProtocolFactory.getProtocol(otrans)

            try:
                while True:
                    self.processor.process(iprot, oprot)
            except TTransport.TTransportException:
                pass
            except Exception as x:
                logger.exception(x)

            itrans.close()
            otrans.close()

    def _process_requests(self, client_transport):
        processor = self._processor_factory.get_processor(client_transport)
        transport = self._transport_factory.get_transport(client_transport)
        protocol = self._protocol_factory.get_protocol(transport)
