#
# Autogenerated by Frugal Compiler (1.15.0)
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

from base.BaseFoo import *
from base.ttypes import *


class Iface(object):

    def basePing(self, ctx):
        """
        Args:
            ctx: FContext
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
            'basePing': Method(self._basePing, middleware),
        }

    def basePing(self, ctx):
        """
        Args:
            ctx: FContext
        """
        return self._methods['basePing']([ctx])

    @gen.coroutine
    def _basePing(self, ctx):
        delta = timedelta(milliseconds=ctx.get_timeout())
        future = gen.with_timeout(delta, Future())
        self._transport.register(ctx, self._recv_basePing(ctx, future))
        yield self._send_basePing(ctx)

        try:
            result = yield future
        finally:
            self._transport.unregister(ctx)
        raise gen.Return(result)

    @gen.coroutine
    def _send_basePing(self, ctx):
        oprot = self._oprot
        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin('basePing', TMessageType.CALL, 0)
            args = basePing_args()
            args.write(oprot)
            oprot.writeMessageEnd()
            yield oprot.get_transport().flush()

    def _recv_basePing(self, ctx, future):
        def basePing_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            iprot.read_response_headers(ctx)
            _, mtype, _ = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                future.set_exception(x)
                raise x
            result = basePing_result()
            result.read(iprot)
            iprot.readMessageEnd()
            future.set_result(None)
        return basePing_callback


class Processor(FBaseProcessor):

    def __init__(self, handler):
        """
        Create a new Processor.

        Args:
            handler: Iface
        """
        super(Processor, self).__init__()
        self.add_to_processor_map('basePing', _basePing(handler, self.get_write_lock()))


class _basePing(FProcessorFunction):

    def __init__(self, handler, lock):
        self._handler = handler
        self._lock = lock

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        args = basePing_args()
        args.read(iprot)
        iprot.readMessageEnd()
        result = basePing_result()
        yield gen.maybe_future(self._handler.basePing(ctx))
        with self._lock:
            oprot.write_response_headers(ctx)
            oprot.writeMessageBegin('basePing', TMessageType.REPLY, 0)
            result.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()


