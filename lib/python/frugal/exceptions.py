
class FException(Exception):

    def __init__(self, message=None):
        super(FException, self).__init__(message)


class FrugalVersionException(FException):

    def __init__(self, message=None):
        super(FrugalVersionException, self).__init__(message)


class FContextHeaderException(FException):

    def __init__(self, message=None):
        super(FContextHeaderException, self).__init__(message)


class FProtocolException(FException):

    UNKNOWN = 0
    INVALID_DATA = 1
    BAD_VERSION = 2

    def __init__(self, type=UNKNOWN, message=None):
        super(FProtocolException, self).__init__(message)
        self.type = type


class FMessageSizeException(FException):

    def __init__(self, message=None):
        super(FMessageSizeException, self).__init__(message)


class FOperationIdNotFoundException(FException):

    def __init__(self, message=None):
        super(FOperationIdNotFoundException, self).__init__(message)
