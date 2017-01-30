from thrift.Thrift import TApplicationException
from thrift.transport.TTransport import TTransportException


class FrugalTTransportExceptionType(object):
    """Exception types for TTransportExceptions"""
    UNKNOWN = TTransportException.UNKNOWN
    NOT_OPEN = TTransportException.NOT_OPEN
    ALREADY_OPEN = TTransportException.ALREADY_OPEN
    TIMED_OUT = TTransportException.TIMED_OUT
    END_OF_FILE = TTransportException.END_OF_FILE

    REQUEST_TOO_LARGE = 100
    RESPONSE_TOO_LARGE = 101


class FrugalTApplicationExceptionType(object):
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
