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

import unittest

from frugal.context import FContext, _DEFAULT_TIMEOUT, _OPID_HEADER


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
        context.set_request_header("_opid", "12345")
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

    def test_copy(self):
        ctx = FContext()
        ctx.set_request_header('foo', 'bar')
        copied = ctx.copy()
        ctxHeaders = ctx.get_request_headers()
        copiedHeaders = copied.get_request_headers()

        # Should not have the same opid
        self.assertNotEqual(ctxHeaders[_OPID_HEADER], copiedHeaders[_OPID_HEADER])

        # Everything else should be the same
        del ctxHeaders[_OPID_HEADER]
        del copiedHeaders[_OPID_HEADER]
        self.assertEqual(ctxHeaders, copiedHeaders)

        # Modifying the originals headers shouldn't affect the clone
        ctx.set_request_header('baz', 'qux')
        self.assertIsNone(copied.get_request_header('baz'))

