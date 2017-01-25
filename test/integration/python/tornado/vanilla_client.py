from __future__ import print_function

import argparse
import sys

sys.path.append('gen-py')
sys.path.append('..')

from frugal.context import FContext
from frugal.provider import FServiceProvider
from frugal.transport.http_transport import THttpTransport

from common.test_definitions import rpc_test_definitions
from common.utils import *
from frugal_test.f_FrugalTest import Client as FrugalTestClient

middleware_called = False


def main():
    parser = argparse.ArgumentParser(description="Run a vanilla python client")
    parser.add_argument('--port', dest='port', default='9090')
    parser.add_argument('--protocol', dest='protocol_type', default="binary", choices="binary, compact, json")
    parser.add_argument('--transport', dest='transport_type', default="http", choices="http")

    args = parser.parse_args()

    protocol_factory = get_protocol_factory(args.protocol_type)

    if args.transport_type == "http":
        transport = THttpTransport("http://localhost:" + str(args.port))
    else:
        print("Unknown transport type: {}".format(args.transport_type))
        sys.exit(1)

    transport.open()

    ctx = FContext("test")
    client = FrugalTestClient(FServiceProvider(transport, protocol_factory), client_middleware)

    # Scope generation is not currently supported with vanilla python
    # TODO: Add Pub/Sub test once scopes are supported
    test_rpc(client, ctx)

    global middleware_called
    if not middleware_called:
        print("Client middleware never invoked")
        exit(1)

    transport.close()


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

    try:
        client.testOneway(ctx, 1)
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
        print("{}".format(ret))
        return ret
    return handler


if __name__ == '__main__':
    main()
