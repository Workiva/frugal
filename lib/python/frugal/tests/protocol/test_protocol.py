import unittest

from thrift.protocol.TBinaryProtocol import TBinaryProtocolFactory
from thrift.transport.TTransport import TMemoryBuffer

from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.context import FContext


class TestFProtocol(unittest.TestCase):

    def test_write_header(self):

        t_protocol_factory = TBinaryProtocolFactory()
        f_protocol_factory = FProtocolFactory(t_protocol_factory)

        transport = TMemoryBuffer()

        protocol = f_protocol_factory.get_protocol(transport)

        context = FContext("fooid")
        context.set_request_header("foo", "bar")

        headers = context.get_request_headers()

        protocol._write_headers(headers)

        parsed_headers = protocol._read_headers(transport.getvalue())

        self.assertEquals("fooid", parsed_headers['_cid'])
        self.assertEquals("bar", parsed_headers['foo'])
