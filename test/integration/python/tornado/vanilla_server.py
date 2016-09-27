from __future__ import print_function

import argparse
import sys
import thread

sys.path.append('..')
sys.path.append('gen-py')

from frugal.context import FContext
from frugal.provider import FScopeProvider
from frugal.server.http_server import FHttpServer
from frugal.tornado.transport import FNatsScopeTransportFactory
from nats.io.client import Client as NATS
from tornado import gen, ioloop

from common.FrugalTestHandler import FrugalTestHandler
from common.utils import *
from frugal_test.f_FrugalTest import Processor
# Explicitly importing from gen_py_tornado
from gen_py_tornado.frugal_test.f_Events_publisher import EventsPublisher
from gen_py_tornado.frugal_test.f_Events_subscriber import EventsSubscriber
from gen_py_tornado.frugal_test.ttypes import Event


@gen.coroutine
def main():
    parser = argparse.ArgumentParser(description="Run a python server")
    parser.add_argument('--port', dest='port', type=int, default=9090)
    parser.add_argument('--protocol', dest='protocol_type', default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest="transport_type", default="http", choices="http")

    args = parser.parse_args()

    protocol_factory = get_protocol_factory(args.protocol_type)

    handler = FrugalTestHandler()
    subject = args.port
    processor = Processor(handler)

    if args.transport_type != "http":
        print("Unknown transport type: {}".format(args.transport_type))
        sys.exit(1)

    # Start up thread for pub/sub, see notes below
    thread.start_new_thread(tornado_thread, (subject, protocol_factory))

    server = FHttpServer(processor, ('', subject), protocol_factory)
    print("Starting {} server...".format(args.transport_type))
    server.serve()


# Use the tornado pub/sub since vanilla python code generation doesn't support it
# Clients in the cross language tests will fail if they try to publish and don't receive a response
# TODO: Modify the crossrunner to support running tests with or without scopes
@gen.coroutine
def pub_sub(subject, protocol_factory):
    nats_client = NATS()
    yield nats_client.connect(**get_nats_options())

    # Setup subscriber, send response upon receipt
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, protocol_factory)
    publisher = EventsPublisher(provider)
    yield publisher.open()

    def response_handler(context, event):
        print("received {} : {}".format(context, event))
        response_event = Event(Message="Sending Response")
        response_context = FContext("Call")
        publisher.publish_EventCreated(response_context, "{}-response".format(subject), response_event)
        print("Published event={}".format(response_event))
        publisher.close()

    subscriber = EventsSubscriber(provider)
    yield subscriber.subscribe_EventCreated("{}-call".format(subject), response_handler)


def tornado_thread(subject, protocol_factory):
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(pub_sub, subject, protocol_factory)
    io_loop.start()


if __name__ == '__main__':
    main()
