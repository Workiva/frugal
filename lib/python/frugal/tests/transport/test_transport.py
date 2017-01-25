import unittest

from frugal.transport import FTransport


class TestFTransportBase(unittest.TestCase):

    def setUp(self):
        self.transport = FTransport()

        super(TestFTransportBase, self).setUp()

    def test_is_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            self.transport.is_open()

    def test_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            self.transport.open()

    def test_close_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            self.transport.close()

    def test_set_monitor_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            self.transport.set_monitor(None)

    def test_oneway_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            self.transport.oneway(None, [])

    def test_request_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            self.transport.request(None, [])

    def test_get_request_size_limit(self):
        size = 1024
        transport = FTransport(request_size_limit=size)
        self.assertEqual(size, transport.get_request_size_limit())

