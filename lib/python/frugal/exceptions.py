from thrift.protocol.TProtocol import TProtocolException
from thrift.transport.TTransport import TTransportException
from thrift.Thrift import TApplicationException
from thrift.Thrift import TException


class FException(TException):
    """Basic Frugal exception."""

    def __init__(self, message=None):
        super(FException, self).__init__(message)


class FApplicationException(TApplicationException):

    RESPONSE_TOO_LARGE = 100
    RATE_LIMIT_EXCEEDED = 102


class FTransportException(TTransportException):

    REQUEST_TOO_LARGE = 100
    RESPONSE_TOO_LARGE = 101


class FContextException(FException):
    """Indicates a problem with an FContext"""
    def __init__(self, message=None):
        super(FException, self).__init__(message=message)


class FContextHeaderException(FException):
    """Indicates an invalid header key on an FContext"""
    def __init__(self, message=None):
        super(FContextHeaderException, self).__init__(message)


class FProtocolException(TProtocolException):
    """Indicates a problem with a protocol."""
    def __init__(self, kind=TProtocolException.UNKNOWN, message=None):
        super(FProtocolException, self).__init__(type=kind, message=message)


class FMessageSizeException(TTransportException):
    """Indicates a message was too large for a transport to handle."""

    def __init__(self, type=FTransportException.REQUEST_TOO_LARGE,
                 message=None):
        super(FMessageSizeException, self).__init__(type=type, message=message)

    @classmethod
    def request(cls, message=None):
        return cls(type=FTransportException.REQUEST_TOO_LARGE, message=message)

    @classmethod
    def response(cls, message=None):
        return cls(type=FTransportException.RESPONSE_TOO_LARGE,
                   message=message)


class FTimeoutException(FException):
    """Indicates a request took too long."""
    def __init__(self, message=None):
        super(FTimeoutException, self).__init__(message=message)


class FRateLimitException(TApplicationException):
    """
    FRateLimitException indicates that an application has exceeded a rate
    limit threshold.
    """

    def __init__(self, message="rate limit exceeded"):
        """
        Args:
            message: exception message to provide with the rate limit error.
            Defaults to "rate limit exceeded".
        """
        super(FRateLimitException, self).__init__(
            type=FApplicationException.RATE_LIMIT_EXCEEDED, message=message)
