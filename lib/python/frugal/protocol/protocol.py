import functools
import sys

from thrift.protocol.TBinaryProtocol import TBinaryProtocol, TBinaryProtocolFactory
from thrift.protocol.TCompactProtocol import CLEAR, TCompactProtocol, TCompactProtocolFactory
from thrift.protocol.TJSONProtocol import TJSONProtocol, TJSONProtocolFactory
from thrift.protocol.TProtocol import TProtocolBase, TProtocolException
from thrift.protocol.TProtocolDecorator import TProtocolDecorator

from frugal.context import FContext, _OPID_HEADER, _CID_HEADER, _get_next_op_id
from frugal.util.headers import _Headers

_V0 = 0


def _state_reset_decorator(func):
    """
    Decorator that resets the state of the TCompactProtocol as a hacky
    workaround for when an exception  occurs so the protocol can be reused, i.e.
    if "REQUEST_TOO_LARGE" error is thrown. This is only required for the
    compact protocol as other protocols don't track internal state as a sanity
    check.
    """
    @functools.wraps(func)
    def wrapper(self, *args, **kwargs):
        if not isinstance(self._wrapped_protocol, TCompactProtocol):
            return func(self, *args, **kwargs)

        try:
            return func(self, *args, **kwargs)
        except Exception:
            self._wrapped_protocol.state = CLEAR
            raise

    return wrapper


class FProtocol(TProtocolDecorator, object):
    """
    FProtocol is an extension of thrift TProtocol with the addition of headers
    """

    def __init__(self, wrapped_protocol):
        """Initialize FProtocol.

        Args:
            wrapped_protocol: wrapped thrift protocol extending TProtocolBase.
        """
        self._wrapped_protocol = wrapped_protocol
        super(FProtocol, self).__init__(self._wrapped_protocol)

    def get_transport(self):
        """Return the extended TProtocolBase's underlying tranpsort

        Returns:
            TTransportBase
        """
        return self.trans

    @_state_reset_decorator
    def write_request_headers(self, context):
        """Write the request headers to the underlying TTranpsort."""

        self._write_headers(context.get_request_headers())

    @_state_reset_decorator
    def write_response_headers(self, context):
        """Write the response headers to the underlying TTransport."""
        self._write_headers(context.get_response_headers())

    def _write_headers(self, headers):
        buff = _Headers._write_to_bytearray(headers)
        self.get_transport().write(buff)

    def read_request_headers(self):
        """Reads the request headers out of the underlying TTransportBase and
        return an FContext

        Returns:
            FContext
        """
        headers = _Headers._read(self.get_transport())

        context = FContext()

        for key, value in headers.items():
            context.set_request_header(key, value)

        op_id = headers[_OPID_HEADER]
        context._set_response_op_id(op_id)
        # Put a new opid in the request headers so this context an be
        # used/propagated on the receiver
        context.set_request_header(_OPID_HEADER, _get_next_op_id())

        cid = context.correlation_id
        if cid:
            context.set_response_header(_CID_HEADER, cid)
        return context

    def read_response_headers(self, context):
        """Read the response headers from the underlying transport and set them
        on a given FContext

        Returns:
            FContext
        """
        headers = _Headers._read(self.get_transport())

        for key, value in headers.items():
            # Don't want to overwrite the opid header we set for a propagated
            # response
            if key == _OPID_HEADER:
                continue
            context.set_response_header(key, value)

    @_state_reset_decorator
    def writeMessageBegin(self, name, ttype, seqid):
        self._wrapped_protocol.writeMessageBegin(name, ttype, seqid)

    @_state_reset_decorator
    def writeMessageEnd(self):
        self._wrapped_protocol.writeMessageEnd()

    @_state_reset_decorator
    def writeStructBegin(self, name):
        self._wrapped_protocol.writeStructBegin(name)

    @_state_reset_decorator
    def writeStructEnd(self):
        self._wrapped_protocol.writeStructEnd()

    @_state_reset_decorator
    def writeFieldBegin(self, name, ttype, fid):
        self._wrapped_protocol.writeFieldBegin(name, ttype, fid)

    @_state_reset_decorator
    def writeFieldEnd(self):
        self._wrapped_protocol.writeFieldEnd()

    @_state_reset_decorator
    def writeFieldStop(self):
        self._wrapped_protocol.writeFieldStop()

    @_state_reset_decorator
    def writeMapBegin(self, ktype, vtype, size):
        self._wrapped_protocol.writeMapBegin(ktype, vtype, size)

    @_state_reset_decorator
    def writeMapEnd(self):
        self._wrapped_protocol.writeMapEnd()

    @_state_reset_decorator
    def writeListBegin(self, etype, size):
        self._wrapped_protocol.writeListBegin(etype, size)

    @_state_reset_decorator
    def writeListEnd(self):
        self._wrapped_protocol.writeListEnd()

    @_state_reset_decorator
    def writeSetBegin(self, etype, size):
        self._wrapped_protocol.writeSetBegin(etype, size)

    @_state_reset_decorator
    def writeSetEnd(self):
        self._wrapped_protocol.writeSetEnd()

    @_state_reset_decorator
    def writeBool(self, bool_val):
        self._wrapped_protocol.writeBool(bool_val)

    @_state_reset_decorator
    def writeByte(self, byte):
        self._wrapped_protocol.writeByte(byte)

    @_state_reset_decorator
    def writeI16(self, i16):
        self._wrapped_protocol.writeI16(i16)

    @_state_reset_decorator
    def writeI32(self, i32):
        self._wrapped_protocol.writeI32(i32)

    @_state_reset_decorator
    def writeI64(self, i64):
        self._wrapped_protocol.writeI64(i64)

    @_state_reset_decorator
    def writeDouble(self, dub):
        self._wrapped_protocol.writeDouble(dub)

    @_state_reset_decorator
    def writeString(self, value):
        """
        Write a string to the protocol, if python 2, encode to utf-8
        bytes from a unicode string.
        """
        if sys.version_info[0] == 2:
            self._wrapped_protocol.writeString(value.encode('utf-8'))
        else:
            self._wrapped_protocol.writeString(value)

    @_state_reset_decorator
    def writeBinary(self, value):
        self._wrapped_protocol.writeBinary(value)

    def readString(self):
        """
        Read a string from the protocol, if python 2, decode from utf-8
        bytes to a unicode string.
        """
        if sys.version_info[0] == 2:
            return self._wrapped_protocol.readString().decode('utf-8')
        return self._wrapped_protocol.readString()


class FUniversalProtocol(TProtocolBase):
    """ Automatically detects and switches between Binary, Compact, and JSON Thrift protocols depending
        on the client's protocol.  Really ought to be used only as an input protocol...

        Note: If Binary is the expected protocol, be sure to initialize using 'strictRead=True'
              so the Binary protocol will be attempted first

        Args:
            trans: Passed to TProtocolBase.
                   Must be a supported transport type: TBufferedTransport, TFileObjectTransport,
                                                       TFramedTransport, TMemoryBuffer, or TSaslClientTransport
            kwargs: TBinary: strictRead, strictWrite, string_length_limit, container_length_limit
                    TCompact: string_length_limit, container_length_limit
    """
    def __init__(self, trans, **kwargs):
        TProtocolBase.__init__(self, trans)

        # Determine the transport's buffer property
        self.buffer = None
        for name in ('_buffer', '__rbuf'):
            if hasattr(trans, name):
                self.buffer = getattr(trans, name)
                break

        if not hasattr(self.buffer, 'seek'):
            raise ValueError("Only seekable transports are supported (trans = %s)" % (type(trans)))

        self.prot = None  # Not determined until initialize_protocol() / readMessageBegin()
        self.trans = trans
        self.prot_kwargs = kwargs.copy()

        self.binary_protocol = TBinaryProtocolFactory(**self.prot_kwargs).getProtocol(trans)
        self.compact_protocol = TCompactProtocolFactory(**self.prot_kwargs).getProtocol(trans)
        self.json_protocol = TJSONProtocolFactory().getProtocol(trans)

    def oprot_factory(self, trans):
        """ Returns a TProtocolBase suitable for use in the response which matches the detected request protocol """
        if not self.prot:
            self.initialize_protocol()

        if isinstance(self.prot, TBinaryProtocol):
            return TBinaryProtocolFactory(**self.prot_kwargs).getProtocol(trans)
        elif isinstance(self.prot, TCompactProtocol):
            return TCompactProtocolFactory(**self.prot_kwargs).getProtocol(trans)
        elif isinstance(self.prot, TJSONProtocol):
            return TJSONProtocolFactory().getProtocol(trans)

        raise TypeError("Unknown protocol or not initialized")

    def initialize_protocol(self):
        # If strictRead is not enabled for Binary, we need to try it last in order
        prot_order = (self.binary_protocol, self.compact_protocol, self.json_protocol)
        if not self.binary_protocol.strictRead:
            prot_order = (self.compact_protocol, self.json_protocol, self.binary_protocol)

        buffer_pos = self.buffer.tell()

        for prot in prot_order:
            try:
                prot.readMessageBegin()
                self.prot = prot
                self.buffer.seek(buffer_pos)
                return
            except TProtocolException:
                self.buffer.seek(buffer_pos)

        raise TProtocolException('Unknown or unsupported protocol')

    def readMessageBegin(self):
        if not self.prot:
            self.initialize_protocol()
        return self.prot.readMessageBegin()

    def writeMessageBegin(self, name, ttype, seqid):
        return self.prot.writeMessageBegin(self, name, ttype, seqid)

    def writeMessageEnd(self):
        return self.prot.writeMessageEnd()

    def writeStructBegin(self, name):
        return self.prot.writeStructBegin(name)

    def writeStructEnd(self):
        return self.prot.writeStructEnd

    def writeFieldBegin(self, name, ttype, fid):
        return self.prot.writeFieldBegin(name, ttype, fid)

    def writeFieldEnd(self):
        return self.prot.writeFieldEnd()

    def writeFieldStop(self):
        return self.prot.writeFieldStop()

    def writeMapBegin(self, ktype, vtype, size):
        return self.prot.writeMapBegin(ktype, vtype, size)

    def writeMapEnd(self):
        return self.prot.writeMapEnd()

    def writeListBegin(self, etype, size):
        return self.prot.writeListBegin(etype, size)

    def writeListEnd(self):
        return self.prot.writeListEnd()

    def writeSetBegin(self, etype, size):
        return self.prot.writeSetBegin(etype, size)

    def writeSetEnd(self):
        return self.prot.writeSetEnd()

    def writeBool(self, bool_val):
        return self.prot.writeBool(bool_val)

    def writeByte(self, byte):
        return self.prot.writeByte(byte)

    def writeI16(self, i16):
        return self.prot.writeI16(i16)

    def writeI32(self, i32):
        return self.prot.writeI32(i32)

    def writeI64(self, i64):
        return self.prot.writeI64(i64)

    def writeDouble(self, dub):
        return self.prot.writeDouble(dub)

    def writeBinary(self, str_val):
        return self.prot.writeBinary(str_val)

    def readMessageEnd(self):
        return self.prot.readMessageEnd()

    def readStructBegin(self):
        return self.prot.readStructBegin()

    def readStructEnd(self):
        return self.prot.readStructEnd()

    def readFieldBegin(self):
        return self.prot.readFieldBegin()

    def readFieldEnd(self):
        return self.prot.readFieldEnd()

    def readMapBegin(self):
        return self.prot.readMapBegin()

    def readMapEnd(self):
        return self.prot.readMapEnd()

    def readListBegin(self):
        return self.prot.readListBegin()

    def readListEnd(self):
        return self.prot.readListEnd()

    def readSetBegin(self):
        return self.prot.readSetBegin()

    def readSetEnd(self):
        return self.prot.readSetEnd()

    def readBool(self):
        return self.prot.readBool()

    def readByte(self):
        return self.prot.readByte()

    def readI16(self):
        return self.prot.readI16()

    def readI32(self):
        return self.prot.readI32()

    def readI64(self):
        return self.prot.readI64()

    def readDouble(self):
        return self.prot.readDouble()

    def readString(self):
        return self.prot.readString()

    def readBinary(self):
        return self.prot.readBinary()


