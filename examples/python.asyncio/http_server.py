# -*- coding: utf-8 -*-
import os
import logging
import sys

from aiohttp import web

from thrift.protocol import TBinaryProtocol
from frugal.protocol import FProtocolFactory
from frugal.aio.server.http_handler import new_http_handler

sys.path.append(os.path.join(os.path.dirname(__file__), "gen-py.asyncio"))
from v1.music.f_Store import Processor as FStoreProcessor  # noqa
from v1.music.f_Store import Iface  # noqa
from v1.music.ttypes import Album, Track  # noqa


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
        album.ASIN = ASIN
        album.duration = 12000
        return album

    def enterAlbumGiveaway(self, ctx, email, name):
        """
        Always return success (true)
        """
        return True


if __name__ == '__main__':
    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between clients and servers.
    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    # Create a new server processor.
    # Incoming requests to the processor are passed to the handler.
    # Results from the handler are returned back to the client.
    processor = FStoreProcessor(StoreHandler())

    store_handler = new_http_handler(processor, prot_factory)
    app = web.Application()
    app.router.add_route("*", "/frugal", store_handler)
    web.run_app(app)
