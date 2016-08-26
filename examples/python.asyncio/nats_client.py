import logging
import sys
import asyncio

from nats.aio.client import Client as NatsClient
from thrift.protocol import TBinaryProtocol
from thrift.transport.TTransport import TTransportException
from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.provider import FScopeProvider
from frugal.aio.transport import (
    FNatsTransport,
    FNatsScopeTransportFactory,
)

sys.path.append('gen-py.asyncio')
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


async def main():
    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between clients and servers.
    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    # Open a NATS connection to send requests
    nats_client = NatsClient()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }
    await nats_client.connect(**options)

    # Create a nats transport using the connected client
    # The transport sends data on the music-service NATS topic
    nats_transport = FNatsTransport(nats_client, "music-service")
    try:
        await nats_transport.open()
    except TTransportException as ex:
        root.error(ex)
        return

    # Using the configured transport and protocol, create a client
    # to talk to the music store service.
    store_client = FStoreClient(nats_transport, prot_factory,
                                middleware=logging_middleware)

    album = await store_client.buyAlbum(FContext(),
                                        uuid.uuid4(),
                                        "ACT-12345")

    root.info("Bought an album %s\n", album)

    await store_client.EnterAlbumGiveaway(FContext(),
                                          "kevin@workiva.com",
                                          "Kevin")

    await nats_transport.close()
    await nats_client.close()


def logging_middleware(next):
    def handler(method, args):
        print('==== CALLING %s ====', method.__name__)
        ret = next(method, args)
        print('==== CALLED  %s ====', method.__name__)
        return ret
    return handler


if __name__ == '__main__':
    io_loop = asyncio.get_event_loop()
    io_loop.run_until_complete(main())
