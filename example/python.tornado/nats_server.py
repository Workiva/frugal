import logging
import sys

from thrift.protocol import TBinaryProtocol

from tornado import gen, ioloop

from nats.io.client import Client as NATS

from frugal.protocol import FProtocolFactory
from frugal.tornado.server import FStatelessNatsTornadoServer

sys.path.append('gen-py.tornado')
sys.path.append('example_handler.py')

from music.f_Store import Processor as FStoreProcessor  # noqa
from example_handler import ExampleHandler  # noqa


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

    # Open a NATS connection to receive requests
    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    yield nats_client.connect(**options)

    handler = ExampleHandler()
    processor = FStoreProcessor(handler)
    subject = "music-service"
    server = FStatelessNatsTornadoServer(nats_client,
                                         subject,
                                         processor,
                                         prot_factory)

    root.info("Starting server...")

    yield server.serve()

if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
