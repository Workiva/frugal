import logging
import sys
sys.path.append('gen-py.asyncio')
sys.path.append('example_handler.py')

from thrift.protocol import TBinaryProtocol

import asyncio

from aiohttp import web
from nats.aio.client import Client as NatsClient

from frugal.protocol import FProtocolFactory
from frugal.aio.server import FNatsServer

from event.f_Foo import Processor as FFooProcessor
from example_handler import ExampleHandler


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


async def run_nats_server():
    nats_client = NatsClient()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    await nats_client.connect(**options)

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    handler = ExampleHandler()
    processor = FFooProcessor(handler)

    subject = "foo"
    server = FNatsServer(nats_client, subject, processor, prot_factory)

    root.info("Starting Nats rpc server...")

    await server.serve()


if __name__ == '__main__':
    io_loop = asyncio.get_event_loop()
    asyncio.ensure_future(run_nats_server())
    io_loop.run_forever()
