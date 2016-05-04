import struct

from thrift.protocol.TProtocol import TProtocolBase

from frugal.context import FContext, _OP_ID
from frugal.exceptions import FrugalVersionException
from frugal.util.headers import _Headers

_V0 = 0


class FProtocol(TProtocolBase, object):
    """
    FProtocol is an extension of thrift TProtocol with the addition of headers
    """

    def __init__(self, wrapped_protocol):
        """Initialize FProtocol.

        Args:
            wrapped_protocol: wrapped thrift protocol extending TProtocolBase.
        """
        self._wrapped_protocol = wrapped_protocol
        super(FProtocol, self).__init__(self._wrapped_protocol.trans)

    def get_transport(self):
        """Return the extended TProtocolBase's underlying tranpsort

        Returns:
            TTransportBase
        """
        return self.trans

    def write_request_headers(self, context):
        """Write the request headers to the underlying TTranpsort."""

        self._write_headers(context.get_request_headers())

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

        for key, value in headers.iteritems():
            context._set_request_header(key, value)

        op_id = headers[_OP_ID]
        context._set_response_op_id(op_id)
        return context

    def read_response_headers(self, context):
        """Read the response headers from the underlying transport and set them
        on a given FContext

        Returns:
            FContext
        """
        headers = _Headers._read(self.get_transport())

        for key, value in headers.iteritems():
            context._set_response_header(key, value)

    # Thrift Transport pass through methods

    def writeMessageBegin(self, name, ttype, seqid):
        self._wrapped_protocol.writeMessageBegin(name, ttype, seqid)

    def writeMessageEnd(self):
        self._wrapped_protocol.writeMessageEnd()

    def writeStructBegin(self, name):
        self._wrapped_protocol.writeStructBegin(name)

    def writeStructEnd(self):
        self._wrapped_protocol.writeStructEnd()

    def writeFieldBegin(self, name, ttype, fid):
        self._wrapped_protocol.writeFieldBegin(name, ttype, fid)

    def writeFieldEnd(self):
        self._wrapped_protocol.writeFieldEnd()

    def writeFieldStop(self):
        self._wrapped_protocol.writeFieldStop()

    def writeMapBegin(self, ktype, vtype, size):
        self._wrapped_protocol.writeMapBegin(ktype, vtype, size)

    def writeMapEnd(self):
        self._wrapped_protocol.writeMapEnd()

    def writeListBegin(self, etype, size):
        self._wrapped_protocol.writeListBegin(etype, size)

    def writeListEnd(self):
        self._wrapped_protocol.writeListEnd()

    def writeSetBegin(self, etype, size):
        self._wrapped_protocol.writeSetBegin(etype, size)

    def writeSetEnd(self):
        self._wrapped_protocol.writeSetEnd()

    def writeBool(self, bool_val):
        self._wrapped_protocol.writeBool(bool_val)

    def writeByte(self, byte):
        self._wrapped_protocol.writeByte(byte)

    def writeI16(self, i16):
        self._wrapped_protocol.writeI16(i16)

    def writeI32(self, i32):
        self._wrapped_protocol.writeI32(i32)

    def writeI64(self, i64):
        self._wrapped_protocol.writeI64(i64)

    def writeDouble(self, dub):
        self._wrapped_protocol.writeDouble(dub)

    def writeString(self, str_val):
        self._wrapped_protocol.writeString(str_val)

    def readMessageBegin(self):
        return self._wrapped_protocol.readMessageBegin()

    def readMessageEnd(self):
        return self._wrapped_protocol.readMessageEnd()

    def readStructBegin(self):
        return self._wrapped_protocol.readStructBegin()

    def readStructEnd(self):
        return self._wrapped_protocol.readStructEnd()

    def readFieldBegin(self):
        return self._wrapped_protocol.readFieldBegin()

    def readFieldEnd(self):
        return self._wrapped_protocol.readFieldEnd()

    def readMapBegin(self):
        return self._wrapped_protocol.readMapBegin()

    def readMapEnd(self):
        return self._wrapped_protocol.readMapEnd()

    def readListBegin(self):
        return self._wrapped_protocol.readListBegin()

    def readListEnd(self):
        return self._wrapped_protocol.readListEnd()

    def readSetBegin(self):
        return self._wrapped_protocol.readSetBegin()

    def readSetEnd(self):
        return self._wrapped_protocol.readSetEnd()

    def readBool(self):
        return self._wrapped_protocol.readBool()

    def readByte(self):
        return self._wrapped_protocol.readByte()

    def readI16(self):
        return self._wrapped_protocol.readI16()

    def readI32(self):
        return self._wrapped_protocol.readI32()

    def readI64(self):
        return self._wrapped_protocol.readI64()

    def readDouble(self):
        return self._wrapped_protocol.readDouble()

    def readString(self):
        return self._wrapped_protocol.readString()
