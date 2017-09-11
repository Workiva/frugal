import logging
import sys
import uuid

from nats.io.client import Client as NATS
from tornado import ioloop, gen
from thrift.protocol import TBinaryProtocol
from thrift.transport.TTransport import TTransportException
from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.provider import FServiceProvider
from frugal.tornado.transport import FNatsTransport
sys.path.append('gen-py.tornado')

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


@gen.coroutine
def main():
    # Declare the protocol stack used for serialization.
    # Protocol stacks must match between clients and servers.
    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    # Open a NATS connection to send requests
    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }
    yield nats_client.connect(**options)

    # Create a nats transport using the connected client
    # The transport sends data on the music-service NATS topic
    nats_transport = FNatsTransport(nats_client, "music-service")

    try:
        yield nats_transport.open()
    except TTransportException as ex:
        root.error(ex)
        raise gen.Return()

    # Using the configured transport and protocol, create a client
    # to talk to the music store service.
    store_client = FStoreClient(FServiceProvider(nats_transport, prot_factory),
                                middleware=logging_middleware)

    album = yield store_client.buyAlbum(FContext(),
                                        str(uuid.uuid4()),
                                        "ACT-12345")

    root.info("Bought an album %s\n", album)

    yield store_client.enterAlbumGiveaway(FContext(),
                                          "kevin@workiva.com",
                                          "Kevin")

    yield nats_transport.close()
    yield nats_client.close()


def logging_middleware(next):
    def handler(method, args):
        service = '%s.%s' % (method.im_self.__module__,
                             method.im_class.__name__)
        print '==== CALLING %s.%s ====' % (service, method.im_func.func_name)
        ret = next(method, args)
        print '==== CALLED  %s.%s ====' % (service, method.im_func.func_name)
        return ret
    return handler


if __name__ == '__main__':
    # Since we can exit after the client calls use `run_sync`
    ioloop.IOLoop.instance().run_sync(main)
