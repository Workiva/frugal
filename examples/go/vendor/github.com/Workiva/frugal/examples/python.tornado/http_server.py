#!/usr/bin/python
# -*- coding: utf-8 -*-
import logging
import sys
import uuid

import tornado
from thrift.protocol import TBinaryProtocol
from frugal.protocol import FProtocolFactory
from frugal.tornado.server.http_handler import FHttpHandler

sys.path.append('gen-py.tornado')
from http_client import logging_middleware
from v1.music.f_Store import Processor as FStoreProcessor  # noqa
from v1.music.f_Store import Iface  # noqa
from v1.music.ttypes import Album, Track, PerfRightsOrg  # noqa


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
        album.ASIN = str(uuid.uuid4())
        album.duration = 12000
        album.tracks = [Track(title="Comme des enfants",
                              artist="Coeur de pirate",
                              publisher="Grosse Boîte",
                              composer="Béatrice Martin",
                              duration=169,
                              pro=PerfRightsOrg.ASCAP)]

        return album

    def enterAlbumGiveaway(self, ctx, email, name):
        """
        Always return success (true)
        """
        return True


if __name__ == "__main__":
    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between clients and servers.
    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    # Create a new server processor.
    # Incoming requests to the processor are passed to the handler.
    # Results from the handler are returned back to the client.
    processor = FStoreProcessor(StoreHandler())

    # Optionally add middleware to the processor before starting the server.
    # add_middleware can take a list or single middleware.
    processor.add_middleware(logging_middleware)

    # Create a new music store server using the a tornado handler
    # and our configured processor and protocol
    application = tornado.web.Application([
        (r"/frugal",
            FHttpHandler,
            dict(processor=processor, protocol_factory=prot_factory))
    ])
    application.listen(9090)
    tornado.ioloop.IOLoop.current().start()
