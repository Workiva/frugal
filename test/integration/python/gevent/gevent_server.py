from __future__ import print_function

import SimpleHTTPServer
import SocketServer
import argparse
import logging
import sys
import thread
import gevent

sys.path.append('..')
sys.path.append('gen_py_gevent')


from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.provider import FScopeProvider

from frugal_test.f_Events_publisher import EventsPublisher
from frugal_test.f_Events_subscriber import EventsSubscriber
from frugal_test.f_FrugalTest import Processor
from frugal_test.ttypes import Event

from frugal.gevent.server import FNatsServer
from frugal.gevent.transport import FNatsPublisherTransportFactory
from frugal.gevent.transport import FNatsSubscriberTransportFactory

from common.FrugalTestHandler import FrugalTestHandler
from common.utils import *

from gnats import Client as NATS


publisher = None
port = 0


def main():
    parser = argparse.ArgumentParser(description="Run a gevent python server")
    parser.add_argument('--port', dest='port', default='9090')
    parser.add_argument('--protocol', dest='protocol_type',
                        default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest="transport_type",
                        default="stateless", choices="stateless, http")

    args = parser.parse_args()

    protocol_factory = get_protocol_factory(args.protocol_type)

    nats_client = NATS()
    nats_client.connect(**get_nats_options())

    global port
    port = args.port

    handler = FrugalTestHandler()
    subject = args.port
    processor = Processor(handler)

    def run_gevent_server():
        if args.transport_type == "stateless":
            server = FNatsServer(nats_client, [subject],
                                 processor, protocol_factory)

            # start healthcheck so the test runner knows the server is running
            thread.start_new_thread(healthcheck, (port, ))
            print("Starting {} server...".format(args.transport_type))
            server.serve()

        else:
            logging.error("Unknown transport type: %s", args.transport_type)
            sys.exit(1)

    def run_gevent_subscriber():
        # Setup subscriber, send response upon receipt
        pub_transport_factory = FNatsPublisherTransportFactory(nats_client)
        sub_transport_factory = FNatsSubscriberTransportFactory(nats_client)
        provider = FScopeProvider(
            pub_transport_factory, sub_transport_factory, protocol_factory)
        global publisher
        publisher = EventsPublisher(provider)
        publisher.open()

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
            publisher.publish_EventCreated(response_context, preamble, ramble, "response", "{}".format(port), response_event)
            print("Published event={}".format(response_event))

        subscriber = EventsSubscriber(provider)
        subscriber.subscribe_EventCreated("*", "*", "call", "{}".format(args.port), response_handler)

    gevent.joinall([gevent.spawn(run_gevent_server),
                    gevent.spawn(run_gevent_subscriber)])


def healthcheck(port):
    health_handler = SimpleHTTPServer.SimpleHTTPRequestHandler
    healthcheck = SocketServer.TCPServer(("", int(port)), health_handler)
    healthcheck.serve_forever()


if __name__ == '__main__':
    gevent.joinall([gevent.spawn(main)])
