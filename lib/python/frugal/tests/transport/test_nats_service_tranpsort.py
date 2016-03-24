import mock

from tornado.testing import gen_test, AsyncTestCase

from frugal.transport.nats_service_transport import TNatsServiceTransport


class TestTNatsServiceTransport(AsyncTestCase):

    def setUp(self):
        self.subject = "foo"
        self.timeout = 20000
        self.max_missed_heartbeats = 3
        super(TestTNatsServiceTransport, self).setUp()

    @gen_test
    def test_init(self):
        mock_nats_client = mock.Mock()

        transport = TNatsServiceTransport(
            mock_nats_client,
            self.subject,
            self.timeout,
            self.max_missed_heartbeats
        )

        self.assertEqual(self.subject, transport._connection_subject)
        self.assertEqual(self.timeout, transport._connection_timeout)

        self.assertFalse(transport._is_open)

    @gen_test
    def test_open(self):
        mock_nats_client = mock.Mock()

        transport = TNatsServiceTransport(
            mock_nats_client,
            self.subject,
            self.timeout,
            self.max_missed_heartbeats
        )

        yield transport.open()

        self.assertEqual(True, transport._is_open())
