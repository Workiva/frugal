from thrift.Thrift import TType, TException


class Event(object):

    thrift_spec = (
        (1, TType.I64, 'ID', None, -1, ),
        (2, TType.STRING, 'Message', None, None, ),
    )

    def __init__(self, ID=thrift_spec[1][4], Message=None,):
        if ID is self.thrift_spec[1][4]:
            ID = -1
        self.ID = ID
        self.Message = Message

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 1:
                if ftype == TType.I64:
                    self.ID = iprot.readI64()
                else:
                    iprot.skip(ftype)
            if fid == 2:
                if ftype == TType.STRING:
                    self.Message = iprot.readString()
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()

    def write(self, oprot):
        oprot.writeStructBegin('Event')
        if self.ID is not None:
            oprot.writeFieldBegin('ID', TType.I64, 1)
            oprot.writeI64(self.ID)
            oprot.writeFieldEnd()
        if self.Message is not None:
            oprot.writeFieldBegin('Message', TType.STRING, 2)
            oprot.writeString(self.Message)
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value * 31) ^ hash(self.ID)
        value = (value * 31) ^ hash(self.Message)
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


class AwesomeException(TException):

    thrift_spec = (
        (1, TType.I64, 'ID', None, -1, ),
        (2, TType.STRING, 'Reason', None, None, ),
    )

    def __init__(self, ID=None, Reason=None,):
        self.ID = ID
        self.Reason = Reason

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 1:
                if ftype == TType.I64:
                    self.ID = iprot.readI64()
                else:
                    iprot.skip(ftype)
            if fid == 2:
                if ftype == TType.STRING:
                    self.Reason = iprot.readString()
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()

    def write(self, oprot):
        oprot.writeStructBegin('AwesomeException')
        if self.ID is not None:
            oprot.writeFieldBegin('ID', TType.I64, 1)
            oprot.writeI64(self.ID)
            oprot.writeFieldEnd()
        if self.Message is not None:
            oprot.writeFieldBegin('Reason', TType.STRING, 2)
            oprot.writeString(self.Reason)
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value * 31) ^ hash(self.ID)
        value = (value * 31) ^ hash(self.Reason)
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


