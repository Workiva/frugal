import logging
import os
import sys
import asyncio

from thrift.protocol import TBinaryProtocol

from nats.aio.client import Client as NatsClient

from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.provider import FScopeProvider
from frugal.aio.transport import FNatsSubscriberTransportFactory

sys.path.append(os.path.join(os.path.dirname(__file__), "gen-py.asyncio"))
from v1.music.f_AlbumWinners_subscriber import AlbumWinnersSubscriber  # noqa


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


async def main():
    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between publishers and subscribers.
    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    # Open a NATS connection to receive requests
    nats_client = NatsClient()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    await nats_client.connect(**options)

    # Create a pub sub scope using the configured transport and protocol
    transport_factory = FNatsSubscriberTransportFactory(nats_client)
    provider = FScopeProvider(None, transport_factory, prot_factory)

    subscriber = AlbumWinnersSubscriber(provider)

    def event_handler(ctx, req):
        root.info("You won! {}".format(req.ASIN))

    def start_contest_handler(ctx, albums):
        root.info("Contest started, available albums: {}".format(albums))

    await subscriber.subscribe_Winner(event_handler)
    await subscriber.subscribe_ContestStart(start_contest_handler)

    root.info("Subscriber starting...")

if __name__ == '__main__':
    io_loop = asyncio.get_event_loop()
    asyncio.ensure_future(main())
    io_loop.run_forever()
