import logging
import sys
sys.path.append('gen-py.tornado')

from thrift.protocol import TBinaryProtocol

from tornado import ioloop
from tornado import gen

from nats.io.client import Client as NATS

from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.provider import FScopeProvider
from frugal.transport.nats_scope_transport import FNatsScopeTransportFactory

from event.f_Events_subscriber import EventsSubscriber


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

    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    yield nats_client.connect(**options)

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, prot_factory)

    subscriber = EventsSubscriber(provider)

    def event_handler(ctx, req):
        print "Received an event with ID: {} and Message {}".format(req.ID,
                                                                    req.Message)

    yield subscriber.subscribe_EventCreated("barUser", event_handler)

    logging.info("Subscriber starting...")

if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
