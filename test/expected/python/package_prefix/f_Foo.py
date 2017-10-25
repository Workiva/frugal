#
# Autogenerated by Frugal Compiler (2.11.0)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#



from threading import Lock

from frugal.middleware import Method
from frugal.exceptions import TApplicationExceptionType
from frugal.exceptions import TTransportExceptionType
from frugal.processor import FBaseProcessor
from frugal.processor import FProcessorFunction
from frugal.util.deprecate import deprecated
from thrift.Thrift import TApplicationException
from thrift.Thrift import TMessageType
from thrift.transport.TTransport import TTransportException

import generic_package_prefix.actual_base.python.f_BaseFoo
import generic_package_prefix.actual_base.python.ttypes
import generic_package_prefix.actual_base.python.constants
from .ttypes import *


class Iface(generic_package_prefix.actual_base.python.f_BaseFoo.Iface):

    def get_thing(self, ctx, the_thing):
        """
        Args:
            ctx: FContext
            the_thing: base.thing
        """
        pass


class Client(generic_package_prefix.actual_base.python.f_BaseFoo.Client, Iface):

    def __init__(self, provider, middleware=None):
        """
        Create a new Client with an FServiceProvider containing a transport
        and protocol factory.

        Args:
            provider: FServiceProvider with TSynchronousTransport
            middleware: ServiceMiddleware or list of ServiceMiddleware
        """
        middleware = middleware or []
        if middleware and not isinstance(middleware, list):
            middleware = [middleware]
        super(Client, self).__init__(provider, middleware=middleware)
        middleware += provider.get_middleware()
        self._methods.update({
            'get_thing': Method(self._get_thing, middleware),
        })

    def get_thing(self, ctx, the_thing):
        """
        Args:
            ctx: FContext
            the_thing: base.thing
        """
        return self._methods['get_thing']([ctx, the_thing])

    def _get_thing(self, ctx, the_thing):
        self._send_get_thing(ctx, the_thing)
        return self._recv_get_thing(ctx)

    def _send_get_thing(self, ctx, the_thing):
        oprot = self._oprot
        with self._write_lock:
            oprot.get_transport().set_timeout(ctx.timeout)
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin('get_thing', TMessageType.CALL, 0)
            args = get_thing_args()
            args.the_thing = the_thing
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def _recv_get_thing(self, ctx):
        self._iprot.read_response_headers(ctx)
        _, mtype, _ = self._iprot.readMessageBegin()
        if mtype == TMessageType.EXCEPTION:
            x = TApplicationException()
            x.read(self._iprot)
            self._iprot.readMessageEnd()
            if x.type == TApplicationExceptionType.RESPONSE_TOO_LARGE:
                raise TTransportException(type=TTransportExceptionType.RESPONSE_TOO_LARGE, message=x.message)
            raise x
        result = get_thing_result()
        result.read(self._iprot)
        self._iprot.readMessageEnd()
        if result.success is not None:
            return result.success
        x = TApplicationException(TApplicationExceptionType.MISSING_RESULT, "get_thing failed: unknown result")
        raise x

class Processor(generic_package_prefix.actual_base.python.f_BaseFoo.Processor):

    def __init__(self, handler, middleware=None):
        """
        Create a new Processor.

        Args:
            handler: Iface
        """
        if middleware and not isinstance(middleware, list):
            middleware = [middleware]

        super(Processor, self).__init__(handler, middleware=middleware)
        self.add_to_processor_map('get_thing', _get_thing(Method(handler.get_thing, middleware), self.get_write_lock()))


class _get_thing(FProcessorFunction):

    def __init__(self, handler, lock):
        super(_get_thing, self).__init__(handler, lock)

    def process(self, ctx, iprot, oprot):
        args = get_thing_args()
        args.read(iprot)
        iprot.readMessageEnd()
        result = get_thing_result()
        try:
            result.success = self._handler([ctx, args.the_thing])
        except TApplicationException as ex:
            with self._lock:
                _write_application_exception(ctx, oprot, "get_thing", exception=ex)
                return
        except Exception as e:
            with self._lock:
                _write_application_exception(ctx, oprot, "get_thing", ex_code=TApplicationExceptionType.INTERNAL_ERROR, message=e.message)
            raise
        with self._lock:
            try:
                oprot.write_response_headers(ctx)
                oprot.writeMessageBegin('get_thing', TMessageType.REPLY, 0)
                result.write(oprot)
                oprot.writeMessageEnd()
                oprot.get_transport().flush()
            except TTransportException as e:
                # catch a request too large error because the TMemoryOutputBuffer always throws that if too much data is written
                if e.type == TTransportExceptionType.REQUEST_TOO_LARGE:
                    raise _write_application_exception(ctx, oprot, "get_thing", ex_code=TApplicationExceptionType.RESPONSE_TOO_LARGE, message=e.args[0])
                else:
                    raise e


def _write_application_exception(ctx, oprot, method, ex_code=None, message=None, exception=None):
    if exception is not None:
        x = exception
    else:
        x = TApplicationException(type=ex_code, message=message)
    oprot.write_response_headers(ctx)
    oprot.writeMessageBegin(method, TMessageType.EXCEPTION, 0)
    x.write(oprot)
    oprot.writeMessageEnd()
    oprot.get_transport().flush()
    return x

class get_thing_args(object):
    """
    Attributes:
     - the_thing
    """
    def __init__(self, the_thing=None):
        self.the_thing = the_thing

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 1:
                if ftype == TType.STRUCT:
                    self.the_thing = generic_package_prefix.actual_base.python.ttypes.thing()
                    self.the_thing.read(iprot)
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()
        self.validate()

    def write(self, oprot):
        self.validate()
        oprot.writeStructBegin('get_thing_args')
        if self.the_thing is not None:
            oprot.writeFieldBegin('the_thing', TType.STRUCT, 1)
            self.the_thing.write(oprot)
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value * 31) ^ hash(self.the_thing)
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
            for key, value in self.__dict__.items()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

    def __ne__(self, other):
        return not (self == other)

class get_thing_result(object):
    """
    Attributes:
     - success
    """
    def __init__(self, success=None):
        self.success = success

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 0:
                if ftype == TType.STRUCT:
                    self.success = generic_package_prefix.actual_base.python.ttypes.thing()
                    self.success.read(iprot)
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()
        self.validate()

    def write(self, oprot):
        self.validate()
        oprot.writeStructBegin('get_thing_result')
        if self.success is not None:
            oprot.writeFieldBegin('success', TType.STRUCT, 0)
            self.success.write(oprot)
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value * 31) ^ hash(self.success)
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
            for key, value in self.__dict__.items()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

    def __ne__(self, other):
        return not (self == other)

