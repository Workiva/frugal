import unittest

from frugal.exceptions import FMessageSizeException
from frugal.transport.f_bounded_memory_buffer import FBoundedMemoryBuffer


class TestFBoundedMemoryBuffer(unittest.TestCase):

    def setUp(self):
        super(TestFBoundedMemoryBuffer, self).setUp()

        self.buffer = FBoundedMemoryBuffer(10)

    def test_write(self):
        self.buffer.write(b'foo')
        self.assertSequenceEqual(b'foo', self.buffer.getvalue())

    def test_write_size_exception(self):
        self.assertEqual(0, len(self.buffer))
        self.buffer.write(bytearray(10))
        self.assertEqual(10, len(self.buffer))
        with self.assertRaises(FMessageSizeException):
            self.buffer.write(bytearray(1))
        self.assertEqual(0, len(self.buffer))

    def test_len(self):
        self.buffer.write(b'12345')
        self.assertEqual(len(self.buffer), 5)
