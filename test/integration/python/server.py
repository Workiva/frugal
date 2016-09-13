from __future__ import print_function

import SimpleHTTPServer
import SocketServer
import argparse
import logging
import sys
import thread

sys.path.append('gen_py_tornado')


from frugal.context import FContext
from frugal.processor import FProcessorFactory
from frugal.protocol import FProtocolFactory
from frugal.provider import FScopeProvider

from frugal_test.f_Events_publisher import EventsPublisher
from frugal_test.f_Events_subscriber import EventsSubscriber
from frugal_test.f_FrugalTest import Processor
from frugal_test.ttypes import Event

from frugal.tornado.server import FNatsTornadoServer, FStatelessNatsTornadoServer, FTornadoHttpHandler
from frugal.tornado.transport import (
    FMuxTornadoTransportFactory,
    FNatsScopeTransportFactory,
    TNatsServiceTransport,
)

from FrugalTestHandler import FrugalTestHandler

from nats.io.client import Client as NATS
from thrift.protocol import TBinaryProtocol, TCompactProtocol, TJSONProtocol
from tornado import ioloop, gen
from tornado.web import Application


publisher = None
port = 0


@gen.coroutine
def main():
    parser = argparse.ArgumentParser(description="Run a python server")
    parser.add_argument('--port', dest='port', default=9090)
    parser.add_argument('--protocol', dest='protocol_type', default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest="transport_type", default="stateless", choices="stateless, stateful, stateless-stateful, http")

    args = parser.parse_args()

    if args.protocol_type == "binary":
        protocol_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    elif args.protocol_type == "compact":
        protocol_factory = FProtocolFactory(TCompactProtocol.TCompactProtocolFactory())
    elif args.protocol_type == "json":
        protocol_factory = FProtocolFactory(TJSONProtocol.TJSONProtocolFactory())
    else:
        logging.error("Unknown protocol type: %s", args.protocol_type)
        sys.exit(1)

    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }
    yield nats_client.connect(**options)

    global port
    port = args.port

    handler = FrugalTestHandler()
    subject = args.port
    processor = Processor(handler)

    if args.transport_type == "stateless":
        server = FStatelessNatsTornadoServer(nats_client,
                                             subject,
                                             processor,
                                             protocol_factory)
        # start healthcheck so the test runner knows the server is running
        thread.start_new_thread(healthcheck, (port, ))
        print("Starting {} server...".format(args.transport_type))
        yield server.serve()
    elif args.transport_type == "stateful" or args.transport_type == "stateless-stateful":  # @Deprecated TODO: Remove in 2.0
        transport_factory = FMuxTornadoTransportFactory()
        heartbeat_interval = 10000
        max_missed_heartbeats = 3
        server = FNatsTornadoServer(nats_client,
                                    subject,
                                    max_missed_heartbeats,
                                    FProcessorFactory(Processor(handler)),
                                    transport_factory,
                                    protocol_factory,
                                    heartbeat_interval)
        # start healthcheck so the test runner knows the server is running
        thread.start_new_thread(healthcheck, (port, ))
        print("Starting {} server...".format(args.transport_type))
        yield server.serve()
    elif args.transport_type == "http":
        server = Application([
            (r'/', FTornadoHttpHandler, dict(processor=processor, protocol_factory=protocol_factory))
            ])
        print("Starting {} server...".format(args.transport_type))
        server.listen(port)
    else:
        logging.error("Unknown transport type: %s", args.transport_type)
        sys.exit(1)

    # Setup subscriber, send response upon receipt
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, protocol_factory)
    global publisher
    publisher = EventsPublisher(provider)
    yield publisher.open()

    def response_handler(context, event):
        print("received {} : {}".format(context, event))
        response_event = Event(Message="Sending Response")
        response_context = FContext("Call")
        global publisher
        global port
        publisher.publish_EventCreated(response_context, "{}-response".format(port), response_event)
        print("Published event={}".format(response_event))

    subscriber = EventsSubscriber(provider)
    yield subscriber.subscribe_EventCreated("{}-call".format(args.port), response_handler)


def healthcheck(port):
    health_handler = SimpleHTTPServer.SimpleHTTPRequestHandler
    healthcheck = SocketServer.TCPServer(("", int(port)), health_handler)
    healthcheck.serve_forever()


if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
