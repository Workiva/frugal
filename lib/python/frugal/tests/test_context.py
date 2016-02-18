import unittest
from mock import patch

from frugal.context import FContext


class TestContext(unittest.TestCase):

    correlation_id = "fooid"

    def test_correlation_id(self):
        context = FContext("fooid")
        self.assertEqual("fooid", context.get_correlation_id())

    @patch('uuid.uuid4')
    def test_empty_correlation_id(self, mock_uuid):
        mock_uuid.return_value = "12345"

        context = FContext()
        self.assertEqual("12345", context.get_correlation_id())

    def test_op_id(self):
        context = FContext(self.correlation_id)
        context.put_request_header("_opid", "12345")
        self.assertEqual(self.correlation_id, context.get_correlation_id())
        self.assertEqual("12345", context.get_request_header("_opid"))

    def test_request_header(self):
        context = FContext(self.correlation_id)
        context.put_request_header("foo", "bar")
        self.assertEqual("bar", context.get_request_header("foo"))
        self.assertEqual(self.correlation_id,
                         context.get_request_header("_cid"))

    def test_response_header(self):
        context = FContext(self.correlation_id)
        context.put_response_header("foo", "bar")
        self.assertEqual("bar", context.get_response_header("foo"))
        self.assertEqual(self.correlation_id,
                         context.get_request_header("_cid"))

    def test_request_headers(self):
        context = FContext(self.correlation_id)
        context.put_request_header("foo", "bar")
        headers = context.get_request_headers()
        self.assertEqual("bar", headers['foo'])

    def test_response_headers(self):
        context = FContext(self.correlation_id)
        context.put_response_header("foo", "bar")
        headers = context.get_response_headers()
        self.assertEqual("bar", headers['foo'])

    def test_request_header_put_only_allows_string(self):
        context = FContext(self.correlation_id)
        self.assertRaises(TypeError, context.put_request_header, 1, "foo")
        self.assertRaises(TypeError, context.put_request_header, "foo", 3)

    def test_response_header_put_only_allows_string(self):
        context = FContext(self.correlation_id)
        self.assertRaises(TypeError, context.put_response_header, 1, "foo")
        self.assertRaises(TypeError, context.put_response_header, "foo", 3)
