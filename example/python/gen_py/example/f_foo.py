from threading import Lock

from tornado import concurrent, gen

from gen_py.base import f_base_foo

from frugal.registry import FClientRegistry
from thrift.Thrift import TType, TMessageType, TApplicationException

from .ttypes import Event, AwesomeException
from gen_py.base.ttypes import api_exception


class Iface(f_base_foo.Iface):

    def blah(self, context, num, Str, event):
        pass

    def ping(self, context):
        pass

    def one_way(self, context, id, req):
        pass


class Client(f_base_foo.Client, Iface):

    def __init__(self, transport, protocol_factory):
        """Initialize a Client with a transport and protocol factory creating a
        new FClientRegistry

            Args:
                transport: FTransport
                protocol_factory: FProtocolFactory
        """
        f_base_foo.Client.__init__(self, transport, protocol_factory)
        self._transport = transport
        self._transport.set_registry(FClientRegistry())
        self._protocol_factory = protocol_factory
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def one_way(self, context, id, req):
        """ oneway methods don't receive a response from the server

        Args:
            context: FContext
            req: dict key values to send (will be converted to JSON string)
        """
        oprot = self._oprot
        with self._write_lock:
            oprot.write_request_headers(context)
            oprot.writeMessageBegin("oneWay", TMessageType.ONEWAY, 0)
            args = one_way_args()
            args.id = id
            args.req = req
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def ping(self, context):
        future = concurrent.Future()
        self.send_ping(context, future)
        return future

    def blah(self, context, num, Str, event):
        future = concurrent.Future()
        print("calling blah")
        self.send_blah(context, future, num, Str, event)
        return future

    def send_ping(self, context, future):
        oprot = self._oprot
        self._transport.register(context, self.recv_ping(context, future))
        with self._write_lock:
            oprot.write_request_headers(context)
            oprot.writeMessageBegin('ping', TMessageType.CALL, 0)
            args = ping_args()
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def recv_ping(self, context, future):
        def ping_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            ctx = iprot.read_response_headers(context)
            (fname, mtype, fid) = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                raise x
            print("Received a ping response")
            result = ping_result()
            result.read(iprot)
            iprot.readMessageEnd()
            raise gen.Return(ctx)
        return ping_callback

    def send_blah(self, context, future, num, Str, event):
        oprot = self._oprot
        self._transport.register(context, self.recv_blah(context, future))
        with self._write_lock:
            oprot.write_request_headers(context)
            oprot.writeMessageBegin('blah', TMessageType.CALL, 0)
            args = blah_args()
            args.num = num
            args.Str = Str
            args.event = event
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def recv_blah(self, context, future):
        def blah_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            ctx = iprot.read_response_headers(context)
            (fname, mtype, fid) = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                raise x
            result = blah_result()
            result.read(iprot)
            iprot.readMessageEnd()
            if result.success is not None:
                return result.success
            if result.awe is not None:
                raise result.awe
            if result.api is not None:
                raise result.api
            raise TApplicationException(TApplicationException.MISSING_RESULT,
                                        "blah failed: unknown result")
        return blah_callback


class ping_args(object):

    thrift_spec = (
    )

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()

    def write(self, oprot):
        oprot.writeStructBegin('ping_args')
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
             for key, value in self.__dict__.iteritems()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return (isinstance(other, self.__class__) and
                self.__dict__ == other.__dict__)

    def __ne__(self, other):
        return not (self == other)


class ping_result(object):

    thrift_spec = (
    )

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            print(ftype)
            if ftype == TType.STOP:
                break
            else:
                iprot.skip(ftype)
            iprot.readField()
        iprot.readStructEnd()

    def write(self, oprot):
        oprot.writeStructBegin('ping_result')
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
             for key, value in self.__dict__.iteritems()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return (isinstance(other, self.__class__) and
                self.__dict__ == other.__dict__)

    def __ne__(self, other):
        return not (self == other)


class blah_args(object):

    thrift_spec = (
        None,
        (1, TType.I32, 'num', None, None, ),
        (2, TType.STRING, 'Str', None, None, ),
        (3, TType.STRUCT, 'event', (Event, Event.thrift_spec), None, ),
    )

    def __init__(self, num=None, Str=None, event=None,):
        self.num = num
        self.Str = Str
        self.event = event

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 1:
                if ftype == TType.I32:
                    self.num = iprot.readI32()
                else:
                    iprot.skip(ftype)
            elif fid == 2:
                if ftype == TType.STRING:
                    self.Str = iprot.readString()
                else:
                    iprot.skip(ftype)
            elif fid == 3:
                if ftype == TType.STRUCT:
                    self.event = Event()
                    self.event.read(iprot)
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()

    def write(self, oprot):
        oprot.writeStructBegin('blah_args')
        if self.num is not None:
            oprot.writeFieldBegin('num', TType.I32, 1)
            oprot.writeI32(self.num)
            oprot.writeFieldEnd()
        if self.Str is not None:
            oprot.writeFieldBegin('Str', TType.STRING, 2)
            oprot.writeString(self.Str)
            oprot.writeFieldEnd()
        if self.event is not None:
            oprot.writeFieldBegin('event', TType.STRUCT, 3)
            self.event.write(oprot)
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value * 31) ^ hash(self.num)
        value = (value * 31) ^ hash(self.Str)
        value = (value * 31) ^ hash(self.event)
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
             for key, value in self.__dict__.iteritems()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return (isinstance(other, self.__class__) and
                self.__dict__ == other.__dict__)

    def __ne__(self, other):
        return not (self == other)


class blah_result(object):

    thrift_spec = (
        (0, TType.I64, 'success', None, None, ),
        (1, TType.STRUCT, 'awe', (AwesomeException,
                                  AwesomeException.thrift_spec), None, ),
        (2, TType.STRUCT, 'api', (api_exception,
                                  api_exception.thrift_spec), None, ),
    )

    def __init__(self, success=None, awe=None, api=None,):
        self.success = success
        self.awe = awe
        self.api = api

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 0:
                if ftype == TType.I64:
                    self.success = iprot.readI64()
                else:
                    iprot.skip(ftype)
            elif fid == 1:
                if ftype == TType.STRUCT:
                    self.awe = AwesomeException()
                    self.awe.read(iprot)
                else:
                    iprot.skip(ftype)
            elif fid == 2:
                if ftype == TType.STRUCT:
                    self.api = api_exception()
                    self.api.read(iprot)
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()

    def write(self, oprot):
        oprot.writeStructBegin('blah_result')
        if self.success is not None:
            oprot.writeFieldBegin('num', TType.I64, 0)
            oprot.writeI64(self.success)
            oprot.writeFieldEnd()
        if self.awe is not None:
            oprot.writeFieldBegin('awe', TType.STRUCT, 1)
            self.awe.write(oprot)
            oprot.writeFieldEnd()
        if self.event is not None:
            oprot.writeFieldBegin('api', TType.STRUCT, 2)
            self.api.write(oprot)
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value * 31) ^ hash(self.sucess)
        value = (value * 31) ^ hash(self.awe)
        value = (value * 31) ^ hash(self.api)
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
             for key, value in self.__dict__.iteritems()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return (isinstance(other, self.__class__) and
                self.__dict__ == other.__dict__)

    def __ne__(self, other):
        return not (self == other)


class one_way_args(object):

    thrift_spec = (
        None,
        (1, TType.I64, 'id', None, None),
        (2, TType.MAP, 'req', (TType.I32, None, TType.STRING, None), None, ),
    )

    def __init__(self, id=None, req=None,):
        self.id = id
        self.req = req

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 1:
                if ftype == TType.I64:
                    self.id = iprot.readI64()
                else:
                    iprot.skip(ftype)
            elif fid == 2:
                if ftype == TType.MAP:
                    self.req = {}
                    (_ktype1, _vtype2, _size0) = iprot.readMapBegin()
                    for _i4 in xrange(_size0):
                        _key5 = iprot.readI32()
                        _val6 = iprot.readString()
                        self.req[_key5] = _val6
                    iprot.readMapEnd()
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()

    def write(self, oprot):
        oprot.writeStructBegin('one_way_args')
        if self.id is not None:
            oprot.writeFieldBegin('id', TType.I64, 1)
            oprot.writeI64(self.id)
            oprot.writeFieldEnd()
        if self.req is not None:
            oprot.writeFieldBegin('req', TType.MAP, 2)
            oprot.writeMapBegin('req', TType.MAP, 2)
            for kiter7, viter8 in self.req.items():
                oprot.writeI32(kiter7)
                oprot.writeString(viter8)
            oprot.writeMapEnd()
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value*31) ^ hash(self.id)
        value = (value*31) ^ hash(self.req)
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
             for key, value in self.__dict__iteritems()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return (isinstance(other, self.__class__) and
                self.__dict__ == other.__dict__)

    def __ne__(self, other):
        return not (self == other)

