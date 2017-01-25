from asyncio import Future
from asyncio import gather
from asyncio import sleep

from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FAsyncTransport
from frugal.context import FContext
from frugal.exceptions import FException
from frugal.exceptions import FMessageSizeException
from frugal.exceptions import FProtocolException
from frugal.exceptions import FTransportException
from frugal.tests import utils
from frugal.tests.aio import utils as aio_utils


class FAsyncTransportImpl(FAsyncTransport):
    def __init__(self, response=None, e=None, flush_wait=0, is_open=True,
                 *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._payload = None
        self._response = response
        self._exception = e
        self._flush_wait = flush_wait
        self._is_open = is_open

    def is_open(self):
        return self._is_open

    async def flush(self, payload):
        self._payload = payload
        if self._flush_wait > 0:
            await sleep(self._flush_wait)
        if self._response:
            await self.handle_response(self._response)
        if self._exception:
            raise self._exception


class TestFAsyncTransport(aio_utils.AsyncIOTestCase):

    @aio_utils.async_runner
    async def test_oneway(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl()
        self.assertIsNone(await transport.oneway(ctx, frame))
        self.assertEqual(frame, transport._payload)

    @aio_utils.async_runner
    async def test_oneway_not_open(self):
        ctx = FContext("fooid")
        ctx.timeout = 10
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(is_open=False)
        with self.assertRaises(TTransportException) as cm:
            await transport.oneway(ctx, frame)
        self.assertEqual(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertIsNone(transport._payload)

    @aio_utils.async_runner
    async def test_oneway_size_exception(self):
        ctx = FContext("fooid")
        ctx.timeout = 10
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(request_size_limit=1)
        with self.assertRaises(FMessageSizeException) as cm:
            await transport.oneway(ctx, frame)
        self.assertEqual(FTransportException.REQUEST_TOO_LARGE,
                         cm.exception.type)
        self.assertIsNone(transport._payload)

    @aio_utils.async_runner
    async def test_oneway_timeout(self):
        ctx = FContext("fooid")
        ctx.timeout = 10
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(flush_wait=1)
        with self.assertRaises(TTransportException) as cm:
            await transport.oneway(ctx, frame)
        self.assertEqual(TTransportException.TIMED_OUT, cm.exception.type)
        self.assertEqual(frame, transport._payload)

    @aio_utils.async_runner
    async def test_request(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(response=frame)
        response_transport = await transport.request(ctx, frame)
        self.assertEqual(frame, response_transport.getvalue())
        self.assertEqual(0, len(transport._futures))
        self.assertEqual(frame, transport._payload)

    @aio_utils.async_runner
    async def test_request_not_open(self):
        ctx = FContext("fooid")
        ctx.timeout = 10
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(is_open=False)
        with self.assertRaises(TTransportException) as cm:
            await transport.request(ctx, frame)
        self.assertEqual(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEqual(0, len(transport._futures))
        self.assertIsNone(transport._payload)

    @aio_utils.async_runner
    async def test_request_size_exception(self):
        ctx = FContext("fooid")
        ctx.timeout = 10
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(request_size_limit=1)
        with self.assertRaises(FMessageSizeException) as cm:
            await transport.request(ctx, frame)
        self.assertEqual(FTransportException.REQUEST_TOO_LARGE,
                         cm.exception.type)
        self.assertEqual(0, len(transport._futures))
        self.assertIsNone(transport._payload)

    @aio_utils.async_runner
    async def test_request_flush_exception(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        e = TTransportException(
            type=TTransportException.END_OF_FILE,
            message="oh no!"
        )
        transport = FAsyncTransportImpl(e=e)
        with self.assertRaises(TTransportException) as cm:
            await transport.request(ctx, frame)
        self.assertEqual(e, cm.exception)
        self.assertEqual(0, len(transport._futures))
        self.assertEqual(frame, transport._payload)

    @aio_utils.async_runner
    async def test_request_flush_timeout(self):
        ctx = FContext("fooid")
        ctx.timeout = 10
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(flush_wait=1)
        with self.assertRaises(TTransportException) as cm:
            await transport.request(ctx, frame)
        self.assertEqual(TTransportException.TIMED_OUT, cm.exception.type)
        self.assertEqual(0, len(transport._futures))
        self.assertEqual(frame, transport._payload)

    @aio_utils.async_runner
    async def test_request_response_timeout(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl()
        with self.assertRaises(TTransportException) as cm:
            await transport.request(ctx, frame)
        self.assertEqual(TTransportException.TIMED_OUT, cm.exception.type)
        self.assertEqual(0, len(transport._futures))
        self.assertEqual(frame, transport._payload)

    @aio_utils.async_runner
    async def test_request_pending(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl()
        with self.assertRaises(TTransportException) as cm:
            await gather(
                transport.request(ctx, frame),
                transport.request(ctx, frame)
            )
        self.assertEqual(TTransportException.UNKNOWN, cm.exception.type)
        self.assertEqual("request already in flight for context",
                         cm.exception.message)
        # We still have one request pending
        self.assertEqual(1, len(transport._futures))

    @aio_utils.async_runner
    async def test_handle_response_none(self):
        transport = FAsyncTransport(1024)
        ctx = FContext()
        future = Future()
        transport._futures[str(ctx._get_op_id())] = future
        await transport.handle_response(None)
        self.assertFalse(future.done())

    @aio_utils.async_runner
    async def test_handle_response_bad_frame(self):
        transport = FAsyncTransport(1024)

        with self.assertRaises(FProtocolException) as cm:
            await transport.handle_response(b"foo")

        self.assertEquals("Invalid frame size: 3", str(cm.exception))

    @aio_utils.async_runner
    async def test_handle_response_missing_op_id(self):
        transport = FAsyncTransport(1024)
        frame = bytearray(b'\x00\x00\x00\x00\x00\x80\x01\x00\x02\x00\x00\x00'
                          b'\x08basePing\x00\x00\x00\x00\x00')

        with self.assertRaises(FException) as cm:
            await transport.handle_response(frame)

        self.assertEquals("Frame missing op_id", str(cm.exception))

    @aio_utils.async_runner
    async def test_handle_response_unregistered_op_id(self):
        transport = FAsyncTransport(1024)
        ctx1 = FContext()
        ctx2 = FContext()
        future = Future()
        transport._futures[str(ctx1._get_op_id())] = future
        await transport.handle_response(utils.mock_frame(ctx2))
        self.assertFalse(future.done())

    @aio_utils.async_runner
    async def test_flush_not_implemented(self):
        transport = FAsyncTransport(1024)
        with self.assertRaises(NotImplementedError):
            await transport.flush(None)
