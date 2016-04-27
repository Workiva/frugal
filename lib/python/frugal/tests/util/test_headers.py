import unittest

from frugal.context import FContext
from frugal.util.headers import _Headers


class TestHeaders(unittest.TestCase):

    def setUp(self):
        self.headers = _Headers()

    def test_write_header_given_fcontext(self):
        ctx = FContext("corrId")
        expected = bytearray(b'\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x00\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId')
        buff = self.headers._write_to_bytearray(ctx.get_request_headers())

        self.assertEquals(expected, buff)

    def test_read(self):
        buff = '\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x00\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId'

        headers = self.headers._read(buff)

        self.assertEquals("0", headers["_opid"])
        self.assertEquals("corrId", headers["_cid"])
