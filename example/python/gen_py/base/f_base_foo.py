from threading import Lock

from thrift.Thrift import TMessageType, TType, TApplicationException
from tornado import concurrent

from frugal.processor.processor import FProcessor
from frugal.registry import FClientRegistry


class Iface(object):

    def base_ping(context):
        pass


class Client(Iface):

    def __init__(self, transport, protocol_factory):
        """Initialize a Frugal Client

        Args:
            transport: FTransport
            protocl_factory: FProtocolFactory for creating FProtocols
        """

        self._transport = transport
        self._transport.set_registry(FClientRegistry())
        self._protocol_factory = protocol_factory
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def base_ping(self, context):
        """ base ping

        Args:
            context: FContext
        """
        future = concurrent.Future()
        self.send_basePing(context, future)
        return future

    def send_basePing(self, context, future):
        oprot = self._oprot
        self._transport.register(context, self._recv_basePing(context, future))

        with self._write_lock:
            oprot.write_request_headers(context)
            oprot.writeMessageBegin('basePing', TMessageType.CALL, 0)
            args = basePing_args()
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def recv_basePing(self, context, future):
        def basePing_callback(transport):
            iprot = self._protocol_factory.get_protocol(transport)
            iprot.read_response_headers(context)
            (fname, mtype, fid) = iprot.readMessageBegin()
            if mtype == TMessageType.EXCEPTION:
                x = TApplicationException()
                x.read(iprot)
                iprot.readMessageEnd()
                future.set_exception(x)
                raise x
            result = basePing_result()
            result.read(iprot)
            iprot.readMessageEnd()
            future.set_result('')
        return basePing_callback


class Processor(Iface, FProcessor):

    def __init__(self, handler):
        self._handler = handler
        self._process_map = {}
        self._process_map["basePing"] = Processor.process_basePing

    def process(self, context, iprot, oprot):
        (name, type, seqid) = iprot.readMessageBegin()
        if name not in self._process_map:
            iprot.skip(TType.STRUCT)
            iprot.readMessageEnd()
            x = TApplicationException(TApplicationException.UNKNOWN_METHOD,
                                      "Unknown function {}".format(name))
            oprot.writeMessageBegin(name, TMessageType.EXCEPTION, seqid)
            x.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()
        else:
            return


class basePing_args(object):

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
        oprot.writeStructBegin('basePing_args')
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


class basePing_result(object):

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
        oprot.writeStructBegin('basePing_result')
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

