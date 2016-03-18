import unittest
import mock

from frugal.context import FContext
from frugal.transport.transport import FMuxTransport
from frugal.registry import FRegistry


class TestFMuxTransport(unittest.TestCase):

    def setUp(self):
        self.mock_thrit_transport = mock.Mock()
        self.transport = FMuxTransport(self.mock_thrit_transport)

    def test_set_registry_with_none_throws_error(self):
        with self.assertRaises(StandardError):
            self.transport.set_registry(None)

    def test_set_registry_sets_registry(self):
        mock_registry = mock.Mock()

        self.transport.set_registry(mock_registry)

        self.assertEqual(mock_registry, self.transport._registry)

    def test_register(self):
        mock_registry = mock.Mock()
        self.transport.set_registry(mock_registry)

        def cb():
            pass

        ctx = FContext()

        self.transport.register(ctx, cb)

        mock_registry.register.assert_called_with(ctx, cb)

    def test_register_none_registry(self):
        def cb():
            pass

        ctx = FContext()

        with self.assertRaises(StandardError):
            self.transport.register(ctx, cb)

    def test_unregister(self):
        mock_registry = mock.Mock()
        self.transport.set_registry(mock_registry)
        ctx = FContext()

        self.transport.unregister(ctx)

        mock_registry.unregister.assert_called_with(ctx)

    def test_unregister_none_registry(self):
        ctx = FContext()

        with self.assertRaises(StandardError):
            self.transport.unregister(ctx)

    def test_is_open_true(self):
        mock_registry = mock.Mock()
        self.transport.set_registry(mock_registry)
        self.mock_thrit_transport.isOpen.return_value = True

        self.assertTrue(self.transport.is_open())

    def test_is_open_false_registry_none(self):
        self.mock_thrit_transport.isOpen.return_value = True

        self.assertFalse(self.transport.is_open())

    def test_is_open_false_transport_not_open(self):
        mock_registry = mock.Mock()
        self.transport.set_registry(mock_registry)
        self.mock_thrit_transport.isOpen.return_value = False

        self.assertFalse(self.transport.is_open())

    def test_open_calls_open_on_transport(self):
        self.transport.open()

        self.mock_thrit_transport.open.assert_called()

    # TODO - figure out how to test .read(); it's calling the TFramedTransport
