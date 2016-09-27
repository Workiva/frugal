import mock

from tornado.testing import gen_test, AsyncTestCase

from frugal.exceptions import FException
from frugal.tornado.transport import FTornadoTransport


class TestFTornadoTranpsort(AsyncTestCase):

    def setUp(self):
        self.transport = FTornadoTransport()

        super(TestFTornadoTranpsort, self).setUp()

    def test_default_values(self):
        self.assertEquals(1024*1024, self.transport._max_message_size)
        self.assertEquals(0, len(self.transport._wbuf.getvalue()))

    @gen_test
    def test_set_registry(self):
        registry = mock.Mock()

        yield self.transport.set_registry(registry)

        self.assertEquals(registry, self.transport._registry)

    @gen_test
    def test_set_registry_none_raises_value_error(self):
        with self.assertRaises(ValueError) as cm:
            yield self.transport.set_registry(None)

        self.assertEquals("registry cannot be null.", cm.exception.message)

    @gen_test
    def test_set_registry_if_already_set_raises_standard_error(self):
        yield self.transport.set_registry(mock.Mock())
        with self.assertRaises(FException) as cm:
            yield self.transport.set_registry(mock.Mock())

        self.assertEquals("registry already set.", cm.exception.message)

    @gen_test
    def test_register_raises_standard_error_if_registry_not_set(self):
        with self.assertRaises(FException) as cm:
            yield self.transport.register("", "")

        self.assertEquals("registry cannot be null.", cm.exception.message)

    @gen_test
    def test_register_calls_registry_register(self):
        mock_registry = mock.Mock()
        self.transport.set_registry(mock_registry)

        yield self.transport.register("ctx", "callback")

        mock_registry.register.assert_called_with("ctx", "callback")

    @gen_test
    def test_unregister_calls_registry_unregister(self):
        mock_registry = mock.Mock()
        self.transport.set_registry(mock_registry)

        yield self.transport.unregister("ctx")

        mock_registry.unregister.assert_called_with("ctx")

    def test_is_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            self.transport.isOpen()

        self.assertEquals("You must override this.", cm.exception.message)

    @gen_test
    def test_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            yield self.transport.open()

        self.assertEquals("You must override this.", cm.exception.message)

    @gen_test
    def test_close_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            yield self.transport.close()

        self.assertEquals("You must override this.", cm.exception.message)

    def test_read_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            self.transport.read(4)

        self.assertEquals("Don't call this.", cm.exception.message)

    def test_write_adds_to_write_buffer(self):
        self.transport.write('asdf')

        self.assertEquals('asdf', self.transport._wbuf.getvalue())

    @gen_test
    def test_flush_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            yield self.transport.close()

        self.assertEquals("You must override this.", cm.exception.message)

    def test_reset_write_buffer_clears_write_buffer(self):
        self.transport.write('asdf')

        self.transport.reset_write_buffer()

        self.assertEquals('', self.transport._wbuf.getvalue())

    def test_execute_frame_calls_registry_execute_without_frame_size(self):
        frame = 'asdfg'

        mock_registry = mock.Mock()
        self.transport.set_registry(mock_registry)

        self.transport.execute_frame(frame)

        mock_registry.execute.assert_called_with('g')




