import logging

from thrift.protocol import TBinaryProtocol
from thrift.transport import TTransport

from tornado import ioloop
from tornado import gen

from nats.io.client import Client as NATS

from frugal.context import FContext
from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.transport.transport_factory import FMuxTransportFactory
from frugal.transport.nats_service_transport import TNatsServiceTransport

from gen_py.example.f_foo import Client as FFooClient


@gen.coroutine
def main():

    print("Starting.....")

    nats_client = NATS()
    options = {"verbose": True, "servers": ["nats://127.0.0.1:4222"]}
    yield nats_client.connect(**options)

    transport_factory = FMuxTransportFactory()
    nats_transport = TNatsServiceTransport(nats_client, "foo", 50000, 3)
    transport = transport_factory.get_transport(nats_transport)

    try:
        yield transport.open()
    except TTransport.TTransportException as ex:
        print("got TTransportException")
        logging.error(ex)
        raise gen.Return()

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    foo_client = FFooClient(transport, prot_factory)
    foo_client.one_way(FContext(), 99, {99: "request"})

    raise gen.Return("")


if __name__ == '__main__':
    ioloop.IOLoop.instance().run_sync(main)
