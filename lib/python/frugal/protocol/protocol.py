# Copyright 2017 Workiva
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import functools
import sys
from thrift.protocol.TProtocolDecorator import TProtocolDecorator
from thrift.protocol.TCompactProtocol import CLEAR, TCompactProtocol

from frugal.context import FContext, _OPID_HEADER, _CID_HEADER, _get_next_op_id
from frugal.util.headers import _Headers

_V0 = 0


def _state_reset_decorator(func):
    """
    Decorator that resets the state of the TCompactProtocol as a hacky
    workaround for when an exception  occurs so the protocol can be reused,
    i.e. if "REQUEST_TOO_LARGE" error is thrown. This is only required for
    the compact protocol as other protocols don't track internal state as a
    sanity check.
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
        """
        Initialize FProtocol.

        Args:
            wrapped_protocol: wrapped thrift protocol extending TProtocolBase.
        """
        self._wrapped_protocol = wrapped_protocol
        super(FProtocol, self).__init__(self._wrapped_protocol)

    def get_transport(self):
        """
        Return the extended TProtocolBase's underlying tranpsort

        Returns:
            TTransportBase
        """
        return self.trans

    @_state_reset_decorator
    def write_request_headers(self, context):
        """
        Write the request headers to the underlying TTranpsort.
        """

        self._write_headers(context.get_request_headers())

    @_state_reset_decorator
    def write_response_headers(self, context):
        """
        Write the response headers to the underlying TTransport.
        """

        self._write_headers(context.get_response_headers())

    def _write_headers(self, headers):
        buff = _Headers._write_to_bytearray(headers)
        self.get_transport().write(buff)

    def read_request_headers(self):
        """
        Reads the request headers out of the underlying TTransportBase and
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
        """
        Read the response headers from the underlying transport
        and set them on a given FContext.

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
        if sys.version_info[0] == 2 and isinstance(value, unicode):  # noqa: F821,E501
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
