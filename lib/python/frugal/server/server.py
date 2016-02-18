from thrift.protocol import TBinaryProtocol
from thrift.transport import TTransport


class FServer(object):
    """Base interface for a server, which must have a serve() method.
    Three constructors for all servers:
    1) (processor, serverTransport)
    2) (processor, serverTransport, transportFactory, protocolFactory)
    3) (processor, serverTransport,
        inputTransportFactory, outputTransportFactory,
        inputProtocolFactory, outputProtocolFactory)
    """
    def __init__(self, *args):
        if (len(args) == 2):
            self.__initArgs__(args[0],
                              args[1],
                              TTransport.TTransportFactoryBase(),
                              TTransport.TTransportFactoryBase(),
                              TBinaryProtocol.TBinaryProtocolFactory(),
                              TBinaryProtocol.TBinaryProtocolFactory())
        elif (len(args) == 4):
            self.__initArgs__(args[0], args[1],
                              args[2], args[2],
                              args[3], args[3])
        elif (len(args) == 6):
            self.__initArgs__(args[0], args[1],
                              args[2], args[3],
                              args[4], args[5])

    def __initArgs__(self, processor, serverTransport,
                     inputTransportFactory, outputTransportFactory,
                     inputProtocolFactory, outputProtocolFactory):
        self.processor = processor
        self.serverTransport = serverTransport
        self.inputTransportFactory = inputTransportFactory
        self.outputTransportFactory = outputTransportFactory
        self.inputProtocolFactory = inputProtocolFactory
        self.outputProtocolFactory = outputProtocolFactory

    def serve(self):
        pass

    def stop(self):
        pass
