from thrift.Thrift import TType, TMessageType

import base.BaseFoo


class Iface(base.BaseFoo.Iface):

    def one_way(self, id, req):
        pass


class Client(base.BaseFoo.Client, Iface):

    def __init__(self, transport, iprot_factory, oprot_factory=None):
        base.BaseFoo.Client.__init__(self, transport,
                                     iprot_factory, oprot_factory)

    def one_way(self, id, req):
        self._seqid += 1
        self.send_one_way(id, req)

    def send_one_way(self, id, req):
        oprot = self._oprot_factory.get_protocol(self._transport)
        oprot.writeMessageBegin('oneWay', TMessageType.ONEWAY, self._seqid)
        args = one_way_args()
        args.id = id
        args.req = req
        args.write(oprot)
        oprot.writeMessageEnd()
        oprot.trans.flush()


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
