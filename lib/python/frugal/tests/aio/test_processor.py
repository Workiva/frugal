from asyncio import Future
from mock import Mock

from thrift.protocol.TBinaryProtocol import TBinaryProtocolFactory
from thrift.Thrift import TException
from thrift.transport.TTransport import TMemoryBuffer

from frugal.aio.processor import FBaseProcessor
from frugal.context import _OPID_HEADER
from frugal.protocol import FProtocolFactory
from frugal.transport import TMemoryOutputBuffer
from frugal.tests.aio import utils


class TestFBaseProcessor(utils.AsyncIOTestCase):

    @utils.async_runner
    async def test_process_processor_exception(self):
        processor = FBaseProcessor()
        proc = Mock()
        e = TException(message='foo bar exception')
        proc.process.side_effect = e
        processor.add_to_processor_map("basePing", proc)
        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = TMemoryBuffer(value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        oprot = Mock()

        await processor.process(iprot, oprot)

    @utils.async_runner
    async def test_process_missing_function(self):
        processor = FBaseProcessor()
        frame = bytearray(
            b'\x00\x00\x00\x004\x00\x00\x00\x04_cid\x00\x00\x00\x06someid'
            b'\x00\x00\x00\x05_opid\x00\x00\x00\x011\x00\x00\x00\x08_timeout'
            b'\x00\x00\x00\x045000'  # End of context
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = TMemoryBuffer(value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        otrans = TMemoryOutputBuffer(1000)
        oprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(otrans)

        await processor.process(iprot, oprot)

        expected_response = bytearray(
            b'\x80\x01\x00\x03\x00\x00\x00\x08basePing\x00\x00\x00\x00'
            b'\x0b\x00\x01\x00\x00\x00\x1aUnknown function: basePing\x08\x00'
            b'\x02\x00\x00\x00\x01\x00'
        )
        self.assertEqual(otrans.getvalue()[41:], expected_response)

    @utils.async_runner
    async def test_process(self):
        processor = FBaseProcessor()
        proc = Mock()
        future = Future()
        future.set_result(None)
        proc.process.return_value = future
        processor.add_to_processor_map("basePing", proc)
        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = TMemoryBuffer(value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        oprot = Mock()
        await processor.process(iprot, oprot)
        assert(proc.process.call_args)
        args, _ = proc.process.call_args
        self.assertEqual(args[0].get_response_header(_OPID_HEADER), '1')
        assert(args[1] == iprot)
        assert(args[2] == oprot)
