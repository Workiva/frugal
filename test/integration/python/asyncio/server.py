import argparse
import socketserver
import sys
import threading
import asyncio

from aiohttp import web
import http

sys.path.append('..')
sys.path.append('gen_py_asyncio')


from frugal.context import FContext
from frugal.provider import FScopeProvider

from frugal_test.f_Events_publisher import EventsPublisher
from frugal_test.f_Events_subscriber import EventsSubscriber
from frugal_test.f_FrugalTest import Processor
from frugal_test.ttypes import Event

from frugal.aio.server import FNatsServer
from frugal.aio.server.http_handler import new_http_handler
from frugal.aio.transport.nats_scope_transport import FNatsScopeTransportFactory

from common.FrugalTestHandler import FrugalTestHandler
from common.utils import *

from nats.aio.client import Client as NatsClient


publisher = None
port = 0


async def main():
    parser = argparse.ArgumentParser(description="Run an asyncio python server")
    parser.add_argument('--port', dest='port', default=9090)
    parser.add_argument('--protocol', dest='protocol_type', default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest="transport_type", default="stateless", choices="stateless, stateless-stateful, http")

    args = parser.parse_args()

    protocol_factory = get_protocol_factory(args.protocol_type)

    nats_client = NatsClient()
    await nats_client.connect(**get_nats_options())

    global port
    port = args.port

    handler = FrugalTestHandler()
    subject = args.port
    processor = Processor(handler)

    # Setup subscriber, send response upon receipt
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, protocol_factory)
    global publisher
    publisher = EventsPublisher(provider)
    await publisher.open()

    async def response_handler(context, event):
        response_event = Event(Message="Sending Response")
        response_context = FContext("Call")
        global publisher
        global port
        await publisher.publish_EventCreated(response_context, "{}-response".format(port), response_event)

    subscriber = EventsSubscriber(provider)
    await subscriber.subscribe_EventCreated("{}-call".format(args.port), response_handler)

    if args.transport_type in ["stateless", "stateless-stateful"]:
        server = FNatsServer(nats_client,
                             subject,
                             processor,
                             protocol_factory)
        # start healthcheck so the test runner knows the server is running
        threading.Thread(target=healthcheck,
            args=(port,)
        ).start()

        print("Starting {} server...".format(args.transport_type))
        await server.serve()

    elif args.transport_type == "http":
        print('starting http server')
        handler = new_http_handler(processor, protocol_factory)
        app = web.Application(loop=asyncio.get_event_loop())
        app.router.add_route("*", "/", handler)
        srv = await asyncio.get_event_loop().create_server(
            app.make_handler(), '0.0.0.0', port)

    else:
        logging.error("Unknown transport type: %s", args.transport_type)
        sys.exit(1)


def healthcheck(port):
    health_handler = http.server.SimpleHTTPRequestHandler
    healthcheck = socketserver.TCPServer(("", int(port)), health_handler)
    healthcheck.serve_forever()


if __name__ == '__main__':
    io_loop = asyncio.get_event_loop()
    asyncio.ensure_future(main())
    io_loop.run_forever()
