from __future__ import print_function

import SimpleHTTPServer
import SocketServer
import argparse
import logging
import sys
import thread

sys.path.append('..')
sys.path.append('gen_py_tornado')


from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.provider import FScopeProvider

from frugal_test.f_Events_publisher import EventsPublisher
from frugal_test.f_Events_subscriber import EventsSubscriber
from frugal_test.f_FrugalTest import Processor
from frugal_test.ttypes import Event

from frugal.tornado.server import FNatsServer, FHttpHandler
from frugal.tornado.transport import FNatsPublisherTransportFactory
from frugal.tornado.transport import FNatsSubscriberTransportFactory

from common.FrugalTestHandler import FrugalTestHandler
from common.utils import *

from nats.io.client import Client as NATS
from thrift.protocol.TBinaryProtocol import TBinaryProtocolFactory
from thrift.protocol.TCompactProtocol import TCompactProtocolFactory
from thrift.protocol.TJSONProtocol import TJSONProtocolFactory
from tornado import ioloop, gen
from tornado.web import Application


publisher = None
port = 0


@gen.coroutine
def main():
    parser = argparse.ArgumentParser(description="Run a tornado python server")
    parser.add_argument('--port', dest='port', default='9090')
    parser.add_argument('--protocol', dest='protocol_type',
                        default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest="transport_type",
                        default="stateless", choices="stateless, http")

    args = parser.parse_args()

    if args.protocol_type == "binary":
        protocol_factory = FProtocolFactory(TBinaryProtocolFactory())
    elif args.protocol_type == "compact":
        protocol_factory = FProtocolFactory(TCompactProtocolFactory())
    elif args.protocol_type == "json":
        protocol_factory = FProtocolFactory(TJSONProtocolFactory())
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
    subject = "frugal.*.*.rpc.{}".format(args.port)
    processor = Processor(handler)

    if args.transport_type == "stateless":
        server = FNatsServer(nats_client, [subject],
                                    processor, protocol_factory)

        # start healthcheck so the test runner knows the server is running
        thread.start_new_thread(healthcheck, (port, ))
        print("Starting {} server...".format(args.transport_type))
        yield server.serve()

    elif args.transport_type == "http":
        factories = {'processor': processor,
                     'protocol_factory': protocol_factory}

        server = Application([(r'/', FHttpHandler, factories)])

        print("Starting {} server...".format(args.transport_type))
        server.listen(port)

    else:
        logging.error("Unknown transport type: %s", args.transport_type)
        sys.exit(1)

    # Setup subscriber, send response upon receipt
    pub_transport_factory = FNatsPublisherTransportFactory(nats_client)
    sub_transport_factory = FNatsSubscriberTransportFactory(nats_client)
    provider = FScopeProvider(
        pub_transport_factory, sub_transport_factory, protocol_factory)
    global publisher
    publisher = EventsPublisher(provider)
    yield publisher.open()

    @gen.coroutine
    def response_handler(context, event):
        print("received {} : {}".format(context, event))
        preamble = context.get_request_header(PREAMBLE_HEADER)
        if preamble is None or preamble == "":
            logging.error("Client did not provide preamble header")
            return
        ramble = context.get_request_header(RAMBLE_HEADER)
        if ramble is None or ramble == "":
            logging.error("Client did not provide ramble header")
            return
        response_event = Event(Message="Sending Response")
        response_context = FContext("Call")
        global publisher
        global port
        yield publisher.publish_EventCreated(response_context, preamble, ramble, "response", "{}".format(port), response_event)
        print("Published event={}".format(response_event))

    subscriber = EventsSubscriber(provider)
    yield subscriber.subscribe_EventCreated("*", "*", "call", "{}".format(args.port), response_handler)


def healthcheck(port):
    health_handler = SimpleHTTPServer.SimpleHTTPRequestHandler
    healthcheck = SocketServer.TCPServer(("", int(port)), health_handler)
    healthcheck.serve_forever()


if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
