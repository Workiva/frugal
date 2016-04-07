import unittest

from frugal.context import FContext
from frugal.util.headers import Writer, Reader


class TestHeaders(unittest.TestCase):

    def setUp(self):
        self.writer = Writer()
        self.reader = Reader(self.writer)

    def test_write_header_given_fcontext(self):
        ctx = FContext("corrId")
        expected = bytearray(
            b'\x00\x00\x00\x00\x12\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId')

        buff = self.writer.write_headers_to_buffer(ctx.get_request_headers())

        self.assertEquals(expected, buff)

    def test_read_headers_given_fcontext(self):
        ctx = FContext("corrId")

        b = self.writer.write_headers_to_buffer(ctx.get_request_headers())

        self.writer.write_bytes(b)

        expected_ctx = self.reader.read_request_headers()

        self.assertEquals(expected_ctx, {'_cid': 'corrId'})
