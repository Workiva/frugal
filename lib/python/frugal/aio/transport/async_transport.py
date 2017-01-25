import asyncio
import async_timeout

from thrift.transport.TTransport import TMemoryBuffer
from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FTransportBase
from frugal.context import _OPID_HEADER
from frugal.exceptions import FException
from frugal.context import FContext
from frugal.util.headers import _Headers


class FAsyncTransport(FTransportBase):
    """
     FAsyncTransport is an extension of FTransportBase that asynchronous
     frameworks can implement. Implementations need only implement flush to
     send request data and call handle_response when asynchronous responses
     are received.
    """
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._futures = {}
        self._futures_lock = asyncio.Lock()

    async def oneway(self, context: FContext, payload):
        self._preflight_request_check(payload)
        try:
            with async_timeout.timeout(context.timeout / 1000):
                await self.flush(payload)
        except asyncio.TimeoutError:
            raise TTransportException(
                type=TTransportException.TIMED_OUT,
                message='request timed out'
            ) from None

    async def request(self, context: FContext, payload):
        self._preflight_request_check(payload)
        op_id = str(context._get_op_id())
        future = asyncio.Future()
        async with self._futures_lock:
            if op_id in self._futures:
                raise TTransportException(
                    type=TTransportException.UNKNOWN,
                    message="request already in flight for context"
                )
            self._futures[op_id] = future

        try:
            with async_timeout.timeout(context.timeout / 1000):
                await self.flush(payload)
                return TMemoryBuffer(await future)
        except asyncio.TimeoutError:
            raise TTransportException(
                type=TTransportException.TIMED_OUT,
                message='request timed out'
            ) from None
        finally:
            async with self._futures_lock:
                del self._futures[op_id]

    async def flush(self, payload):
        """Flush the payload to the server."""
        raise NotImplementedError('You must override this')

    async def handle_response(self, frame):
        """
        Complete the future associated with the data frame.

        Args:
            frame: The response frame
        """
        if not frame:
            return
        headers = _Headers.decode_from_frame(frame)
        op_id = headers.get(_OPID_HEADER, None)

        if not op_id:
            raise FException("Frame missing op_id")

        async with self._futures_lock:
            future = self._futures.get(op_id, None)
            if not future:
                return

            future.set_result(frame)
