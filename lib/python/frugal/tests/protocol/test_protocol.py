import unittest
import mock

from thrift.protocol.TProtocol import TProtocolException
from thrift.protocol.TBinaryProtocol import TBinaryProtocol
from thrift.protocol.TCompactProtocol import TCompactProtocol
from thrift.protocol.TJSONProtocol import TJSONProtocol
from thrift.transport.THttpClient import THttpClient
from thrift.transport.TTransport import TMemoryBuffer

from frugal.protocol.protocol import FProtocol, FUniversalProtocol
from frugal.context import FContext, _OPID_HEADER, _CID_HEADER


class TestFProtocol(unittest.TestCase):

    def setUp(self):
        self.mock_wrapped_protocol = mock.Mock()

        self.protocol = FProtocol(self.mock_wrapped_protocol)

    @mock.patch('frugal.protocol.protocol._Headers._read')
    def test_read_request_headers(self, mock_read):
        headers = {_OPID_HEADER: "0", _CID_HEADER: "someid"}
        mock_read.return_value = headers

        ctx = self.protocol.read_request_headers()

        # The opid sent on the request headers and the opid received on the
        # request headers should be different to allow propagation
        self.assertNotEqual(
            headers[_OPID_HEADER], ctx.get_request_header(_OPID_HEADER))

        # The opid in the response headers should match the opid originally
        # sent on the request headers
        self.assertEqual(
            headers[_OPID_HEADER], ctx.get_response_header(_OPID_HEADER))
        self.assertEqual("someid", ctx.get_response_header(_CID_HEADER))

    @mock.patch('frugal.protocol.protocol._Headers._read')
    def test_read_response_headers(self, mock_read):
        headers = {_OPID_HEADER: "0", "_cid": "someid"}
        mock_read.return_value = headers

        context = FContext("someid")

        self.protocol.read_response_headers(context)

        # Ensure the opid is not set when the response headers are read in
        self.assertIsNone(context.get_response_header(_OPID_HEADER))
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


class TestFUniversalProtocolFactory(unittest.TestCase):
    def test_getProtocolUnsupported(self):
        trans = THttpClient('http://example.com', 8080, 'foo')
        factory = FUniversalProtocolFactory()
        self.assertRaises(ValueError, factory.getProtocol, trans)

    def test_getProtocol(self):
        trans = TMemoryBuffer()
        factory = FUniversalProtocolFactory()
        prot = factory.getProtocol(trans, strictRead=True,strictWrite=False,
                                   string_length_limit=1, container_length_limit=123)

        self.assertIsInstance(prot.binary_protocol, TBinaryProtocol)
        self.assertIsInstance(prot.compact_protocol, TCompactProtocol)
        self.assertIsInstance(prot.json_protocol, TJSONProtocol)

        self.assertEqual(trans, prot.binary_protocol.trans)
        self.assertEqual(trans, prot.compact_protocol.trans)
        self.assertEqual(trans, prot.json_protocol.trans)

        self.assertTrue(prot.binary_protocol.strictRead)
        self.assertFalse(prot.binary_protocol.strictWrite)
        self.assertEqual(prot.binary_protocol.string_length_limit, 1)
        self.assertEqual(prot.binary_protocol.container_length_limit, 123)
        self.assertEqual(prot.compact_protocol.string_length_limit, 1)
        self.assertEqual(prot.compact_protocol.container_length_limit, 123)


class TestFUniversalProtocol(unittest.TestCase):
    def setUp(self):
        self.trans = TMemoryBuffer()
        self.factory = FUniversalProtocolFactory()
        self.tuple = ('name', 1, 2)

    @mock.patch('thrift.protocol.TBinaryProtocol.TBinaryProtocol.readByte')
    @mock.patch('thrift.protocol.TBinaryProtocol.TBinaryProtocol.readMessageBegin')
    @mock.patch('thrift.protocol.TCompactProtocol.TCompactProtocol.readMessageBegin')
    @mock.patch('thrift.protocol.TJSONProtocol.TJSONProtocol.readMessageBegin')
    def test_readMessageBeginParsingBinary(self, mock_json, mock_compact, mock_binary, mock_byte):
        mock_binary.return_value = self.tuple
        mock_compact.side_effect = TProtocolException()
        mock_json.side_effect = TProtocolException()
        mock_byte.return_value = 0xff

        # With strictRead = False, binary should be attempted last
        prot = self.factory.getProtocol(self.trans, strictRead=False)
        name, ttype, seqid = prot.readMessageBegin()

        self.assertIsInstance(prot.prot, TBinaryProtocol)
        self.assertTrue(mock_compact.called)
        self.assertTrue(mock_json.called)
        self.assertTrue(mock_binary.called)
        self.assertEqual(name, self.tuple[0])
        self.assertEqual(ttype, self.tuple[1])
        self.assertEqual(seqid, self.tuple[2])

        self.assertEqual(mock_byte.return_value, prot.readByte())

        mock_compact.reset_mock()
        mock_binary.reset_mock()
        mock_json.reset_mock()

        # With strictRead = True, binary should now be attempted first
        prot = self.factory.getProtocol(self.trans, strictRead=True)
        name, ttype, seqid = prot.readMessageBegin()

        self.assertIsInstance(prot.prot, TBinaryProtocol)
        self.assertFalse(mock_compact.called)
        self.assertFalse(mock_json.called)
        self.assertTrue(mock_binary.called)
        self.assertEqual(name, self.tuple[0])
        self.assertEqual(ttype, self.tuple[1])
        self.assertEqual(seqid, self.tuple[2])

        self.assertEqual(mock_byte.return_value, prot.readByte())

    @mock.patch('thrift.protocol.TCompactProtocol.TCompactProtocol.readByte')
    @mock.patch('thrift.protocol.TBinaryProtocol.TBinaryProtocol.readMessageBegin')
    @mock.patch('thrift.protocol.TCompactProtocol.TCompactProtocol.readMessageBegin')
    @mock.patch('thrift.protocol.TJSONProtocol.TJSONProtocol.readMessageBegin')
    def test_readMessageBeginParsingCompact(self, mock_json, mock_compact, mock_binary, mock_byte):
        mock_binary.side_effect = TProtocolException()
        mock_compact.return_value = self.tuple
        mock_json.side_effect = TProtocolException()
        mock_byte.return_value = 0xff

        prot = self.factory.getProtocol(self.trans)
        name, ttype, seqid = prot.readMessageBegin()

        self.assertIsInstance(prot.prot, TCompactProtocol)
        self.assertFalse(mock_binary.called)
        self.assertTrue(mock_compact.called)
        self.assertFalse(mock_json.called)
        self.assertEqual(name, self.tuple[0])
        self.assertEqual(ttype, self.tuple[1])
        self.assertEqual(seqid, self.tuple[2])

        self.assertEqual(mock_byte.return_value, prot.readByte())

    @mock.patch('thrift.protocol.TJSONProtocol.TJSONProtocol.readByte')
    @mock.patch('thrift.protocol.TBinaryProtocol.TBinaryProtocol.readMessageBegin')
    @mock.patch('thrift.protocol.TCompactProtocol.TCompactProtocol.readMessageBegin')
    @mock.patch('thrift.protocol.TJSONProtocol.TJSONProtocol.readMessageBegin')
    def test_readMessageBeginParsingJSON(self, mock_json, mock_compact, mock_binary, mock_byte):
        mock_binary.side_effect = TProtocolException()
        mock_compact.side_effect = TProtocolException()
        mock_json.return_value = self.tuple
        mock_byte.return_value = 0xff

        prot = self.factory.getProtocol(self.trans)
        name, ttype, seqid = prot.readMessageBegin()

        self.assertIsInstance(prot.prot, TJSONProtocol)
        self.assertTrue(mock_compact.called)
        self.assertTrue(mock_json.called)
        self.assertFalse(mock_binary.called)
        self.assertEqual(name, self.tuple[0])
        self.assertEqual(ttype, self.tuple[1])
        self.assertEqual(seqid, self.tuple[2])

        self.assertEqual(mock_byte.return_value, prot.readByte())

    @mock.patch('thrift.protocol.TBinaryProtocol.TBinaryProtocol.readMessageBegin')
    @mock.patch('thrift.protocol.TCompactProtocol.TCompactProtocol.readMessageBegin')
    @mock.patch('thrift.protocol.TJSONProtocol.TJSONProtocol.readMessageBegin')
    def test_readMessageBeginParsingUnknown(self, mock_json, mock_compact, mock_binary):
        mock_binary.side_effect = TProtocolException()
        mock_compact.side_effect = TProtocolException()
        mock_json.side_effect = TProtocolException()

        prot = self.factory.getProtocol(self.trans)
        self.assertRaises(TProtocolException, prot.readMessageBegin)
        self.assertTrue(mock_binary.called)
        self.assertTrue(mock_compact.called)
        self.assertTrue(mock_json.called)
