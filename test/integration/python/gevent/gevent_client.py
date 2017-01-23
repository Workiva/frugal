from __future__ import print_function

import sys
import argparse
import gevent

sys.path.append('gen_py_gevent')
sys.path.append('..')

from frugal.context import FContext
from frugal.provider import FScopeProvider

from frugal.gevent.transport import (
    FNatsPublisherTransportFactory,
    FNatsSubscriberTransportFactory,
    FNatsTransport
)

from frugal_test.ttypes import Xception, Insanity, Xception2, Event
from frugal_test.f_Events_publisher import EventsPublisher
from frugal_test.f_Events_subscriber import EventsSubscriber
from frugal_test.f_FrugalTest import Client as FrugalTestClient

from gnats import Client as NATS
from thrift.transport.TTransport import TTransportException

from common.utils import *
from common.test_definitions import rpc_test_definitions


response_received = False
middleware_called = False


def main():
    parser = argparse.ArgumentParser(description="Run a python gevent client")
    parser.add_argument('--port', dest='port', default= '9090')
    parser.add_argument('--protocol', dest='protocol_type', default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest='transport_type', default="stateless",
                        choices="stateless, http")

    args = parser.parse_args()

    protocol_factory = get_protocol_factory(args.protocol_type)

    nats_client = NATS()

    logging.debug("Connecting to NATS")
    nats_client.connect(**get_nats_options())

    transport = None

    if args.transport_type == "stateless":
        transport = FNatsTransport(nats_client, str(args.port))
    # elif args.transport_type == "http":
    #     transport = FHttpTransport("http://localhost:" + str(args.port))
    else:
        print("Unknown transport type: {}".format(args.transport_type))
        sys.exit(1)

    try:
        transport.open()
    except TTransportException as ex:
        logging.error(ex)
        raise ex


    client = FrugalTestClient(transport, protocol_factory, client_middleware)

    ctx = FContext("test")

    test_rpc(client, ctx)
    test_pub_sub(nats_client, protocol_factory, args.port)

    global middleware_called
    if not middleware_called:
        print("Client middleware never invoked")
        exit(1)

    # Cleanup after tests
    nats_client.close()


# test_pub_sub publishes an event and verifies that a response is received
def test_pub_sub(nats_client, protocol_factory, port):
    global response_received
    pub_transport_factory = FNatsPublisherTransportFactory(nats_client)
    sub_transport_factory = FNatsSubscriberTransportFactory(nats_client)
    provider = FScopeProvider(
        pub_transport_factory, sub_transport_factory, protocol_factory)
    publisher = EventsPublisher(provider)
    publisher.open()

    def subscribe_handler(context, event):
        print("Response received {}".format(event))
        global response_received
        if context:
            response_received = True

    # Subscribe to response
    preamble = "foo"
    ramble = "bar"
    subscriber = EventsSubscriber(provider)
    subscriber.subscribe_EventCreated(preamble, ramble, "response", "{}".format(port), subscribe_handler)

    event = Event(Message="Sending Call")
    context = FContext("Call")
    context.set_request_header(PREAMBLE_HEADER, preamble)
    context.set_request_header(RAMBLE_HEADER, ramble)
    print("Publishing...")
    publisher.publish_EventCreated(context, preamble, ramble, "call", "{}".format(port), event)

    # Loop with sleep interval. Fail if not received within 3 seconds
    total_time = 0
    interval = .1
    while total_time < 3:
        if response_received:
            break
        else:
            gevent.sleep(interval)
            total_time += interval

    if not response_received:
        print("Pub/Sub response timed out!")
        exit(1)

    publisher.close()
    exit(0)


# test_rpc makes RPC calls with each type defined in FrugalTest.frugal
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
                result = method(ctx, *args)
            else:
                result = method(ctx)
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
        log_future(ret)
        return ret
    return handler


def log_future(future):
    try:
        print("{}".format(future))
    except Exception as ex:
        print("{}".format(ex))


if __name__ == '__main__':
    gevent.joinall([gevent.spawn(main)])
