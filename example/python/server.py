from thrift.protocol import TBinaryProtocol
from thrift.transport import TTransport

from tornado import ioloop
from tornado import gen

from nats.io.client import Client as NATS

from frugal.context import FContext
from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.transport.transport_factory import FMuxTransportFactory
from frugal.transport.nats_service_transport import TNatsServiceTransport

from gen_py.example.f_foo import Processor as FFooProcessor

@gen.coroutine
def main():
    nats_client = NATS()
    options = {"verbose": True, "servers": ["nats://127.0.0.1:4222"]}
    yield nats_client.connect(**options)


    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    transport_factory = FMuxTransportFactory()

    handler = FooHanlder()
    processor = FFooProcessor(handler)
    server = FNatsServer(nats_client,
                         "foo",
                         20 * 1000,
                         2,
                         FProcessorFactory(processor),
                         transport_factory,
                         prot_factory)

    print("starting NATS server on foo")


if __name__ == '__main__':
    main()
