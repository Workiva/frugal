#
# Autogenerated by Frugal Compiler (3.15.4)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#

from thrift.Thrift import TType, TMessageType, TException, TApplicationException

from frugal.util import make_hashable
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol, TProtocol


class base_health_condition(int):
    PASS = 1
    WARN = 2
    FAIL = 3
    UNKNOWN = 4

    _VALUES_TO_NAMES = {
        1: "PASS",
        2: "WARN",
        3: "FAIL",
        4: "UNKNOWN",
    }

    _NAMES_TO_VALUES = {
        "PASS": 1,
        "WARN": 2,
        "FAIL": 3,
        "UNKNOWN": 4,
    }

class thing(object):
    """
    Attributes:
     - an_id
     - a_string
    """
    def __init__(self, an_id=None, a_string=None):
        self.an_id = an_id
        self.a_string = a_string

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 1:
                if ftype == TType.I32:
                    self.an_id = iprot.readI32()
                else:
                    iprot.skip(ftype)
            elif fid == 2:
                if ftype == TType.STRING:
                    self.a_string = iprot.readString()
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()
        self.validate()

    def write(self, oprot):
        self.validate()
        oprot.writeStructBegin('thing')
        if self.an_id is not None:
            oprot.writeFieldBegin('an_id', TType.I32, 1)
            oprot.writeI32(self.an_id)
            oprot.writeFieldEnd()
        if self.a_string is not None:
            oprot.writeFieldBegin('a_string', TType.STRING, 2)
            oprot.writeString(self.a_string)
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value * 31) ^ hash(make_hashable(self.an_id))
        value = (value * 31) ^ hash(make_hashable(self.a_string))
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
            for key, value in self.__dict__.items()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

    def __ne__(self, other):
        return not (self == other)

class nested_thing(object):
    """
    Attributes:
     - things
    """
    def __init__(self, things=None):
        self.things = things

    def read(self, iprot):
        iprot.readStructBegin()
        while True:
            (fname, ftype, fid) = iprot.readFieldBegin()
            if ftype == TType.STOP:
                break
            if fid == 1:
                if ftype == TType.LIST:
                    self.things = []
                    (_, elem73) = iprot.readListBegin()
                    for _ in range(elem73):
                        elem74 = thing()
                        elem74.read(iprot)
                        self.things.append(elem74)
                    iprot.readListEnd()
                else:
                    iprot.skip(ftype)
            else:
                iprot.skip(ftype)
            iprot.readFieldEnd()
        iprot.readStructEnd()
        self.validate()

    def write(self, oprot):
        self.validate()
        oprot.writeStructBegin('nested_thing')
        if self.things is not None:
            oprot.writeFieldBegin('things', TType.LIST, 1)
            oprot.writeListBegin(TType.STRUCT, len(self.things))
            for elem75 in self.things:
                elem75.write(oprot)
            oprot.writeListEnd()
            oprot.writeFieldEnd()
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __hash__(self):
        value = 17
        value = (value * 31) ^ hash(make_hashable(self.things))
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
            for key, value in self.__dict__.items()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

    def __ne__(self, other):
        return not (self == other)

class api_exception(TException):
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
        self.validate()

    def write(self, oprot):
        self.validate()
        oprot.writeStructBegin('api_exception')
        oprot.writeFieldStop()
        oprot.writeStructEnd()

    def validate(self):
        return

    def __str__(self):
        return repr(self)

    def __hash__(self):
        value = 17
        return value

    def __repr__(self):
        L = ['%s=%r' % (key, value)
            for key, value in self.__dict__.items()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

    def __ne__(self, other):
        return not (self == other)

