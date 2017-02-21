from thrift.protocol.TProtocolDecorator import TProtocolDecorator

from frugal.context import FContext, _OPID_HEADER, _CID_HEADER
from frugal.util.headers import _Headers

_V0 = 0


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

        for key, value in headers.items():
            context.set_request_header(key, value)

        op_id = headers[_OPID_HEADER]
        context._set_response_op_id(op_id)
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
            context.set_response_header(key, value)
