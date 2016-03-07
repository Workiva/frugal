from threading import Lock

from gen_py.base import f_base_foo, base_foo

from frugal.registry import FClientRegistry
from thrift.Thrift import TType, TMessageType


class Iface(f_base_foo.Iface):

    def one_way(self, id, req):
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

    def one_way(self, ctx, id, req):
        """ oneway methods don't receive a response from the server

        Args:
            ctx: FContext
            req: dict key values to send (will be converted to JSON string)
        """
        oprot = self._oprot
        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin("one_way", TMessageType.ONEWAY, 0)
            args = one_way_args()
            args.id = id
            args.req = req
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()


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

