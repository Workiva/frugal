
class FException(Exception):

    def __init__(self, message=None):
        super(FException, self).__init__(message)


class FrugalVersionException(FException):

    def __init__(self, message=None):
        super(FrugalVersionException, self).__init__(message)


class FContextHeaderException(FException):

    def __init__(self, message=None):
        super(FContextHeaderException, self).__init__(message)
