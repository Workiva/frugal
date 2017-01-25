import os
import logging
import sys
import uuid

from thrift.protocol import TBinaryProtocol

from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.provider import FServiceProvider
from frugal.transport.http_transport import THttpTransport

sys.path.append(os.path.join(os.path.dirname(__file__), "gen-py"))
from v1.music.f_Store import Client as FStoreClient  # noqa
from v1.music.ttypes import Album  # noqa


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


URL = 'http://localhost:9090'


def main():
    logging.info("Starting...")

    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between clients and servers.
    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    # Create an HTTP transport for the server URL
    transport = THttpTransport(URL)
    transport.open()

    # Using the configured transport and protocol, create a client
    # to talk to the music store service.
    store_client = FStoreClient(FServiceProvider(transport, prot_factory))

    album = store_client.buyAlbum(FContext(),
                                  str(uuid.uuid4()),
                                  "ACT-12345")

    root.info("Bought an album %s\n", album)

    store_client.enterAlbumGiveaway(FContext(),
                                    "kevin@workiva.com",
                                    "Kevin")

    # Close the transport
    transport.close()


def logging_middleware(next):
    def handler(method, args):
        root.info('==== CALLING %s ====', method.__name__)
        ret = next(method, args)
        root.info('==== CALLED  %s ====', method.__name__)
        return ret
    return handler


if __name__ == '__main__':
    main()
