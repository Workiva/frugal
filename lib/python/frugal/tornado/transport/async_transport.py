from datetime import timedelta

from thrift.protocol.TProtocol import TProtocolException
from thrift.transport.TTransport import TMemoryBuffer
from thrift.transport.TTransport import TTransportException
from tornado import gen
from tornado import locks

from frugal.context import _OPID_HEADER
from frugal.exceptions import FrugalTTransportExceptionType
from frugal.tornado.transport.transport import FTransportBase
from frugal.util.headers import _Headers


class FAsyncTransport(FTransportBase):
    """
     FAsyncTransport is an extension of FTransportBase that asynchronous
     frameworks can implement. Implementations need only implement flush to
     send request data and call handle_response when asynchronous responses
     are received.
    """
    def __init__(self, *args, **kwargs):
        super(FAsyncTransport, self).__init__(*args, **kwargs)
        self._futures = {}
        self._futures_lock = locks.Lock()

    @gen.coroutine
    def oneway(self, context, payload):
        self._preflight_request_check(payload)
        try:
            yield gen.with_timeout(
                timedelta(milliseconds=context.timeout),
                self.flush(payload)
            )
        except gen.TimeoutError:
            raise TTransportException(
                type=FrugalTTransportExceptionType.TIMED_OUT,
                message='oneway timed out'
            )

    @gen.coroutine
    def request(self, context, payload):
        self._preflight_request_check(payload)
        op_id = str(context._get_op_id())
        future = gen.Future()
        with (yield self._futures_lock.acquire()):
            if op_id in self._futures:
                raise TTransportException(
                    type=FrugalTTransportExceptionType.UNKNOWN,
                    message="request already in flight for context"
                )
            self._futures[op_id] = future

        try:
            @gen.coroutine
            def flush_and_wait():
                yield self.flush(payload)
                data = yield future
                raise gen.Return(data)

            data = yield gen.with_timeout(
                timedelta(milliseconds=context.timeout),
                flush_and_wait()
            )
            raise gen.Return(TMemoryBuffer(data))
        except gen.TimeoutError:
            raise TTransportException(
                type=FrugalTTransportExceptionType.TIMED_OUT,
                message='request timed out'
            )
        finally:
            with (yield self._futures_lock.acquire()):
                del self._futures[op_id]

    @gen.coroutine
    def flush(self, payload):
        """Flush the payload to the server."""
        raise NotImplementedError('You must override this')

    @gen.coroutine
    def handle_response(self, frame):
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
            raise TProtocolException(message="Frame missing op_id")

        with (yield self._futures_lock.acquire()):
            future = self._futures.get(op_id, None)
            if not future:
                return

            future.set_result(frame)
