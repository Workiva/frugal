import unittest
import mock
from io import BytesIO
from struct import unpack_from

from frugal.context import FContext
from frugal.exceptions import FProtocolException
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

        headers = self.headers._read(BytesIO(buff))

        self.assertEquals("0", headers["_opid"])
        self.assertEquals("corrId", headers["_cid"])

    def test_write_read(self):
        context = FContext("corrId")
        context.set_request_header("foo", "bar")

        expected = context.get_request_headers()

        buff = self.headers._write_to_bytearray(expected)

        actual = self.headers._read(BytesIO(buff))

        self.assertEquals(expected["_opid"], actual["_opid"])
        self.assertEquals(expected["_cid"], actual["_cid"])
        self.assertEquals(expected["foo"], actual["foo"])

    def test_decode_from_frame_throws_fprotocol_exception_frame_too_short(self):

        frame = b'\x00'

        try:
            self.headers.decode_from_frame(frame)
            self.fail()
        except FProtocolException as ex:
            self.assertEquals(FProtocolException.INVALID_DATA, ex.type)
            self.assertEquals("Invalid frame size: 1", ex.message)

    def test_decode_from_frame_throws_bad_version(self):

        frame = b'\x00\x00\x00\x00\x01\x00\x00\x00\x00'

        try:
            self.headers.decode_from_frame(frame)
            self.fail()
        except FProtocolException as ex:
            self.assertEquals(FProtocolException.BAD_VERSION, ex.type)
            self.assertEquals("Wrong Frugal version. Found 1, wanted 0.",
                              ex.message)

    def test_decode_from_frame_reads_pairs(self):
        buff = b'\x00\x00\x00\x00\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x00\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId'

        headers = self.headers.decode_from_frame(buff)

        self.assertEquals("0", headers["_opid"])
        self.assertEquals("corrId", headers["_cid"])

    def test_read_pairs(self):
        buff = b'\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x00\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId'
        size = unpack_from('!I', buff[1:5])[0]
        print("size {}".format(size))

        headers = self.headers._read_pairs(buff, 5, size + 5)

        self.assertEquals("0", headers["_opid"])
        self.assertEquals("corrId", headers["_cid"])

    def test_read_pars_bad_key_throws_error(self):
        buff = b'\x00\x00\x00\x00 \x00\x00\x00\x20_opid\x00\x00\x00\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId'
        size = unpack_from('!I', buff[1:5])[0]
        print("size {}".format(size))

        try:
            self.headers._read_pairs(buff, 5, size + 5)
        except FProtocolException as ex:
            self.assertEquals(FProtocolException.INVALID_DATA, ex.type)
            self.assertEquals("invalid protocol header name size: 32", ex.message)

    def test_read_pars_bad_value_throws(self):
        buff = b'\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x01\x000\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId'
        size = unpack_from('!I', buff[1:5])[0]
        print("size {}".format(size))

        try:
            self.headers._read_pairs(buff, 5, size + 5)
        except FProtocolException as ex:
            self.assertEquals(FProtocolException.INVALID_DATA, ex.type)
            self.assertEquals("invalid protocol header value size: 256",
                              ex.message)

