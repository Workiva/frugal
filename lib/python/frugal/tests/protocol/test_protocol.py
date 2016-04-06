import unittest
import mock
from thrift.protocol.TBinaryProtocol import TBinaryProtocolFactory
from thrift.transport.TTransport import TMemoryBuffer

from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.protocol.protocol import FProtocol
from frugal.context import FContext


class TestFProtocol(unittest.TestCase):

#   def test_write_header(self):
#
#        t_protocol_factory = TBinaryProtocolFactory()
#        f_protocol_factory = FProtocolFactory(t_protocol_factory)
#
#        transport = TMemoryBuffer()
#
#        protocol = f_protocol_factory.get_protocol(transport)
#
#        context = FContext("fooid")
#        context.set_request_header("foo", "bar")
#
#        headers = context.get_request_headers()
#
#        protocol._write_headers(headers)
#
#        parsed_headers = protocol._read_headers(transport.getvalue())
#
#        self.assertEquals("fooid", parsed_headers['_cid'])
#        self.assertEquals("bar", parsed_headers['foo'])

    def setUp(self):
        self.mock_wrapped_protocol = mock.Mock()

        self.protocol = FProtocol(self.mock_wrapped_protocol)

    def test_writeMessageBegin(self):
        self.protocol.writeMessageBegin("name", "type", 1)

        self.mock_wrapped_protocol.writeMessageBegin.assert_called_with("name",
                                                                        "type",
                                                                        1)

    def test_writeMessageEnd(self):
        self.protocol.writeMessageEnd()

        self.mock_wrapped_protocol.writeMessageEnd.assert_called_with()

    def test_writeStructBegin(self):
        self.protocol.writeStructBegin("foo")

        self.mock_wrapped_protocol.writeStructBegin.assert_called_with("foo")

    def test_writeStructEnd(self):
        self.protocol.writeStructEnd()

        self.mock_wrapped_protocol.writeStructEnd.assert_called_with()

    def test_writeFieldStop(self):
        self.protocol.writeFieldStop()

        self.mock_wrapped_protocol.writeFieldStop.assert_called_with()

    def test_readMessageBegin(self):
        self.protocol.readMessageBegin()

        self.mock_wrapped_protocol.readMessageBegin.assert_called_with()

    def test_readStructBegin(self):
        self.protocol.readStructBegin()

        self.mock_wrapped_protocol.readStructBegin.assert_called_with()

    def test_readFieldBegin(self):
        self.protocol.readFieldBegin()

        self.mock_wrapped_protocol.readFieldBegin.assert_called_with()

    def test_readField(self):
        self.protocol.readField()

        self.mock_wrapped_protocol.readField.assert_called_with()

    def test_readStructEnd(self):
        self.protocol.readStructEnd()

        self.mock_wrapped_protocol.readStructEnd.assert_called_with()

