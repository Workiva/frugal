import unittest
from mock import Mock

from thrift.protocol.TBinaryProtocol import TBinaryProtocolFactory
from thrift.Thrift import TApplicationException

from frugal.exceptions import FException
from frugal.processor import FBaseProcessor
from frugal.protocol import FProtocolFactory
from frugal.transport import FBoundedMemoryBuffer


class TestFBaseProcessor(unittest.TestCase):

    def test_process_processor_exception(self):
        processor = FBaseProcessor()
        proc = Mock()
        proc.process.side_effect = FException()
        processor.add_to_processor_map("basePing", proc)
        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = FBoundedMemoryBuffer(len(frame), value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        oprot = Mock()
        with self.assertRaises(FException):
            processor.process(iprot, oprot)

    def test_process_missing_function(self):
        processor = FBaseProcessor()
        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = FBoundedMemoryBuffer(len(frame), value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        otrans = FBoundedMemoryBuffer(100)
        oprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(otrans)
        with self.assertRaises(TApplicationException):
            processor.process(iprot, oprot)

        expected_response = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011\x80'
            b'\x01\x00\x03\x00\x00\x00\x08basePing\x00\x00\x00\x00\x0b\x00\x01'
            b'\x00\x00\x00\x1aUnknown function: basePing\x08\x00\x02\x00\x00'
            b'\x00\x01\x00'
        )
        assert(otrans.getvalue() == expected_response)

    def test_process(self):
        processor = FBaseProcessor()
        proc = Mock()
        processor.add_to_processor_map("basePing", proc)
        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = FBoundedMemoryBuffer(len(frame), value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        oprot = Mock()
        processor.process(iprot, oprot)
        assert(proc.process.call_args)
        args, _ = proc.process.call_args
        assert(args[0]._get_op_id() == 1)
        assert(args[1] == iprot)
        assert(args[2] == oprot)
