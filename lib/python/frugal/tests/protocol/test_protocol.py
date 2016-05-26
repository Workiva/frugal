import unittest
import mock

from frugal.protocol.protocol import FProtocol
from frugal.context import FContext, _OP_ID


class TestFProtocol(unittest.TestCase):

    def setUp(self):
        self.mock_wrapped_protocol = mock.Mock()

        self.protocol = FProtocol(self.mock_wrapped_protocol)

    @mock.patch('frugal.protocol.protocol._Headers._read')
    def test_read_request_headers(self, mock_read):
        headers = {_OP_ID: "0"}
        mock_read.return_value = headers

        ctx = self.protocol.read_request_headers()

        self.assertEqual(0, ctx._get_op_id())

    @mock.patch('frugal.protocol.protocol._Headers._read')
    def test_read_response_headers(self, mock_read):
        headers = {_OP_ID: "0", "_cid": "someid"}
        mock_read.return_value = headers

        context = FContext("someid")

        self.protocol.read_response_headers(context)

        self.assertEqual("0", context.get_response_header(_OP_ID))
        self.assertEqual("someid", context.get_response_header("_cid"))

    @mock.patch('frugal.protocol.protocol._Headers._write_to_bytearray')
    def test_write_request_headers(self, mock_write):
        context = FContext("foo")

        mock_write.return_value = "bar"

        mock_trans = mock.Mock()
        self.protocol.trans = mock_trans

        self.protocol.write_request_headers(context)

        mock_write.assert_called_with(context.get_request_headers())
        mock_trans.write.assert_called_with("bar")

    @mock.patch('frugal.protocol.protocol._Headers._write_to_bytearray')
    def test_write_response_headers(self, mock_write):
        context = FContext("foo")

        mock_write.return_value = "bar"

        mock_trans = mock.Mock()
        self.protocol.trans = mock_trans

        self.protocol.write_response_headers(context)

        mock_write.assert_called_with(context.get_response_headers())
        mock_trans.write.assert_called_with("bar")

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

    def test_readStructEnd(self):
        self.protocol.readStructEnd()

        self.mock_wrapped_protocol.readStructEnd.assert_called_with()

