# -*- coding: utf-8 -*-

import unittest
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
        # Manually set the op id to avoid changes to global op id
        ctx._set_op_id(0)
        expected = bytearray(b'\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00'
                             b'\x00\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06'
                             b'corrId\x00\x00\x00\x08_timeout\x00\x00\x00'
                             b'\x044000')
        buff = self.headers._write_to_bytearray(ctx.get_request_headers())

        self.assertEqual(len(expected), len(buff))

    def test_read_throws_bad_version(self):
        buff = bytearray(b'\x01\x00\x00\x00\x00')

        with self.assertRaises(FProtocolException) as cm:
            self.headers._read(BytesIO(buff))

        self.assertEqual(FProtocolException.BAD_VERSION, cm.exception.type)
        self.assertEqual("Wrong Frugal version. Found 1, wanted 0.",
                         str(cm.exception))

    def test_read(self):
        buff = bytearray(b'\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x00'
                         b'\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId')

        headers = self.headers._read(BytesIO(buff))

        self.assertEqual("0", headers["_opid"])
        self.assertEqual("corrId", headers["_cid"])

    def test_write_read(self):
        context = FContext("corrId")
        context.set_request_header("foo", "bar")

        expected = context.get_request_headers()

        buff = self.headers._write_to_bytearray(expected)

        actual = self.headers._read(BytesIO(buff))

        self.assertEqual(expected["_opid"], actual["_opid"])
        self.assertEqual(expected["_cid"], actual["_cid"])
        self.assertEqual(expected["foo"], actual["foo"])

    def test_decode_from_frame_throws_fprotocol_exception_frame_too_short(self):
        frame = bytearray(b'\x00')

        with self.assertRaises(FProtocolException) as cm:
            self.headers.decode_from_frame(frame)

        self.assertEqual(FProtocolException.INVALID_DATA, cm.exception.type)
        self.assertEqual("Invalid frame size: 1", str(cm.exception))

    def test_decode_from_frame_throws_bad_version(self):
        frame = bytearray(b'\x01\x00\x00\x00\x00')

        with self.assertRaises(FProtocolException) as cm:
            self.headers.decode_from_frame(frame)

        self.assertEqual(FProtocolException.BAD_VERSION, cm.exception.type)
        self.assertEqual("Wrong Frugal version. Found 1, wanted 0.",
                         str(cm.exception))

    def test_decode_from_frame_reads_pairs(self):
        buff = bytearray(b'\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x00'
                         b'\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId')

        headers = self.headers.decode_from_frame(buff)

        self.assertEqual("0", headers["_opid"])
        self.assertEqual("corrId", headers["_cid"])

    def test_read_pairs(self):
        buff = bytearray(b'\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x00'
                         b'\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId')
        size = unpack_from('!I', buff[1:5])[0]

        headers = self.headers._read_pairs(buff, 5, size + 5)

        self.assertEqual("0", headers["_opid"])
        self.assertEqual("corrId", headers["_cid"])

    def test_read_pars_bad_key_throws_error(self):
        buff = bytearray(b'\x00\x00\x00\x00 \x00\x00\x00\x20_opid\x00\x00\x00'
                         b'\x010\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId')
        size = unpack_from('!I', buff[1:5])[0]

        with self.assertRaises(FProtocolException) as cm:
            self.headers._read_pairs(buff, 5, size + 5)

        self.assertEqual(FProtocolException.INVALID_DATA, cm.exception.type)
        self.assertEqual("invalid protocol header name size: 32",
                         str(cm.exception))

    def test_read_pars_bad_value_throws(self):
        buff = bytearray(b'\x00\x00\x00\x00 \x00\x00\x00\x05_opid\x00\x00\x01'
                         b'\x000\x00\x00\x00\x04_cid\x00\x00\x00\x06corrId')
        size = unpack_from('!I', buff[1:5])[0]
        with self.assertRaises(FProtocolException) as cm:
            self.headers._read_pairs(buff, 5, size + 5)

        self.assertEqual(FProtocolException.INVALID_DATA, cm.exception.type)
        self.assertEqual("invalid protocol header value size: 256",
                         str(cm.exception))

    def test_encode_decode_utf8(self):
        headers = {
            u'Đ¥ÑØ': u'δάüΓ',
            u'good\u00F1ight': u'moo\u00F1',
        }
        encoded_headers = _Headers._write_to_bytearray(headers)
        decoded_headers = _Headers.decode_from_frame(encoded_headers)
        self.assertEqual(headers, decoded_headers)
