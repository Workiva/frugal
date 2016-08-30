from __future__ import print_function

import argparse
import logging
import sys

sys.path.append('gen_py_tornado')

from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.provider import FScopeProvider

from frugal.tornado.transport import (
    FMuxTornadoTransportFactory,
    FNatsScopeTransportFactory,
    FNatsTransport,
    TNatsServiceTransport,
    FHttpTransport
)

from frugal_test import ttypes, Xception, Insanity, Xception2, Event
from frugal_test.f_Events_publisher import EventsPublisher
from frugal_test.f_Events_subscriber import EventsSubscriber
from frugal_test.f_FrugalTest import Xtruct, Xtruct2, Numberz
from frugal_test.f_FrugalTest import Client as FrugalTestClient

from nats.io.client import Client as NATS
from thrift.protocol import TBinaryProtocol, TCompactProtocol, TJSONProtocol
from thrift.transport.TTransport import TTransportException
from tornado import ioloop, gen


response_received = False


@gen.coroutine
def main():
    parser = argparse.ArgumentParser(description="Run a python client")
    parser.add_argument('--port', dest='port', default=9090)
    parser.add_argument('--protocol', dest='protocol_type', default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest='transport_type', default="stateless", choices="stateless, stateful, stateless-stateful, http")

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

    logging.debug("Connecting to NATS")
    yield nats_client.connect(**options)

    transport = None

    if args.transport_type == "stateless" or args.transport_type == "stateless-stateful":
        transport = FNatsTransport(nats_client, str(args.port))

    elif args.transport_type == "stateful":  # @Deprecated TODO: Remove in 2.0
        transport_factory = FMuxTornadoTransportFactory()
        nats_transport = TNatsServiceTransport.Client(
            nats_client=nats_client,
            connection_subject=str(args.port),
            connection_timeout=2000,
            io_loop=5)
        transport = transport_factory.get_transport(nats_transport)

    elif args.transport_type == "http":
        transport = FHttpTransport("http://localhost:" + str(args.port))
    else:
        print("Unknown transport type: {}".format(args.transport_type))

    try:
        yield transport.open()
    except TTransportException as ex:
        root.error(ex)
        raise gen.Return()

    client = FrugalTestClient(transport, protocol_factory)

    ctx = FContext("test")

    yield test_rpc(client, ctx)
    yield test_pub_sub(nats_client, protocol_factory, args.port)

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
    return_code = 0

    # RPC with no type
    print("test_void()", end="")
    yield client.testVoid(ctx)
    print(" = void")

    # RPC with string
    thing = "thing"
    print('testString("{}") = '.format(thing), end="")
    result = yield client.testString(ctx, "thing")
    if result != thing:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with boolean
    boolean = True
    print("testBool({}) = ".format(boolean), end="")
    result = yield client.testBool(ctx, boolean)
    if result != boolean:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with byte
    byte = 42
    print("testByte({}) = ".format(byte), end="")
    result = yield client.testByte(ctx, byte)
    if result != byte:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with i32
    i32 = 4242
    print("testI32({}) = ".format(i32), end="")
    result = yield client.testI32(ctx, i32)
    if result != i32:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with i64
    i64 = 424242
    print("testI64({}) = ".format(i64), end="")
    result = yield client.testI64(ctx, i64)
    if result != i64:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with double
    double = 42.42
    print("testDouble({}) = ".format(double), end="")
    result = yield client.testDouble(ctx, double)
    if result != double:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with binary
    binary = "0b101010"
    print("testBinary({}) = ".format(binary), end="")
    result = yield client.testBinary(ctx, binary)
    if result != binary:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # # RPC with Xtruct
    struct = Xtruct()
    struct.string_thing = thing
    struct.byte_thing = byte
    struct.i32_thing = i32
    struct.i64_thing = i64
    print("testStruct({}) = ".format(struct), end="")
    result = yield client.testStruct(ctx, struct)
    if result != struct:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with Xtruct2
    struct2 = Xtruct2()
    struct2.struct_thing = struct
    struct2.byte_thing = 0
    struct2.i32_thing = 0
    print("testNest({}) = ".format(struct2), end="")
    result = yield client.testNest(ctx, struct2)
    if result != struct2:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with map
    dictionary = {1: 2, 3: 4, 5: 42}
    print("testMap({}) = ".format(dictionary), end="")
    result = yield client.testMap(ctx, dictionary)
    if result != dictionary:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with map of strings
    string_map = {"a": "2", "b": "blah", "some": "thing"}
    print("testStringMap({}) = ".format(string_map), end="")
    result = yield client.testStringMap(ctx, string_map)
    if result != string_map:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with set
    set = {1, 2, 2, 42}
    print("testSet({}) = ".format(set), end="")
    result = yield client.testSet(ctx, set)
    if result != set:  # How to do this comparison?
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with list
    list = [1, 2, 42]
    print("testList({}) = ".format(list), end="")
    result = yield client.testList(ctx, list)
    if result != list:  # How to do this comparison?
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with enum
    enum = Numberz.TWO
    print("testEnum({}) = ".format(enum), end="")
    result = yield client.testEnum(ctx, enum)
    if result != enum:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with typeDef
    type_def = 42
    print("testTypedef({}) = ".format(type_def), end="")
    result = yield client.testTypedef(ctx,type_def)
    if result != type_def:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # # RPC with map of maps
    d = {4: 4, 3: 3, 2: 2, 1: 1}
    e = {-4: -4, -3: -3, -2: -2, -1: -1}
    mapmap = {-4: e, 4: d}
    print("testMapMap({}) = ".format(42), end="")
    result = yield client.testMapMap(ctx, 42)
    if result != mapmap:  # How to do this comparison?
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with Insanity (xtruct of xtructs)
    truck1 = Xtruct("Goodbye4", 4, 4, 4)
    truck2 = Xtruct("Hello2", 2, 2, 2)
    insanity = Insanity()
    insanity.userMap = {Numberz.FIVE: 5, Numberz.EIGHT: 8}
    insanity.xtructs = [truck1, truck2]
    print("testInsanity({}) = ".format(insanity), end="")
    result = yield client.testInsanity(ctx, insanity)
    expected_result = {1:
                     {2: Insanity(
                         xtructs=[Xtruct(string_thing='Goodbye4', byte_thing=4, i32_thing=4, i64_thing=4),
                                  Xtruct(string_thing='Hello2', byte_thing=2, i32_thing=2, i64_thing=2)],
                         userMap={8: 8, 5: 5}),
                      3: Insanity(
                         xtructs=[Xtruct(string_thing='Goodbye4', byte_thing=4, i32_thing=4, i64_thing=4),
                                  Xtruct(string_thing='Hello2', byte_thing=2, i32_thing=2, i64_thing=2)],
                         userMap={8: 8, 5: 5})}, 2: {}}
    if result != expected_result:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with Multi type
    multi = Xtruct()
    multi.string_thing = "Hello2"
    multi.byte_thing = 42
    multi.i32_thing = 4242
    multi.i64_thing = 424242
    print("testMulti() = ", end="")
    result = yield client.testMulti(ctx, 42, 4242, 424242, {1: "blah", 2: "thing"}, Numberz.EIGHT, 24)
    if result != multi:
        print("\nUnexpected result ", end="")
        return_code = 1
    print(result)

    # RPC with Exception
    message = "Xception"
    print("testException({}) = ".format(message), end="")
    try:
        result = yield client.testException(ctx, message)
    except Xception as exception:
        if exception.errorCode == 1001 and exception.message == "Xception":
            print(result)
        else:
            print("\nUnexpected result {}".format(result), end="")
            return_code = 1

    # RPC with MultiException
    message = "Xception2"
    print("testMultiException({}) = ".format(message), end="")
    try:
        result = yield client.testMultiException(ctx, message, "ignoreme")
        print("\nUnexpected result {}".format(result), end="")
        return_code = 1
    except Xception as exception:
        print("\nUnexpected result {}".format(exception), end="")
        return_code = 1
    except Xception2 as exception:
        if exception.errorCode == 2002 and exception.struct_thing.string_thing == "This is an Xception2":
            print(exception)
        else:
            print("\nUnexpected result {}".format(exception), end="")
            return_code = 1

    # oneWay RPC call (no response)
    seconds = 1
    print("testOneway({}) = ".format(seconds), end="")
    try:
        client.testOneway(ctx, seconds)
        print(" - no error returned")
    except Exception as e:
        print("Unexpected error in testOneway() call: {}".format(e))
        return_code = 1

    if return_code:
        exit(1)


if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
