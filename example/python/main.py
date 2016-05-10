import logging
import sys

from thrift.protocol import TBinaryProtocol
from thrift.transport import TTransport

from tornado import ioloop
from tornado import gen

from nats.io.client import Client as NATS

from frugal.context import FContext
from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.provider import FScopeProvider
from frugal.transport.tornado_transport import FMuxTornadoTransportFactory
from frugal.transport.nats_scope_transport import FNatsScopeTransportFactory
from frugal.transport.nats_service_transport import TNatsServiceTransport

from gen_py.example.events_publisher import EventsPublisher
from gen_py.example.f_foo import Client as FFooClient
from gen_py.example.ttypes import Event


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


@gen.coroutine
def main():

    logging.info("Starting...")

    # Create & connect to NATS Client using python-nats
    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    yield nats_client.connect(**options)

    transport_factory = FMuxTornadoTransportFactory()
    nats_transport = TNatsServiceTransport(nats_client, "foo", 60000, 5)
    tornado_transport = transport_factory.get_transport(nats_transport)

    try:
        yield tornado_transport.open()
    except TTransport.TTransportException as ex:
        logging.error(ex)
        raise gen.Return()

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    foo_client = FFooClient(tornado_transport, prot_factory)

    foo_client.one_way(FContext(), 99, {99: "request"})

    yield foo_client.ping(FContext())

    ctx = FContext()
    event = Event(42, "hello world")
    print("sending blah")
    b = yield foo_client.blah(ctx, 100, "awesomesauce", event)
    print("Blah response {}".format(b))
    print("Response header foo: {}".format(ctx.get_response_header("foo")))

    yield tornado_transport.close()

    ####################################
    # Publisher                        #
    ####################################

    yield nats_client.connect(**options)

    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, prot_factory)

    publisher = EventsPublisher(provider)
    yield publisher.open()

    event = Event(66, "boomtown")
    yield publisher.publish_event_created(FContext(), "barUser", event)
    yield publisher.close()


def start():
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()


if __name__ == '__main__':
    start()
