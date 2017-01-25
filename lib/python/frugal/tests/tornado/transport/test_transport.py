from tornado.testing import gen_test, AsyncTestCase

from frugal.tornado.transport import FTransportBase


class TestFTornadoTransport(AsyncTestCase):

    def setUp(self):
        self.transport = FTransportBase()

        super(TestFTornadoTransport, self).setUp()

    def test_is_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            self.transport.is_open()

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

    @gen_test
    def test_oneway_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            yield self.transport.oneway(None, [])

        self.assertEquals("You must override this.", cm.exception.message)

    @gen_test
    def test_request_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            yield self.transport.request(None, [])

        self.assertEquals("You must override this.", cm.exception.message)

