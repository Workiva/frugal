#
# Autogenerated by Frugal Compiler (1.4.0)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#



from threading import Lock

from frugal.middleware import Method
from frugal.processor import FBaseProcessor
from frugal.processor import FProcessorFunction
from frugal.registry import FClientRegistry
from thrift.Thrift import TApplicationException
from thrift.Thrift import TMessageType
from tornado import gen
from tornado.concurrent import Future

import validStructs
import ValidTypes
from valid.Blah import *
from valid.ttypes import *


class Iface(object):

    def ping(self, ctx):
        """
        Use this to ping the server.
        
        Args:
            ctx: FContext
        """
        pass

    def bleh(self, ctx, one, Two, custom_ints):
        """
        Use this to tell the sever how you feel.
        
        Args:
            ctx: FContext
            one: Thing
            Two: Stuff
            custom_ints: list of int (signed 32 bits)
        """
        pass

    def getThing(self, ctx):
        """
        Args:
            ctx: FContext
        """
        pass

    def getMyInt(self, ctx):
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
            'ping': Method(self._ping, middleware),
            'bleh': Method(self._bleh, middleware),
            'getThing': Method(self._getThing, middleware),
            'getMyInt': Method(self._getMyInt, middleware),
        }

    def ping(self, ctx):
        """
        Use this to ping the server.
        
        Args:
            ctx: FContext
        """
        return self._methods['ping']([ctx])

    def _ping(self, ctx):
        future = Future()
        self._send_ping(ctx, future)
        return future

    def _send_ping(self, ctx, future):
        oprot = self._oprot
        self._transport.register(ctx, self._recv_ping(ctx, future))
        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin('ping', TMessageType.CALL, 0)
            args = ping_args()
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def _recv_ping(self, ctx, future):
        def ping_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            iprot.read_response_headers(ctx)
            _, mtype, _ = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                future.set_exception(x)
                raise x
            result = ping_result()
            result.read(iprot)
            iprot.readMessageEnd()
            future.set_result(None)
        return ping_callback

    def bleh(self, ctx, one, Two, custom_ints):
        """
        Use this to tell the sever how you feel.
        
        Args:
            ctx: FContext
            one: Thing
            Two: Stuff
            custom_ints: list of int (signed 32 bits)
        """
        return self._methods['bleh']([ctx, one, Two, custom_ints])

    def _bleh(self, ctx, one, Two, custom_ints):
        future = Future()
        self._send_bleh(ctx, future, one, Two, custom_ints)
        return future

    def _send_bleh(self, ctx, future, one, Two, custom_ints):
        oprot = self._oprot
        self._transport.register(ctx, self._recv_bleh(ctx, future))
        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin('bleh', TMessageType.CALL, 0)
            args = bleh_args()
            args.one = one
            args.Two = Two
            args.custom_ints = custom_ints
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def _recv_bleh(self, ctx, future):
        def bleh_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            iprot.read_response_headers(ctx)
            _, mtype, _ = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                future.set_exception(x)
                raise x
            result = bleh_result()
            result.read(iprot)
            iprot.readMessageEnd()
            if result.oops is not None:
                future.set_exception(result.oops)
                return
            if result.success is not None:
                future.set_result(result.success)
                return
            x = TApplicationException(TApplicationException.MISSING_RESULT, "bleh failed: unknown result")
            future.set_exception(x)
            raise x
        return bleh_callback

    def getThing(self, ctx):
        """
        Args:
            ctx: FContext
        """
        return self._methods['getThing']([ctx])

    def _getThing(self, ctx):
        future = Future()
        self._send_getThing(ctx, future)
        return future

    def _send_getThing(self, ctx, future):
        oprot = self._oprot
        self._transport.register(ctx, self._recv_getThing(ctx, future))
        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin('getThing', TMessageType.CALL, 0)
            args = getThing_args()
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def _recv_getThing(self, ctx, future):
        def getThing_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            iprot.read_response_headers(ctx)
            _, mtype, _ = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                future.set_exception(x)
                raise x
            result = getThing_result()
            result.read(iprot)
            iprot.readMessageEnd()
            if result.success is not None:
                future.set_result(result.success)
                return
            x = TApplicationException(TApplicationException.MISSING_RESULT, "getThing failed: unknown result")
            future.set_exception(x)
            raise x
        return getThing_callback

    def getMyInt(self, ctx):
        """
        Args:
            ctx: FContext
        """
        return self._methods['getMyInt']([ctx])

    def _getMyInt(self, ctx):
        future = Future()
        self._send_getMyInt(ctx, future)
        return future

    def _send_getMyInt(self, ctx, future):
        oprot = self._oprot
        self._transport.register(ctx, self._recv_getMyInt(ctx, future))
        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin('getMyInt', TMessageType.CALL, 0)
            args = getMyInt_args()
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def _recv_getMyInt(self, ctx, future):
        def getMyInt_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            iprot.read_response_headers(ctx)
            _, mtype, _ = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                future.set_exception(x)
                raise x
            result = getMyInt_result()
            result.read(iprot)
            iprot.readMessageEnd()
            if result.success is not None:
                future.set_result(result.success)
                return
            x = TApplicationException(TApplicationException.MISSING_RESULT, "getMyInt failed: unknown result")
            future.set_exception(x)
            raise x
        return getMyInt_callback


class Processor(FBaseProcessor):

    def __init__(self, handler):
        """
        Create a new Processor.

        Args:
            handler: Iface
        """
        super(Processor, self).__init__()
        self.add_to_processor_map('ping', _ping(handler, self.get_write_lock()))
        self.add_to_processor_map('bleh', _bleh(handler, self.get_write_lock()))
        self.add_to_processor_map('getThing', _getThing(handler, self.get_write_lock()))
        self.add_to_processor_map('getMyInt', _getMyInt(handler, self.get_write_lock()))


class _ping(FProcessorFunction):

    def __init__(self, handler, lock):
        self._handler = handler
        self._lock = lock

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        args = ping_args()
        args.read(iprot)
        iprot.readMessageEnd()
        result = ping_result()
        yield gen.maybe_future(self._handler.ping(ctx))
        with self._lock:
            oprot.write_response_headers(ctx)
            oprot.writeMessageBegin('ping', TMessageType.REPLY, 0)
            result.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()


class _bleh(FProcessorFunction):

    def __init__(self, handler, lock):
        self._handler = handler
        self._lock = lock

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        args = bleh_args()
        args.read(iprot)
        iprot.readMessageEnd()
        result = bleh_result()
        try:
            result.success = yield gen.maybe_future(self._handler.bleh(ctx, args.one, args.Two, args.custom_ints))
        except InvalidOperation as oops:
            result.oops = oops
        with self._lock:
            oprot.write_response_headers(ctx)
            oprot.writeMessageBegin('bleh', TMessageType.REPLY, 0)
            result.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()


class _getThing(FProcessorFunction):

    def __init__(self, handler, lock):
        self._handler = handler
        self._lock = lock

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        args = getThing_args()
        args.read(iprot)
        iprot.readMessageEnd()
        result = getThing_result()
        result.success = yield gen.maybe_future(self._handler.getThing(ctx))
        with self._lock:
            oprot.write_response_headers(ctx)
            oprot.writeMessageBegin('getThing', TMessageType.REPLY, 0)
            result.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()


class _getMyInt(FProcessorFunction):

    def __init__(self, handler, lock):
        self._handler = handler
        self._lock = lock

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        args = getMyInt_args()
        args.read(iprot)
        iprot.readMessageEnd()
        result = getMyInt_result()
        result.success = yield gen.maybe_future(self._handler.getMyInt(ctx))
        with self._lock:
            oprot.write_response_headers(ctx)
            oprot.writeMessageBegin('getMyInt', TMessageType.REPLY, 0)
            result.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()


