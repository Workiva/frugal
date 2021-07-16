# Copyright 2017 Workiva
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import mock

from asyncio import Future
from asyncio import gather
from asyncio import sleep

from thrift.protocol.TProtocol import TProtocolException
from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FAsyncTransport
from frugal.context import FContext
from frugal.exceptions import TTransportExceptionType
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
        self.assertEqual(TTransportExceptionType.NOT_OPEN, cm.exception.type)
        self.assertIsNone(transport._payload)

    @aio_utils.async_runner
    async def test_oneway_size_exception(self):
        ctx = FContext("fooid")
        ctx.timeout = 10
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(request_size_limit=1)
        with self.assertRaises(TTransportException) as cm:
            await transport.oneway(ctx, frame)
        self.assertEqual(TTransportExceptionType.REQUEST_TOO_LARGE,
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
        self.assertEqual(TTransportExceptionType.TIMED_OUT, cm.exception.type)
        self.assertEqual(frame, transport._payload)

    @aio_utils.async_runner
    async def test_request(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        message = utils.mock_message_with_frame(frame)
        transport = FAsyncTransportImpl(response=message)
        response_transport = await transport.request(ctx, frame)
        self.assertEqual(frame[4:], response_transport.getvalue())
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
        self.assertEqual(TTransportExceptionType.NOT_OPEN, cm.exception.type)
        self.assertEqual(0, len(transport._futures))
        self.assertIsNone(transport._payload)

    @aio_utils.async_runner
    async def test_request_size_exception(self):
        ctx = FContext("fooid")
        ctx.timeout = 10
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl(request_size_limit=1)
        with self.assertRaises(TTransportException) as cm:
            await transport.request(ctx, frame)
        self.assertEqual(TTransportExceptionType.REQUEST_TOO_LARGE,
                         cm.exception.type)
        self.assertEqual(0, len(transport._futures))
        self.assertIsNone(transport._payload)

    @aio_utils.async_runner
    async def test_request_flush_exception(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        e = TTransportException(
            type=TTransportExceptionType.END_OF_FILE,
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
        self.assertEqual(TTransportExceptionType.TIMED_OUT, cm.exception.type)
        self.assertEqual(0, len(transport._futures))
        self.assertEqual(frame, transport._payload)

    @aio_utils.async_runner
    async def test_request_response_timeout(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        transport = FAsyncTransportImpl()
        with self.assertRaises(TTransportException) as cm:
            await transport.request(ctx, frame)
        self.assertEqual(TTransportExceptionType.TIMED_OUT, cm.exception.type)
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
        self.assertEqual(TTransportExceptionType.UNKNOWN, cm.exception.type)
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
        await transport.handle_response(utils.mock_message_with_frame(None))
        self.assertTrue(future.done())

    @aio_utils.async_runner
    async def test_handle_response_bad_frame(self):
        transport = FAsyncTransport(1024)

        with self.assertRaises(TProtocolException) as cm:
            message = utils.mock_message_with_frame(b"foobars")
            await transport.handle_response(message)

        self.assertEqual("Invalid frame size: 3", str(cm.exception))

    @aio_utils.async_runner
    async def test_handle_response_missing_op_id(self):
        transport = FAsyncTransport(1024)
        frame = bytearray(b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x80\x01\x00'
                          b'\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00')

        with self.assertRaises(TProtocolException) as cm:
            message = utils.mock_message_with_frame(frame)
            await transport.handle_response(message)

        self.assertEquals("Frame missing op_id", str(cm.exception))

    @aio_utils.async_runner
    async def test_handle_response_unregistered_op_id(self):
        transport = FAsyncTransport(1024)
        ctx1 = FContext()
        ctx2 = FContext()
        future = Future()
        transport._futures[str(ctx1._get_op_id())] = future

        await transport.handle_response(utils.mock_message_with_context(ctx2))
        self.assertFalse(future.done())

    @aio_utils.async_runner
    async def test_service_not_available(self):
        ctx = FContext("fooid")
        frame = utils.mock_frame(ctx)
        message = utils.mock_message_with_frame(bytearray())
        transport = FAsyncTransportImpl(response=message)
        with self.assertRaises(TTransportException) as te:
            await transport.request(ctx, frame)
            self.assertEquals("request: service not available", str(te.exception))
            self.assertTrue(False) # WHY DOES THIS ASSERTION PASS
