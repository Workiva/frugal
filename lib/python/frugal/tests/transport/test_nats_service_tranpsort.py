import mock

from thrift.transport.TTransport import TTransportException
from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase

from frugal.transport.nats_service_transport import TNatsServiceTransport


class TestTNatsServiceTransport(AsyncTestCase):

    def setUp(self):
        self.subject = "foo"
        self.timeout = 20000
        self.max_missed_heartbeats = 3
        super(TestTNatsServiceTransport, self).setUp()

        self.mock_nats_client = mock.Mock()

        self.transport = TNatsServiceTransport(
            self.mock_nats_client,
            self.subject,
            self.timeout,
            self.max_missed_heartbeats
        )

    @gen_test
    def test_init(self):
        self.assertEqual(self.subject, self.transport._connection_subject)
        self.assertEqual(self.timeout, self.transport._connection_timeout)

        self.assertFalse(self.transport._is_open)

    @gen_test
    def test_open_throws_nats_not_connected_exception(self):
        self.mock_nats_client.is_connected.return_value = False

        try:
            yield self.transport.open()
            self.fail()
        except TTransportException as e:
            self.assertEqual(1, e.type)
            self.assertEqual("NATS not connected.", e.message)

    @gen_test
    def test_open_throws_transport_already_open_exception(self):
        self.mock_nats_client.is_connected.return_value = True
        self.transport._is_open = True

        try:
            yield self.transport.open()
            self.fail()
        except TTransportException as e:
            self.assertEqual(2, e.type)
            self.assertEqual("NATS transport already open", e.message)

    @gen_test
    def test_open(self):
        self.mock_nats_client.is_connected.return_value = True

        f = concurrent.Future()
        f.set_result("handshake response 1234")


    @gen_test
    def test_handshake(self):
        f = concurrent.Future()
        f.set_result(1)
        self.mock_nats_client.subscribe.return_value = f

        f2 = concurrent.Future()
        f2.set_result(None)
        self.mock_nats_client.auto_unsubscribe.return_value = f2

        self.mock_nats_client.auto_unsubscribe.return_value = f2
        f3 = concurrent.Future()
        f3.set_result(None)

        self.mock_nats_client.publish_request.return_value = f3

        msg = yield self.transport._handshake()


