import asyncio
import sys
import argparse

sys.path.append('gen_py_asyncio')
sys.path.append('..')

from frugal.context import FContext
from frugal.provider import FScopeProvider
from frugal.aio.transport import (
    FNatsTransport,
    FHttpTransport,
    FNatsScopeTransportFactory,
)

from nats.aio.client import Client as NatsClient

from frugal_test.f_Events_publisher import EventsPublisher
from frugal_test import ttypes, Xception, Insanity, Xception2, Event
from frugal_test.f_Events_subscriber import EventsSubscriber
from frugal_test.f_FrugalTest import Client as FrugalTestClient

from common.utils import *
from common.test_definitions import rpc_test_definitions


response_received = False

async def main():

    parser = argparse.ArgumentParser(description="Run a python asyncio client")
    parser.add_argument('--port', dest='port', default=9090)
    parser.add_argument('--protocol', dest='protocol_type', default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest='transport_type', default="stateless", choices="stateless, stateless-stateful, http")
    args = parser.parse_args()

    protocol_factory = get_protocol_factory(args.protocol_type)

    nats_client = NatsClient()
    await nats_client.connect(**get_nats_options())

    transport = None

    if args.transport_type in ["stateless", "stateless-stateful"]:
        transport = FNatsTransport(nats_client, args.port)
        await transport.open()
    elif args.transport_type == "http":
        transport = FHttpTransport("http://localhost:{port}".format(port=args.port))
    else:
        print("Unknown transport type: {type}".format(type=args.transport_type))
        sys.exit(1)

    client = FrugalTestClient(transport, protocol_factory)
    ctx = FContext("test")

    await test_rpc(client, ctx)
    await test_pub_sub(nats_client, protocol_factory, args.port)

    await nats_client.close()


async def test_rpc(client, ctx):
    test_failed = False

    # Iterate over all expected RPC results
    for rpc, vals in rpc_test_definitions().items():
        method = getattr(client, rpc)
        args = vals['args']
        expected_result = vals['expected_result']

        result = None
        try:
            if args:
                result = await method(ctx, *args)
            else:
                result = await method(ctx)
        except Exception as e:
            result = e

        test_failed = check_for_failure(result, expected_result) or test_failed


    # oneWay RPC call (no response)
    seconds = 1
    try:
        await client.testOneway(ctx, seconds)
    except Exception as e:
        print("Unexpected error in testOneway() call: {}".format(e))
        test_failed = True

    if test_failed:
        exit(1)


# test_pub_sub publishes an event and verifies that a response is received
async def test_pub_sub(nats_client, protocol_factory, port):
    global response_received
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, protocol_factory)
    publisher = EventsPublisher(provider)

    await publisher.open()

    def subscribe_handler(context, event):
        print("Response received {}".format(event))
        global response_received
        if context:
            response_received = True

    # Subscribe to response
    subscriber = EventsSubscriber(provider)
    await subscriber.subscribe_EventCreated("{}-response".format(port), subscribe_handler)

    event = Event(Message="Sending Call")
    context = FContext("Call")
    print("Publishing...")
    await publisher.publish_EventCreated(context, "{}-call".format(port), event)

    # Loop with sleep interval. Fail if not received within 3 seconds
    total_time = 0
    interval = 0.1
    while total_time < 3:
        if response_received:
            break
        else:
            await asyncio.sleep(interval)
            total_time += interval

    if not response_received:
        print("Pub/Sub response timed out!")
        exit(1)

    await publisher.close()
    exit(0)


if __name__ == '__main__':
    io_loop = asyncio.get_event_loop()
    io_loop.run_until_complete(main())
