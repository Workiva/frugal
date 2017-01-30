import unittest

from frugal.context import FContext, _DEFAULT_TIMEOUT


class TestContext(unittest.TestCase):

    correlation_id = "fooid"

    def test_correlation_id(self):
        context = FContext("fooid")
        self.assertEqual("fooid", context.correlation_id)
        self.assertEqual(_DEFAULT_TIMEOUT, context.timeout)

    def test_timeout(self):
        # Check default timeout (5 seconds).
        context = FContext()
        self.assertEqual(5000, context.timeout)
        self.assertEqual("5000", context.get_request_header("_timeout"))

        # Set timeout and check expected values.
        context.set_timeout(10000)
        self.assertEqual(10000, context.timeout)
        self.assertEqual("10000", context.get_request_header("_timeout"))

        # Check timeout passed to constructor.
        context = FContext(timeout=1000)
        self.assertEqual(1000, context.timeout)
        self.assertEqual("1000", context.get_request_header("_timeout"))

    def test_op_id(self):
        context = FContext(self.correlation_id)
        context._set_request_header("_opid", "12345")
        self.assertEqual(self.correlation_id, context.correlation_id)
        self.assertEqual("12345", context.get_request_header("_opid"))

    def test_request_header(self):
        context = FContext(self.correlation_id)
        self.assertEqual(context, context.set_request_header("foo", "bar"))
        self.assertEqual("bar", context.get_request_header("foo"))
        self.assertEqual(self.correlation_id,
                         context.get_request_header("_cid"))

    def test_response_header(self):
        context = FContext(self.correlation_id)
        self.assertEqual(context, context.set_response_header("foo", "bar"))
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

    def test_request_header_put_allows_string_unicode(self):
        context = FContext(self.correlation_id)
        self.assertRaises(TypeError, context.set_request_header, 1, "foo")
        self.assertRaises(TypeError, context.set_request_header, "foo", 3)
        context.set_request_header(u'foo', u'bar')

    def test_response_header_put_allows_string_unicode(self):
        context = FContext(self.correlation_id)
        self.assertRaises(TypeError, context.set_response_header, 1, "foo")
        self.assertRaises(TypeError, context.set_response_header, "foo", 3)
        context.set_request_header(u'foo', u'bar')

    def test_cant_set_cid_public_method(self):
        context = FContext(self.correlation_id)
        context.set_request_header("_cid", "foo")
        self.assertEqual(context.correlation_id, self.correlation_id)

    def test_cant_set_opid_public_method(self):
        context = FContext(self.correlation_id)
        context.set_request_header("_opid", "foo")
        self.assertNotEqual(context.get_request_header("_opid"), "foo")
