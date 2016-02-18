import logging
from . import FServer
from thrift.transport import TTransport

logger = logging.getLogger(__name__)


class FSimpleServer(FServer):
    """Simple single-threaded server that just pumps around one transport."""

    def __init__(self, *args):
        FServer.__init__(self, *args)

    def serve(self):
        self.serverTransport.listen()
        while True:
            client = self.serverTransport.accept()
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

