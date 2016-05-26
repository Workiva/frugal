import unittest
import mock

from frugal.context import FContext, _DEFAULT_TIMEOUT
from frugal.exceptions import FContextHeaderException


class TestContext(unittest.TestCase):

    correlation_id = "fooid"

    def test_correlation_id(self):
        context = FContext("fooid")
        self.assertEqual("fooid", context.get_correlation_id())
        self.assertEqual(_DEFAULT_TIMEOUT, context.get_timeout())

    def test_timeout(self):
        context = FContext("fooid", 123)
        self.assertEqual(123, context.get_timeout())

    def test_op_id(self):
        context = FContext(self.correlation_id)
        context._set_request_header("_opid", "12345")
        self.assertEqual(self.correlation_id, context.get_correlation_id())
        self.assertEqual("12345", context.get_request_header("_opid"))

    def test_request_header(self):
        context = FContext(self.correlation_id)
        context.set_request_header("foo", "bar")
        self.assertEqual("bar", context.get_request_header("foo"))
        self.assertEqual(self.correlation_id,
                         context.get_request_header("_cid"))

    def test_response_header(self):
        context = FContext(self.correlation_id)
        context.set_response_header("foo", "bar")
        self.assertEqual("bar", context.get_response_header("foo"))
        self.assertEqual(self.correlation_id,
                         context.get_request_header("_cid"))

    def test_request_headers(self):
        context = FContext(self.correlation_id)
        context.set_request_header("foo", "bar")
        headers = context.get_request_headers()
        self.assertEqual("bar", headers.get('foo'))

    def test_response_headers(self):
        context = FContext(self.correlation_id)
        context.set_response_header("foo", "bar")
        headers = context.get_response_headers()
        self.assertEqual("bar", headers.get('foo'))

    def test_request_header_put_only_allows_string(self):
        context = FContext(self.correlation_id)
        self.assertRaises(TypeError, context.set_request_header, 1, "foo")
        self.assertRaises(TypeError, context.set_request_header, "foo", 3)

    def test_response_header_put_only_allows_string(self):
        context = FContext(self.correlation_id)
        self.assertRaises(TypeError, context.set_response_header, 1, "foo")
        self.assertRaises(TypeError, context.set_response_header, "foo", 3)

    def test_cant_set_cid_public_method(self):
        context = FContext(self.correlation_id)
        self.assertRaises(FContextHeaderException,
                          context.set_request_header, "_cid", "foo")

    def test_cant_set_opid_public_method(self):
        context = FContext(self.correlation_id)
        self.assertRaises(FContextHeaderException,
                          context.set_request_header, "_opid", "foo")
