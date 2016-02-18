import struct

from thrift.protocol.TProtocol import TProtocolBase

from frugal.context import FContext
from frugal.exceptions import FrugalVersionException


class FProtocol(TProtocolBase):
    """
    FProtocol is an extension of thrift TProtocol with the addition of headers
    """

    def __init__(self, trans):
        """Initialize FProtocol.

        Args:
            trans: wrapped FTransport.
        """
        super(FProtocol, self).__init__(trans)

    def write_request_header(self, context):
        self._write_headers(context.get_request_headers())

    def read_request_header(self):
        headers = self._read_headers(self.trans)

        context = FContext()

        for key, value in headers.iteritems():
            context.put_request_header(key, value)

        op_id = headers['_opid']
        context.set_response_op_id(op_id)
        return context

    def write_response_header(self, context):
        self._write_headers(context.get_response_headers())

    def read_response_header(self):
        headers = self._read_headers(self.trans)

        context = FContext(headers['_cid'])

        for key, value in headers.iteritems():
            context.put_response_header(key, value)

        return context

    def _write_headers(self, headers):
        size = 0
        for key, value in headers.iteritems():
            size = size + 8 + len(key) + len(value)

        buff = bytearray(size + 5)

        # TODO: use V0 constant.
        struct.pack_into('>B', buff, 0, 0)
        struct.pack_into('>I', buff, 1, size)

        offset = 5

        for key, value in headers.iteritems():
            struct.pack_into('>I', buff, offset, len(key))
            offset += 4

            struct.pack_into('>{0}s'.format(str(len(key))), buff, offset, key)
            offset += len(key)

            struct.pack_into('>I', buff, offset, len(value))
            offset += 4

            struct.pack_into('>{0}s'.format(str(len(value))), buff,
                             offset, value)
            offset += len(value)

        self.trans.write(buff)

    def _read_headers(self, buff):
        parsed_headers = {}

        version = struct.unpack_from('>B', buff, 0)[0]

        # TODO: use constant.
        if version is not 0:
            raise FrugalVersionException("Wrong Frugal version.")

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

            parsed_headers[key] = val

        return parsed_headers
