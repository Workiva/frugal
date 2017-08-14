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

from thrift.Thrift import TApplicationException
from thrift.transport.TTransport import TTransportException


class TTransportExceptionType(object):
    """Exception types for TTransportExceptions"""
    UNKNOWN = TTransportException.UNKNOWN
    NOT_OPEN = TTransportException.NOT_OPEN
    ALREADY_OPEN = TTransportException.ALREADY_OPEN
    TIMED_OUT = TTransportException.TIMED_OUT
    END_OF_FILE = TTransportException.END_OF_FILE

    REQUEST_TOO_LARGE = 100
    RESPONSE_TOO_LARGE = 101


class TApplicationExceptionType(object):
    """Exception types for TApplicationExceptions"""
    UNKNOWN = TApplicationException.UNKNOWN
    UNKNOWN_METHOD = TApplicationException.UNKNOWN_METHOD
    INVALID_MESSAGE_TYPE = TApplicationException.INVALID_MESSAGE_TYPE
    WRONG_METHOD_NATE = TApplicationException.WRONG_METHOD_NAME
    BAD_SEQUENCE_ID = TApplicationException.BAD_SEQUENCE_ID
    MISSING_RESULT = TApplicationException.MISSING_RESULT
    INTERNAL_ERROR = TApplicationException.INTERNAL_ERROR
    PROTOCOL_ERROR = TApplicationException.PROTOCOL_ERROR
    INVALID_TRANSFORM = TApplicationException.INVALID_TRANSFORM
    INVALID_PROTOCOL = TApplicationException.INVALID_PROTOCOL
    UNSUPPORTED_CLIENT_TYPE = TApplicationException.UNSUPPORTED_CLIENT_TYPE

    RESPONSE_TOO_LARGE = 100
