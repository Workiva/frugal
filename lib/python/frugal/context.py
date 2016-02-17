import uuid
from copy import copy

_C_ID = "_cid"
_OP_ID = "_opid"
_DEFAULT_TIMEOUT = 60 * 1000


class FContext(object):
    """FContext is the message context for a frugal message."""

    def __init__(self, correlation_id=None):
        self._request_headers = {}
        self._response_headers = {}

        if not correlation_id:
            correlation_id = self._generate_cid()

        self._request_headers[_C_ID] = correlation_id

    def get_correlation_id(self):
        """Returns the correlation id for the FContext.
           This is used for distributed tracing purposes.
        """

        return self._request_headers[_C_ID]

    def get_op_id(self):
        """Returns the operation id for the FContext.  This is a unique long per
        operation.  This is protected as operation ids are an internal
        implementation detail.
        """

        return self._request_headers[_OP_ID]

    def set_op_id(self, op_id):
        self._request_headers[_OP_ID] = op_id

    def get_request_headers(self):
        return copy(self._request_headers)

    def get_request_header(self, key):
        """Returns request header for the specified key from the request
        headers dict.
        """

        return self._request_headers[key]

    def put_request_header(self, key, value):
        self._check_string(key)
        self._check_string(value)

        self._request_headers[key] = value

    def get_response_headers(self):
        return copy(self._response_headers)

    def get_response_header(self, key):
        return self._response_headers[key]

    def put_response_header(self, key, value):
        self._check_string(key)
        self._check_string(value)

        self._response_headers[key] = value

    def get_timeout(self):
        return self._timeout or _DEFAULT_TIMEOUT

    def set_timeout(self, timeout):
        self._timeout = timeout

    def _check_string(self, string):
        if not isinstance(string, str):
            raise TypeError("Value should be a string.")

    def _generate_cid(self):
        return str(uuid.uuid4()).replace('-', '')

