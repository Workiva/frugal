#
# Autogenerated by Thrift Compiler (0.9.3-wk-3)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#
#  options string: py:tornado
#

from thrift.Thrift import TType, TMessageType, TException, TApplicationException
import base.BaseFoo
import logging
from ttypes import *
from thrift.Thrift import TProcessor
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol, TProtocol
try:
  from thrift.protocol import fastbinary
except:
  fastbinary = None

from tornado import gen
from tornado import concurrent
from thrift.transport import TTransport

class Iface(base.BaseFoo.Iface):
  """
  This is a thrift service. Frugal will generate bindings that include
  a frugal Context for each service call.
  """
  def ping(self):
    """
    Ping the server.
    """
    pass

  def blah(self, num, Str, event):
    """
    Blah the server.

    Parameters:
     - num
     - Str
     - event
    """
    pass

  def oneWay(self, id, req):
    """
    oneway methods don't receive a response from the server.

    Parameters:
     - id
     - req
    """
    pass


class Client(base.BaseFoo.Client, Iface):
  """
  This is a thrift service. Frugal will generate bindings that include
  a frugal Context for each service call.
  """
  def __init__(self, transport, iprot_factory, oprot_factory=None):
    base.BaseFoo.Client.__init__(self, transport, iprot_factory, oprot_factory)

  def ping(self):
    """
    Ping the server.
    """
    self._seqid += 1
    future = self._reqs[self._seqid] = concurrent.Future()
    self.send_ping()
    return future

  def send_ping(self):
    oprot = self._oprot_factory.getProtocol(self._transport)
    oprot.writeMessageBegin('ping', TMessageType.CALL, self._seqid)
    args = ping_args()
    args.write(oprot)
    oprot.writeMessageEnd()
    oprot.trans.flush()

  def recv_ping(self, iprot, mtype, rseqid):
    if mtype == TMessageType.EXCEPTION:
      x = TApplicationException()
      x.read(iprot)
      iprot.readMessageEnd()
      raise x
    result = ping_result()
    result.read(iprot)
    iprot.readMessageEnd()
    return

  def blah(self, num, Str, event):
    """
    Blah the server.

    Parameters:
     - num
     - Str
     - event
    """
    self._seqid += 1
    future = self._reqs[self._seqid] = concurrent.Future()
    self.send_blah(num, Str, event)
    return future

  def send_blah(self, num, Str, event):
    oprot = self._oprot_factory.getProtocol(self._transport)
    oprot.writeMessageBegin('blah', TMessageType.CALL, self._seqid)
    args = blah_args()
    args.num = num
    args.Str = Str
    args.event = event
    args.write(oprot)
    oprot.writeMessageEnd()
    oprot.trans.flush()

  def recv_blah(self, iprot, mtype, rseqid):
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
    raise TApplicationException(TApplicationException.MISSING_RESULT, "blah failed: unknown result")

  def oneWay(self, id, req):
    """
    oneway methods don't receive a response from the server.

    Parameters:
     - id
     - req
    """
    self._seqid += 1
    self.send_oneWay(id, req)

  def send_oneWay(self, id, req):
    oprot = self._oprot_factory.getProtocol(self._transport)
    oprot.writeMessageBegin('oneWay', TMessageType.ONEWAY, self._seqid)
    args = oneWay_args()
    args.id = id
    args.req = req
    args.write(oprot)
    oprot.writeMessageEnd()
    oprot.trans.flush()

class Processor(base.BaseFoo.Processor, Iface, TProcessor):
  def __init__(self, handler):
    base.BaseFoo.Processor.__init__(self, handler)
    self._processMap["ping"] = Processor.process_ping
    self._processMap["blah"] = Processor.process_blah
    self._processMap["oneWay"] = Processor.process_oneWay

  def process(self, iprot, oprot):
    (name, type, seqid) = iprot.readMessageBegin()
    if name not in self._processMap:
      iprot.skip(TType.STRUCT)
      iprot.readMessageEnd()
      x = TApplicationException(TApplicationException.UNKNOWN_METHOD, 'Unknown function %s' % (name))
      oprot.writeMessageBegin(name, TMessageType.EXCEPTION, seqid)
      x.write(oprot)
      oprot.writeMessageEnd()
      oprot.trans.flush()
      return
    else:
      return self._processMap[name](self, seqid, iprot, oprot)

  @gen.coroutine
  def process_ping(self, seqid, iprot, oprot):
    args = ping_args()
    args.read(iprot)
    iprot.readMessageEnd()
    result = ping_result()
    yield gen.maybe_future(self._handler.ping())
    oprot.writeMessageBegin("ping", TMessageType.REPLY, seqid)
    result.write(oprot)
    oprot.writeMessageEnd()
    oprot.trans.flush()

  @gen.coroutine
  def process_blah(self, seqid, iprot, oprot):
    args = blah_args()
    args.read(iprot)
    iprot.readMessageEnd()
    result = blah_result()
    try:
      result.success = yield gen.maybe_future(self._handler.blah(args.num, args.Str, args.event))
    except AwesomeException as awe:
      result.awe = awe
    except base.ttypes.api_exception as api:
      result.api = api
    oprot.writeMessageBegin("blah", TMessageType.REPLY, seqid)
    result.write(oprot)
    oprot.writeMessageEnd()
    oprot.trans.flush()

  @gen.coroutine
  def process_oneWay(self, seqid, iprot, oprot):
    args = oneWay_args()
    args.read(iprot)
    iprot.readMessageEnd()
    yield gen.maybe_future(self._handler.oneWay(args.id, args.req))


# HELPER FUNCTIONS AND STRUCTURES

class ping_args:

  thrift_spec = (
  )

  def read(self, iprot):
    if iprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and isinstance(iprot.trans, TTransport.CReadableTransport) and self.thrift_spec is not None and fastbinary is not None:
      fastbinary.decode_binary(self, iprot.trans, (self.__class__, self.thrift_spec))
      return
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
    if oprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and self.thrift_spec is not None and fastbinary is not None:
      oprot.trans.write(fastbinary.encode_binary(self, (self.__class__, self.thrift_spec)))
      return
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
    return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

  def __ne__(self, other):
    return not (self == other)

class ping_result:

  thrift_spec = (
  )

  def read(self, iprot):
    if iprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and isinstance(iprot.trans, TTransport.CReadableTransport) and self.thrift_spec is not None and fastbinary is not None:
      fastbinary.decode_binary(self, iprot.trans, (self.__class__, self.thrift_spec))
      return
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
    if oprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and self.thrift_spec is not None and fastbinary is not None:
      oprot.trans.write(fastbinary.encode_binary(self, (self.__class__, self.thrift_spec)))
      return
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
    return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

  def __ne__(self, other):
    return not (self == other)

class blah_args:
  """
  Attributes:
   - num
   - Str
   - event
  """

  thrift_spec = (
    None, # 0
    (1, TType.I32, 'num', None, None, ), # 1
    (2, TType.STRING, 'Str', None, None, ), # 2
    (3, TType.STRUCT, 'event', (Event, Event.thrift_spec), None, ), # 3
  )

  def __init__(self, num=None, Str=None, event=None,):
    self.num = num
    self.Str = Str
    self.event = event

  def read(self, iprot):
    if iprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and isinstance(iprot.trans, TTransport.CReadableTransport) and self.thrift_spec is not None and fastbinary is not None:
      fastbinary.decode_binary(self, iprot.trans, (self.__class__, self.thrift_spec))
      return
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
    if oprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and self.thrift_spec is not None and fastbinary is not None:
      oprot.trans.write(fastbinary.encode_binary(self, (self.__class__, self.thrift_spec)))
      return
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
    return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

  def __ne__(self, other):
    return not (self == other)

class blah_result:
  """
  Attributes:
   - success
   - awe
   - api
  """

  thrift_spec = (
    (0, TType.I64, 'success', None, None, ), # 0
    (1, TType.STRUCT, 'awe', (AwesomeException, AwesomeException.thrift_spec), None, ), # 1
    (2, TType.STRUCT, 'api', (base.ttypes.api_exception, base.ttypes.api_exception.thrift_spec), None, ), # 2
  )

  def __init__(self, success=None, awe=None, api=None,):
    self.success = success
    self.awe = awe
    self.api = api

  def read(self, iprot):
    if iprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and isinstance(iprot.trans, TTransport.CReadableTransport) and self.thrift_spec is not None and fastbinary is not None:
      fastbinary.decode_binary(self, iprot.trans, (self.__class__, self.thrift_spec))
      return
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
          self.api = base.ttypes.api_exception()
          self.api.read(iprot)
        else:
          iprot.skip(ftype)
      else:
        iprot.skip(ftype)
      iprot.readFieldEnd()
    iprot.readStructEnd()

  def write(self, oprot):
    if oprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and self.thrift_spec is not None and fastbinary is not None:
      oprot.trans.write(fastbinary.encode_binary(self, (self.__class__, self.thrift_spec)))
      return
    oprot.writeStructBegin('blah_result')
    if self.success is not None:
      oprot.writeFieldBegin('success', TType.I64, 0)
      oprot.writeI64(self.success)
      oprot.writeFieldEnd()
    if self.awe is not None:
      oprot.writeFieldBegin('awe', TType.STRUCT, 1)
      self.awe.write(oprot)
      oprot.writeFieldEnd()
    if self.api is not None:
      oprot.writeFieldBegin('api', TType.STRUCT, 2)
      self.api.write(oprot)
      oprot.writeFieldEnd()
    oprot.writeFieldStop()
    oprot.writeStructEnd()

  def validate(self):
    return


  def __hash__(self):
    value = 17
    value = (value * 31) ^ hash(self.success)
    value = (value * 31) ^ hash(self.awe)
    value = (value * 31) ^ hash(self.api)
    return value

  def __repr__(self):
    L = ['%s=%r' % (key, value)
      for key, value in self.__dict__.iteritems()]
    return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

  def __eq__(self, other):
    return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

  def __ne__(self, other):
    return not (self == other)

class oneWay_args:
  """
  Attributes:
   - id
   - req
  """

  thrift_spec = (
    None, # 0
    (1, TType.I64, 'id', None, None, ), # 1
    (2, TType.MAP, 'req', (TType.I32,None,TType.STRING,None), None, ), # 2
  )

  def __init__(self, id=None, req=None,):
    self.id = id
    self.req = req

  def read(self, iprot):
    if iprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and isinstance(iprot.trans, TTransport.CReadableTransport) and self.thrift_spec is not None and fastbinary is not None:
      fastbinary.decode_binary(self, iprot.trans, (self.__class__, self.thrift_spec))
      return
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
          (_ktype1, _vtype2, _size0 ) = iprot.readMapBegin()
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
    if oprot.__class__ == TBinaryProtocol.TBinaryProtocolAccelerated and self.thrift_spec is not None and fastbinary is not None:
      oprot.trans.write(fastbinary.encode_binary(self, (self.__class__, self.thrift_spec)))
      return
    oprot.writeStructBegin('oneWay_args')
    if self.id is not None:
      oprot.writeFieldBegin('id', TType.I64, 1)
      oprot.writeI64(self.id)
      oprot.writeFieldEnd()
    if self.req is not None:
      oprot.writeFieldBegin('req', TType.MAP, 2)
      oprot.writeMapBegin(TType.I32, TType.STRING, len(self.req))
      for kiter7,viter8 in self.req.items():
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
    value = (value * 31) ^ hash(self.id)
    value = (value * 31) ^ hash(self.req)
    return value

  def __repr__(self):
    L = ['%s=%r' % (key, value)
      for key, value in self.__dict__.iteritems()]
    return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

  def __eq__(self, other):
    return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

  def __ne__(self, other):
    return not (self == other)
