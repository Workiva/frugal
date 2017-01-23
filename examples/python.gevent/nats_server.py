# -*- coding: utf-8 -*-
import logging
import sys
import uuid
import gevent

from thrift.protocol import TJSONProtocol


from gnats import Client as NATS

from frugal.protocol import FProtocolFactory
from frugal.gevent.server import FNatsServer

sys.path.append('gen-py.gevent')

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
                              publisher="Grosse Boite",
                              composer="Beatrice Martin",
                              duration=169,
                              pro=PerfRightsOrg.ASCAP)]

        print("buy album called")

        return album

    def enterAlbumGiveaway(self, ctx, email, name):
        """
        Always return success (true)
        """
        return True


def main():
    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between clients and servers.
    prot_factory = FProtocolFactory(TJSONProtocol.TJSONProtocolFactory())

    # Open a NATS connection to receive requests
    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    nats_client.connect(**options)

    # Create a new server processor.
    # Incoming requests to the processor are passed to the handler.
    # Results from the handler are returned back to the client.
    processor = FStoreProcessor(StoreHandler())

    # Create a new music store server using the processor,
    # The server will listen on the music-service NATS topic
    server = FNatsServer(nats_client, "music-service",
                         processor, prot_factory)

    root.info("Starting server...")

    server.serve()


if __name__ == '__main__':
    gevent.joinall([gevent.spawn(main)])