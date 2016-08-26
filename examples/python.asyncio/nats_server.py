# -*- coding: utf-8 -*-
import logging
import sys
import asyncio

from aiohttp import web
from nats.aio.client import Client as NatsClient

from thrift.protocol import TBinaryProtocol
from frugal.protocol import FProtocolFactory
from frugal.aio.server import FNatsServer

sys.path.append('gen-py.asyncio')

from music.f_Store import Processor as FStoreProcessor  # noqa
from music.f_Store import Iface  # noqa
from music.ttypes import Album, Track  # noqa


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


class StoreHandler(Iface):
    """
    A handler handles all incoming requests to the server.
    The handler must satisfy the interface the server exposes.
    """

    def buyAlbum(self, ctx, ASIN, acct):
        """
        Return an album; always buy the same one.
        """
        album = Album()
        album.ASIN = uuid.uuid4()
        album.duration = 12000
        return album

    def enterAlbumGiveaway(self, ctx, email, name):
        """
        Always return success (true)
        """
        return True


async def main():
    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between clients and servers.
    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    # Open a NATS connection to receive requests
    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }
    await nats_client.connect(**options)

    # Create a new server processor.
    # Incoming requests to the processor are passed to the handler.
    # Results from the handler are returned back to the client.
    processor = FStoreProcessor(StoreHandler())

    # Create a new music store server using the processor,
    # The sever will listen on the music-service NATS topic
    server = FNatsServer(nats_client, "music-service", processor, prot_factory)

    root.info("Starting Nats server...")

    await server.serve()


if __name__ == '__main__':
    io_loop = asyncio.get_event_loop()
    asyncio.ensure_future(main())
    io_loop.run_forever()
