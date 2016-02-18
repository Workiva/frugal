import unittest
from frugal.protocol.protocol import FProtocol
from frugal.context import FContext
from thrift.transport.TTransport import TTransportBase


class TestFProtocol(unittest.TestCase):

    def test_write_header(self):

        trans = TTransportBase()
        protocol = FProtocol(trans)
        context = FContext("fooid")
        context.put_request_header("foo", "bar")

        headers = context.get_request_headers()

        buff = protocol._write_header(headers)

        parsed_headers = protocol._read_header(buff)

        self.assertEquals("fooid", parsed_headers['_cid'])
        self.assertEquals("1", parsed_headers['_opid'])
        self.assertEquals("bar", parsed_headers['foo'])
