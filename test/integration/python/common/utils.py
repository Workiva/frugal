
import logging
import sys

from thrift.protocol.TBinaryProtocol import TBinaryProtocolFactory
from thrift.protocol.TJSONProtocol import TJSONProtocolFactory
from thrift.protocol.TCompactProtocol import TCompactProtocolFactory

from frugal.protocol import FProtocolFactory


def get_nats_options():
    return {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
        }


def get_protocol_factory(protocol):
    """
    Returns a protocol factory associated with the string protocol passed in
    as a command line argument to the cross runner

    :param protocol: string
    :return: Protocol factory
    """
    if protocol == "binary":
        return FProtocolFactory(TBinaryProtocolFactory())
    elif protocol == "compact":
        return FProtocolFactory(TCompactProtocolFactory())
    elif protocol == "json":
        return FProtocolFactory(TJSONProtocolFactory())
    else:
        logging.error("Unknown protocol type: %s", protocol)
        sys.exit(1)


def check_for_failure(actual, expected):
    """
    Compares the actual return results with the expected results.

    :return: Bool reflecting failure status

    """
    if expected != actual:
        print("Unexpected result, expected:\n{e}\n but received:\n{a} ".format(e=expected, a=actual))
        return True
    return False
