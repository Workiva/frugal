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
        headers = _Headers._read(self.get_transport())

        for key, value in headers.iteritems():
            context._set_response_header(key, value)

        return context

    def write_request_headers(self, context):
        """Write the request headers to the underlying TTranpsort."""

        self._write_headers(context.get_request_headers())

    def write_response_headers(self, context):
        """Write the response headers to the underlying TTransport."""

        self._write_headers(context.get_response_headers())

    def _write_headers(self, headers):
        buff = _Headers._write_to_bytearray(headers)

        self.get_transport().write(buff)

    # Thrift Transport pass through methods

    def writeMessageBegin(self, name, ttype, seqid):
        self._wrapped_protocol.writeMessageBegin(name, ttype, seqid)

    def writeMessageEnd(self):
        self._wrapped_protocol.writeMessageEnd()

    def writeStructBegin(self, name):
        self._wrapped_protocol.writeStructBegin(name)

    def writeStructEnd(self):
        self._wrapped_protocol.writeStructEnd()

    def writeFieldStop(self):
        self._wrapped_protocol.writeFieldStop()

    def readMessageBegin(self):
        self._wrapped_protocol.readMessageBegin()

    def readStructBegin(self):
        self._wrapped_protocol.readStructBegin()

    def readFieldBegin(self):
        self._wrapped_protocol.readFieldBegin()

    def readField(self):
        self._wrapped_protocol.readField()

    def readStructEnd(self):
        self._wrapped_protocol.readStructEnd()
