import uuid
from copy import copy
from frugal.exceptions import FContextHeaderException

_C_ID = "_cid"
_OP_ID = "_opid"
_DEFAULT_TIMEOUT = 60 * 1000
_DEFAULT_OP_ID = 0


class FContext(object):
    """FContext is the message context for a Frugal message."""

    def __init__(self, correlation_id=None, timeout=_DEFAULT_TIMEOUT):
        """Initialize FContext.

        Args:
            correlation_id: string identifier for distributed tracing purposes.
            timeout: number of milliseconds before request times out.
        """
        self._request_headers = {}
        self._response_headers = {}

        self._timeout = timeout

        if not correlation_id:
            correlation_id = self._generate_cid()

        self._request_headers[_C_ID] = correlation_id
        self._request_headers[_OP_ID] = str(_DEFAULT_OP_ID)

    def get_correlation_id(self):
        """Return the correlation id for the FContext.
           This is used for distributed tracing purposes.
        """

        return self._request_headers.get(_C_ID)

    def _get_op_id(self):
        """Return an int operation id for the FContext.  This is a unique long
        per operation.  This is protected as operation ids are an internal
        implementation detail.
        """

        return int(self._request_headers.get(_OP_ID))

    def _set_op_id(self, op_id):
        self._request_headers[_OP_ID] = str(op_id)

    def _set_response_op_id(self, op_id):
        self._response_headers[_OP_ID] = str(op_id)

    def get_request_headers(self):
        """Returns request headers for this FConext."""
        return copy(self._request_headers)

    def get_request_header(self, key):
        """Returns request header for the specified key from the request
        headers dict.
        """

        return self._request_headers.get(key)

    def set_request_header(self, key, value):
        """Set a string key value pair in the request headers dictionary.
        Return the same FContext to allow for call chaining.

        Args:
            key: string key to set in request headers
            value: string value to set for the given key

        Returns:
            FContext

        Throws:
            FContextHeaderException: if user tries to set _cid or _opid.
            TypeError: if user passes non-string for key or value.
        """
        if key in (_OP_ID, _C_ID):
            raise FContextHeaderException(
                "Not allowed to overwrite internal _cid or _opid.")

        self._set_request_header(key, value)
        return self

    def _set_request_header(self, key, value):
        self._check_string(key)
        self._check_string(value)

        self._request_headers[key] = value

    def get_response_headers(self):
        return copy(self._response_headers)

    def get_response_header(self, key):
        return self._response_headers.get(key)

    def set_response_header(self, key, value):
        """Set a string key value pair in the response headers dictionary.
        Return the same FContext to allow for call chaining.

        Args:
            key: string key to set in response headers
            value: string value to set for the given key

        Returns:
            FContext

        Raises:
            FContextHeaderException: if user tries to set _cid or _opid.
            TypeError: if user passes non-string for key or value.
        """
        if key in (_OP_ID, _C_ID):
            raise FContextHeaderException(
                "Not allowed to overwrite internal _cid or _opid")

        self._set_response_header(key, value)
        return self

    def _set_response_header(self, key, value):
        self._check_string(key)
        self._check_string(value)

        self._response_headers[key] = value

    def get_timeout(self):
        return self._timeout

    def set_timeout(self, timeout):
        # TODO: check the type of timeout
        self._timeout = timeout
        return self

    def _check_string(self, string):
        if not isinstance(string, str):
            raise TypeError("Value should be a string.")

    def _generate_cid(self):
        return uuid.uuid4().hex

