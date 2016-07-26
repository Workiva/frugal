import mock

from tornado.testing import gen_test, AsyncTestCase

from frugal.tornado.transport import FTornadoTransport


class TestFTornadoTransport(AsyncTestCase):
    def setUp(self):
        super(TestFTornadoTransport, self).setUp()

        self.request_capacity = 100
        self.transport = FTornadoTransport(
            max_message_size=self.request_capacity)

    @gen_test
    def test_set_registry_two_times(self):
        with self.assertRaises(ValueError):
            self.transport.set_registry(None)

        registry = mock.Mock()
        self.transport.set_registry(registry)
        self.transport.set_registry(registry)

    @gen_test
    def test_register_unregister(self):
        context = mock.Mock()
        callback = mock.Mock()

        with self.assertRaises(StandardError):
            self.transport.register(context, callback)

        with self.assertRaises(StandardError):
            self.transport.unregister(context)

        registry = mock.Mock()
        self.transport.set_registry(registry)

        self.transport.register(context, callback)
        registry.register.assert_called_once_with(context, callback)

        self.transport.unregister(context)
        registry.unregister.assert_called_once_with(context)

