#
# Autogenerated by Frugal Compiler (1.14.0)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#



from datetime import timedelta
from threading import Lock

from frugal.middleware import Method
from frugal.processor import FBaseProcessor
from frugal.processor import FProcessorFunction
from frugal.registry import FClientRegistry
from thrift.Thrift import TApplicationException
from thrift.Thrift import TMessageType
from tornado import gen
from tornado.concurrent import Future

from music.Store import *
from music.ttypes import *


class Iface(object):

    def buyAlbum(self, ctx, ASIN, acct):
        """
        Args:
            ctx: FContext
            ASIN: string
            acct: string
        """
        pass

    def enterAlbumGiveaway(self, ctx, email, name):
        """
        Args:
            ctx: FContext
            email: string
            name: string
        """
        pass


class Client(Iface):

    def __init__(self, transport, protocol_factory, middleware=None):
        """
        Create a new Client with a transport and protocol factory.

        Args:
            transport: FTransport
            protocol_factory: FProtocolFactory
            middleware: ServiceMiddleware or list of ServiceMiddleware
        """
        if middleware and not isinstance(middleware, list):
            middleware = [middleware]
        transport.set_registry(FClientRegistry())
        self._transport = transport
        self._protocol_factory = protocol_factory
        self._oprot = protocol_factory.get_protocol(transport)
        self._write_lock = Lock()
        self._methods = {
            'buyAlbum': Method(self._buyAlbum, middleware),
            'enterAlbumGiveaway': Method(self._enterAlbumGiveaway, middleware),
        }

    def buyAlbum(self, ctx, ASIN, acct):
        """
        Args:
            ctx: FContext
            ASIN: string
            acct: string
        """
        return self._methods['buyAlbum']([ctx, ASIN, acct])

    @gen.coroutine
    def _buyAlbum(self, ctx, ASIN, acct):
        delta = timedelta(milliseconds=ctx.get_timeout())
        future = gen.with_timeout(delta, Future())
        self._transport.register(ctx, self._recv_buyAlbum(ctx, future))
        yield self._send_buyAlbum(ctx, ASIN, acct)

        try:
            result = yield future
        finally:
            self._transport.unregister(ctx)
        raise gen.Return(result)

    @gen.coroutine
    def _send_buyAlbum(self, ctx, ASIN, acct):
        oprot = self._oprot
        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin('buyAlbum', TMessageType.CALL, 0)
            args = buyAlbum_args()
            args.ASIN = ASIN
            args.acct = acct
            args.write(oprot)
            oprot.writeMessageEnd()
            yield oprot.get_transport().flush()

    def _recv_buyAlbum(self, ctx, future):
        def buyAlbum_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            iprot.read_response_headers(ctx)
            _, mtype, _ = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                future.set_exception(x)
                raise x
            result = buyAlbum_result()
            result.read(iprot)
            iprot.readMessageEnd()
            if result.error is not None:
                future.set_exception(result.error)
                return
            if result.success is not None:
                future.set_result(result.success)
                return
            x = TApplicationException(TApplicationException.MISSING_RESULT, "buyAlbum failed: unknown result")
            future.set_exception(x)
            raise x
        return buyAlbum_callback

    def enterAlbumGiveaway(self, ctx, email, name):
        """
        Args:
            ctx: FContext
            email: string
            name: string
        """
        return self._methods['enterAlbumGiveaway']([ctx, email, name])

    @gen.coroutine
    def _enterAlbumGiveaway(self, ctx, email, name):
        delta = timedelta(milliseconds=ctx.get_timeout())
        future = gen.with_timeout(delta, Future())
        self._transport.register(ctx, self._recv_enterAlbumGiveaway(ctx, future))
        yield self._send_enterAlbumGiveaway(ctx, email, name)

        try:
            result = yield future
        finally:
            self._transport.unregister(ctx)
        raise gen.Return(result)

    @gen.coroutine
    def _send_enterAlbumGiveaway(self, ctx, email, name):
        oprot = self._oprot
        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin('enterAlbumGiveaway', TMessageType.CALL, 0)
            args = enterAlbumGiveaway_args()
            args.email = email
            args.name = name
            args.write(oprot)
            oprot.writeMessageEnd()
            yield oprot.get_transport().flush()

    def _recv_enterAlbumGiveaway(self, ctx, future):
        def enterAlbumGiveaway_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            iprot.read_response_headers(ctx)
            _, mtype, _ = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                future.set_exception(x)
                raise x
            result = enterAlbumGiveaway_result()
            result.read(iprot)
            iprot.readMessageEnd()
            if result.success is not None:
                future.set_result(result.success)
                return
            x = TApplicationException(TApplicationException.MISSING_RESULT, "enterAlbumGiveaway failed: unknown result")
            future.set_exception(x)
            raise x
        return enterAlbumGiveaway_callback


class Processor(FBaseProcessor):

    def __init__(self, handler):
        """
        Create a new Processor.

        Args:
            handler: Iface
        """
        super(Processor, self).__init__()
        self.add_to_processor_map('buyAlbum', _buyAlbum(handler, self.get_write_lock()))
        self.add_to_processor_map('enterAlbumGiveaway', _enterAlbumGiveaway(handler, self.get_write_lock()))


class _buyAlbum(FProcessorFunction):

    def __init__(self, handler, lock):
        self._handler = handler
        self._lock = lock

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        args = buyAlbum_args()
        args.read(iprot)
        iprot.readMessageEnd()
        result = buyAlbum_result()
        try:
            result.success = yield gen.maybe_future(self._handler.buyAlbum(ctx, args.ASIN, args.acct))
        except PurchasingError as error:
            result.error = error
        with self._lock:
            oprot.write_response_headers(ctx)
            oprot.writeMessageBegin('buyAlbum', TMessageType.REPLY, 0)
            result.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()


class _enterAlbumGiveaway(FProcessorFunction):

    def __init__(self, handler, lock):
        self._handler = handler
        self._lock = lock

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        args = enterAlbumGiveaway_args()
        args.read(iprot)
        iprot.readMessageEnd()
        result = enterAlbumGiveaway_result()
        result.success = yield gen.maybe_future(self._handler.enterAlbumGiveaway(ctx, args.email, args.name))
        with self._lock:
            oprot.write_response_headers(ctx)
            oprot.writeMessageBegin('enterAlbumGiveaway', TMessageType.REPLY, 0)
            result.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()


