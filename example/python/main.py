import logging
import sys
sys.path.append('gen-py.tornado')

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

from event.f_Events_publisher import EventsPublisher
from event.f_Foo import Client as FFooClient
from event.ttypes import Event


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

    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    logging.debug("Connecting to NATS")
    yield nats_client.connect(**options)

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    if "-client" in sys.argv or len(sys.argv) == 1:
        logging.debug("Running FFooClient")
        yield run_client(nats_client, prot_factory)
    if "-publisher" in sys.argv or len(sys.argv) == 1:
        logging.debug("Running EventsPublisher")
        yield run_publisher(nats_client, prot_factory)

    yield nats_client.close()


@gen.coroutine
def run_client(nats_client, prot_factory):
    transport_factory = FMuxTornadoTransportFactory()
    nats_transport = TNatsServiceTransport(nats_client, "foo", 60000, 5)
    tornado_transport = transport_factory.get_transport(nats_transport)

    try:
        yield tornado_transport.open()
    except TTransport.TTransportException as ex:
        logging.error(ex)
        raise gen.Return()

    foo_client = FFooClient(tornado_transport, prot_factory)

    print 'oneWay()'
    foo_client.oneWay(FContext(), 99, {99: "request"})

    print 'basePing()'
    yield foo_client.basePing(FContext())

    print 'ping()'
    yield foo_client.ping(FContext())

    ctx = FContext()
    event = Event(42, "hello world")
    print 'blah()'
    b = yield foo_client.blah(ctx, 100, "awesomesauce", event)
    print 'Blah response {}'.format(b)
    print 'Response header foo: {}'.format(ctx.get_response_header("foo"))

    yield tornado_transport.close()


@gen.coroutine
def run_publisher(nats_client, prot_factory):
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, prot_factory)

    publisher = EventsPublisher(provider)
    yield publisher.open()

    event = Event(42, "boomtown")
    yield publisher.publish_event_created(FContext(), "barUser", event)
    yield publisher.close()


if __name__ == '__main__':
    # Since we can exit after the client calls use `run_sync`
    ioloop.IOLoop.instance().run_sync(main)
