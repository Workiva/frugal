from thrift.protocol.TBinaryProtocol import TBinaryProtocolFactory
import tornado.ioloop
import tornado.gen
from nats.io.utils import new_inbox
from nats.io.client import Client as NATS

from frugal.context import FContext
from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.transport.mux_transport_factory import FMuxTransportFactory
from frugal.transport.nats_service_transport import TNatsServiceTransport

from example import FFoo


def main():

    nats_client = NATS()

    # Create FProtocolFactory passing it TBinaryProtocolFactory
    protocol_factory = FProtocolFactory(TBinaryProtocolFactory())
    transport_factory = FMuxTransportFactory(5)

    # Yield the connection to nats
    nats_transport = TNatsServiceTransport(nats_client, "foo", 5000, 3)

    # Create the FTransport, pass the factory method a TTransport
    transport = transport_factory.get_transport(nats_transport)
    transport.open()

    foo_client = FFoo.Client(transport, protocol_factory)
    foo_client.one_way(FContext(), 99, {99: "request"})


if __name__ == '__main__':
    tornado.ioloop.IOLoop.instance().run_sync(main)
