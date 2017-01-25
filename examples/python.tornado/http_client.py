import logging
import sys
import uuid

from tornado import ioloop, gen
from thrift.protocol import TBinaryProtocol
from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.provider import FServiceProvider
from frugal.tornado.transport import FHttpTransport
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

    # Create an HTTP transport using the connected client
    transport = FHttpTransport("http://localhost:9090/frugal")
    yield transport.open()

    # Using the configured transport and protocol, create a client
    # to talk to the music store service.
    store_client = FStoreClient(FServiceProvider(transport, prot_factory),
                                middleware=logging_middleware)

    album = yield store_client.buyAlbum(FContext(),
                                        str(uuid.uuid4()),
                                        "ACT-12345")

    root.info("Bought an album %s\n", album)

    yield store_client.enterAlbumGiveaway(FContext(),
                                          "kevin@workiva.com",
                                          "Kevin")

    yield transport.close()


def logging_middleware(next):
    def handler(method, args):
        service = '%s.%s' % (method.im_self.__module__,
                             method.im_class.__name__)
        root.info('==== CALLING %s.%s ===='.format(service,
                                                   method.im_func.func_name))
        ret = next(method, args)
        root.info('==== CALLED  %s.%s ===='.format(service,
                                                   method.im_func.func_name))
        return ret
    return handler


if __name__ == '__main__':
    # Since we can exit after the client calls use `run_sync`
    ioloop.IOLoop.instance().run_sync(main)
