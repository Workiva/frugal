from __future__ import print_function

import sys
import argparse

sys.path.append('gen_py_tornado')
sys.path.append('..')

from frugal.context import FContext
from frugal.provider import FScopeProvider

from frugal.tornado.transport import (
    FNatsScopeTransportFactory,
    FNatsTransport,
    FHttpTransport
)

from frugal_test import ttypes, Xception, Insanity, Xception2, Event
from frugal_test.f_Events_publisher import EventsPublisher
from frugal_test.f_Events_subscriber import EventsSubscriber
from frugal_test.f_FrugalTest import Client as FrugalTestClient

from nats.io.client import Client as NATS
from thrift.transport.TTransport import TTransportException
from tornado import ioloop, gen

from common.utils import *
from common.test_definitions import rpc_test_definitions


response_received = False
middleware_called = False


@gen.coroutine
def main():
    parser = argparse.ArgumentParser(description="Run a python tornado client")
    parser.add_argument('--port', dest='port', default=9090)
    parser.add_argument('--protocol', dest='protocol_type', default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest='transport_type', default="stateless",
                        choices="stateless, http")

    args = parser.parse_args()

    protocol_factory = get_protocol_factory(args.protocol_type)

    nats_client = NATS()

    logging.debug("Connecting to NATS")
    yield nats_client.connect(**get_nats_options())

    transport = None

    if args.transport_type == "stateless":
        transport = FNatsTransport(nats_client, str(args.port))
    elif args.transport_type == "http":
        transport = FHttpTransport("http://localhost:" + str(args.port))
    else:
        print("Unknown transport type: {}".format(args.transport_type))
        sys.exit(1)

    try:
        yield transport.open()
    except TTransportException as ex:
        logging.error(ex)
        raise gen.Return()

    client = FrugalTestClient(transport, protocol_factory, client_middleware)

    ctx = FContext("test")

    yield test_rpc(client, ctx)
    yield test_pub_sub(nats_client, protocol_factory, args.port)

    global middleware_called
    if not middleware_called:
        print("Client middleware never invoked")
        exit(1)

    # Cleanup after tests
    yield nats_client.close()


# test_pub_sub publishes an event and verifies that a response is received
@gen.coroutine
def test_pub_sub(nats_client, protocol_factory, port):
    global response_received
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, protocol_factory)
    publisher = EventsPublisher(provider)
    yield publisher.open()

    def subscribe_handler(context, event):
        print("Response received {}".format(event))
        global response_received
        if context:
            response_received = True

    # Subscribe to response
    subscriber = EventsSubscriber(provider)
    yield subscriber.subscribe_EventCreated("{}-response".format(port), subscribe_handler)

    event = Event(Message="Sending Call")
    context = FContext("Call")
    print("Publishing...")
    publisher.publish_EventCreated(context, "{}-call".format(port), event)

    # Loop with sleep interval. Fail if not received within 3 seconds
    total_time = 0
    interval = .1
    while total_time < 3:
        if response_received:
            break
        else:
            yield gen.sleep(interval)
            total_time += interval

    if not response_received:
        print("Pub/Sub response timed out!")
        exit(1)

    yield publisher.close()
    exit(0)


# test_rpc makes RPC calls with each type defined in FrugalTest.frugal
@gen.coroutine
def test_rpc(client, ctx):
    test_failed = False

    # Iterate over all expected RPC results
    for rpc, vals in rpc_test_definitions().items():
        method = getattr(client, rpc)
        args = vals['args']
        expected_result = vals['expected_result']
        result = None

        try:
            if args:
                result = yield method(ctx, *args)
            else:
                result = yield method(ctx)
        except Exception as e:
            result = e

        test_failed = check_for_failure(result, expected_result) or test_failed

    # oneWay RPC call (no response)
    seconds = 1
    try:
        client.testOneway(ctx, seconds)
    except Exception as e:
        print("Unexpected error in testOneway() call: {}".format(e))
        test_failed = True

    if test_failed:
        exit(1)


def client_middleware(next):
    def handler(method, args):
        global middleware_called
        middleware_called = True
        print("{}({}) = ".format(method.im_func.func_name, args[1:]), end="")
        ret = next(method, args)
        ret.add_done_callback(log_future)
        return ret
    return handler


def log_future(future):
    try:
        print("{}".format(future.result()))
    except Exception as ex:
        print("{}".format(ex))


if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
