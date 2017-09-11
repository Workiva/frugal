import logging
import sys

from thrift.protocol import TBinaryProtocol

from tornado import ioloop
from tornado import gen

from nats.io.client import Client as NATS

from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.provider import FScopeProvider
from frugal.tornado.transport import FNatsSubscriberTransportFactory

sys.path.append('gen-py.tornado')
from v1.music.f_AlbumWinners_subscriber import AlbumWinnersSubscriber  # noqa


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
    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between publishers and subscribers.
    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    # Open a NATS connection to receive requests
    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    yield nats_client.connect(**options)

    # Create a pub sub scope using the configured transport and protocol
    transport_factory = FNatsSubscriberTransportFactory(nats_client)
    provider = FScopeProvider(None, transport_factory, prot_factory)

    subscriber = AlbumWinnersSubscriber(provider)

    def event_handler(ctx, req):
        root.info("You won! {}".format(req))

    def start_contest_handler(ctx, albums):
        root.info("Contest started, available albums: {}".format(albums))

    yield subscriber.subscribe_Winner(event_handler)
    yield subscriber.subscribe_ContestStart(start_contest_handler)

    root.info("Subscriber starting...")

if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
