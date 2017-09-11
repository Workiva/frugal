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

import uuid
from copy import copy
from frugal import _IS_PY2

# Header containing correlation id.
_CID_HEADER = "_cid"

# Header containing op id (uint64 as string).
_OPID_HEADER = "_opid"

# Header containing request timeout (milliseconds as string).
_TIMEOUT_HEADER = "_timeout"

_DEFAULT_TIMEOUT = 5 * 1000

# Global incrementing op id.
_OP_ID = 0


class FContext(object):
    """
    FContext is the context for a Frugal message. Every RPC has an FContext,
    which can be used to set request headers, response headers, and the request
    timeout. The default timeout is five seconds. An FContext is also sent with
    every publish message which is then received by subscribers.

    In addition to headers, the FContext also contains a correlation ID which
    can be used for distributed tracing purposes. A random correlation ID is
    generated for each FContext if one is not provided.

    FContext also plays a key role in Frugal's multiplexing support. A unique,
    per-request operation ID is set on every FContext before a request is made.
    This operation ID is sent in the request and included in the response,
    which is then used to correlate a response to a request. The operation ID
    is an internal implementation detail and is not exposed to the user.

    An FContext should belong to a single request for the lifetime of that
    request. It can be reused once the request has completed, though they
    should generally not be reused. This class is _not_ thread-safe.
    """

    def __init__(self, correlation_id=None, timeout=_DEFAULT_TIMEOUT):
        """
        Initialize FContext.

        Args:
            correlation_id: string identifier for distributed tracing purposes.
            timeout: number of milliseconds before request times out.
        """
        self._request_headers = {}
        self._response_headers = {}

        if not correlation_id:
            correlation_id = self._generate_cid()

        self._request_headers[_CID_HEADER] = correlation_id
        self._request_headers[_TIMEOUT_HEADER] = str(timeout)

        # Take the current op id and increment the counter
        self._request_headers[_OPID_HEADER] = _get_next_op_id()

    @property
    def correlation_id(self):
        """
        Return the correlation id for the FContext. This is used for
        distributed tracing purposes.
        """
        return self._request_headers.get(_CID_HEADER)

    def _get_op_id(self):
        """
        Return an int operation id for the FContext.  This is a unique long
        per operation. This is protected as operation ids are an internal
        implementation detail.
        """

        return int(self._request_headers.get(_OPID_HEADER))

    def _set_op_id(self, op_id):
        self._request_headers[_OPID_HEADER] = str(op_id)

    def _set_response_op_id(self, op_id):
        self._response_headers[_OPID_HEADER] = op_id

    def get_request_headers(self):
        """
        Returns request headers for this FConext.
        """
        return copy(self._request_headers)

    def get_request_header(self, key):
        """
        Returns request header for the specified key from the request
        headers dict.
        """

        return self._request_headers.get(key)

    def set_request_header(self, key, value):
        """
        Set a string key value pair in the request headers dictionary.
        Return the same FContext to allow for call chaining. Changing the
        op ID or correlation ID is disallowed.

        Args:
            key: string key to set in request headers
            value: string value to set for the given key

        Returns:
            FContext

        Throws:
            TypeError: if user passes non-string for key or value.
        """
        self._check_string(key)
        self._check_string(value)

        self._request_headers[key] = value
        return self

    def get_response_headers(self):
        return copy(self._response_headers)

    def get_response_header(self, key):
        return self._response_headers.get(key)

    def set_response_header(self, key, value):
        """
        Set a string key value pair in the response headers dictionary.
        Return the same FContext to allow for call chaining. Changing the
        op ID or correlation ID is disallowed.

        Args:
            key: string key to set in response headers
            value: string value to set for the given key

        Returns:
            FContext

        Raises:
            TypeError: if user passes non-string for key or value.
        """
        self._check_string(key)
        self._check_string(value)

        self._response_headers[key] = value
        return self

    def get_timeout(self):
        """
        Get the timeout for the FContext.
        """
        return int(self._request_headers.get(_TIMEOUT_HEADER))

    def set_timeout(self, timeout):
        """
        Sets the timeout for the FContext.

        Args:
            timeout: number of seconds
        """
        self._request_headers[_TIMEOUT_HEADER] = str(timeout)

    @property
    def timeout(self):
        """
        Get the timeout for the FContext.
        """
        return int(self._request_headers.get(_TIMEOUT_HEADER))

    @timeout.setter
    def timeout(self, timeout):
        """
        Sets the timeout for the FContext.

        Args:
            timeout: number of seconds
        """
        # TODO: check the type of timeout
        self._request_headers[_TIMEOUT_HEADER] = str(timeout)
        return self

    def copy(self):
        """
        Performs a deep copy of an FContext while handling opids correctly.

        Returns:
            A new instance of FContext with identical headers, with the
            exception of _opid.
        """
        copied = FContext()
        copied._request_headers = self.get_request_headers()
        copied._response_headers = self.get_response_headers()
        copied._request_headers[_OPID_HEADER] = _get_next_op_id()
        return copied

    def _check_string(self, string):
        if _IS_PY2 and not \
                (isinstance(string, str) or isinstance(string, unicode)):  # noqa: F821,E501
            raise TypeError("Value should either be a string or unicode.")
        if not _IS_PY2 and not \
                (isinstance(string, str) or isinstance(string, bytes)):
            raise TypeError("Value should be either a string or bytes.")

    def _generate_cid(self):
        return uuid.uuid4().hex


def _get_next_op_id():
    global _OP_ID
    _OP_ID += 1
    return str(_OP_ID)
