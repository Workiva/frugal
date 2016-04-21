import struct

from thrift.protocol.TProtocol import TProtocolBase

from frugal.context import FContext, _OP_ID
from frugal.exceptions import FrugalVersionException


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
        return self.trans

    def write_request_headers(self, context):
        self._write_headers(context.get_request_headers())

    def read_request_headers(self):
        headers = self._read_headers(self.trans)

        context = FContext()

        for key, value in headers.iteritems():
            context._set_request_header(key, value)

        op_id = headers[_OP_ID]
        context.set_response_op_id(op_id)
        return context

    def write_response_headers(self, context):
        self._write_headers(context.get_response_headers())

    def read_response_headers(self, context):
        headers = self._read_headers(self.trans)

        for key, value in headers.iteritems():
            context._set_response_header(key, value)

        return context

    def _write_headers(self, headers):
        size = 0
        offset = 5

        for key, value in headers.iteritems():
            size = size + 8 + len(key) + len(value)

        buff = bytearray(size + offset)

        struct.pack_into('>B', buff, 0, _V0)
        struct.pack_into('>I', buff, 1, size)

        for key, value in headers.iteritems():
            key_length = len(key)
            struct.pack_into('>I', buff, offset, key_length)
            offset += 4

            struct.pack_into('>{0}s'.format(str(key_length)), buff, offset, key)
            offset += len(key)

            struct.pack_into('>I', buff, offset, len(value))
            offset += 4

            struct.pack_into('>{0}s'.format(str(len(value))), buff,
                             offset, value)
            offset += len(value)

        self.get_transport().write(buff)

    def _read_headers(self, buff1):
        buff = buff1.getvalue()
        parsed_headers = {}
        version = struct.unpack_from('>B', buff, 0)[0]
        print("version: {}".format(version))
        if version is not _V0:
            raise FrugalVersionException(
                "Wrong Frugal version.  Found version {0}.  Wanted version {1}"
                .format(version, _V0))

        size = struct.unpack_from('>I', buff, 1)[0]

        offset = 5  # since size is 4 bytes

        while offset < size:
            key_size = struct.unpack_from('>I', buff, offset)[0]
            offset += 4

            # TODO: Check bounds.

            key = struct.unpack_from('>{0}s'.format(key_size), buff, offset)[0]
            offset += len(key)

            # TODO: Check bounds.

            val_size = struct.unpack_from('>I', buff, offset)[0]
            offset += 4

            # TODO: Check bounds.

            val = struct.unpack_from('>{0}s'.format(val_size), buff, offset)[0]
            offset += len(val)
            print("key_size {} key {} val_size {} val {}".format(key_size, key, val_size, val))
            parsed_headers[key] = val

        return parsed_headers

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
