import mock

from tornado.concurrent import Future
from tornado.testing import AsyncTestCase, gen_test

from frugal.context import FContext
from frugal.transport import FMuxTornadoTransport


class TestFmuxTornadoTransport(AsyncTestCase):

    def setUp(self):
        self.mock_thrift_transport = mock.Mock()
        self.transport = FMuxTornadoTransport(self.mock_thrift_transport)

        super(TestFmuxTornadoTransport, self).setUp()

    @gen_test
    def test_open(self):
        future = Future()
        future.set_result(None)
        self.mock_thrift_transport.open.return_value = future

        yield self.transport.open()

        self.mock_thrift_transport.open.assert_called_with()

    @gen_test
    def test_close(self):
        f = Future()
        f.set_result(None)
        self.mock_thrift_transport.close.return_value = f

        yield self.transport.close()

        self.mock_thrift_transport.close.assert_called_with()

    def test_is_open_calls_underlying_transport_is_open(self):
        self.mock_thrift_transport.isOpen.return_value = False

        self.assertFalse(self.transport.isOpen())

        self.mock_thrift_transport.isOpen.return_value = True

        self.assertTrue(self.transport.isOpen())

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
