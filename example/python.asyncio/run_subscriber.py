import logging
import sys
sys.path.append('gen-py.asyncio')

from thrift.protocol import TBinaryProtocol

import asyncio

from nats.aio.client import Client as NatsClient

from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.provider import FScopeProvider
from frugal.aio.transport import FNatsScopeTransportFactory

from event.f_Events_subscriber import EventsSubscriber


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


async def main():

    nats_client = NatsClient()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    await nats_client.connect(**options)

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, prot_factory)

    subscriber = EventsSubscriber(provider)

    def event_handler(ctx, req):
        root.info("Received an event with ID: {} and Message {}".format(req.ID, req.Message))

    await subscriber.subscribe_EventCreated("barUser", event_handler)

    root.info("Subscriber starting...")

if __name__ == '__main__':
    io_loop = asyncio.get_event_loop()
    asyncio.ensure_future(main())
    io_loop.run_forever()
