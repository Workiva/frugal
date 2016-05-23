import mock

from tornado import concurrent, ioloop
from tornado.testing import gen_test, AsyncTestCase

from frugal.server import FNatsTornadoServer
from frugal.server.nats_server import _Client


class TestFNatsTornadoServer(AsyncTestCase):

    def setUp(self):
        super(TestFNatsTornadoServer, self).setUp()

        self.subject = "foo"
        self.mock_nats_client = mock.Mock()
        self.mock_processor_factory = mock.Mock()
        self.mock_transport_factory = mock.Mock()
        self.mock_prot_factory = mock.Mock()

        self.max_missed_heartbeats = 2
        self.heartbeat_interval = 1000

        self.server = FNatsTornadoServer(
            self.mock_nats_client,
            self.subject,
            self.heartbeat_interval,
            self.max_missed_heartbeats,
            self.mock_processor_factory,
            self.mock_transport_factory,
            self.mock_prot_factory
        )

        self.mock_transport = mock.Mock()
        self.client = _Client(self.mock_nats_client,
                              self.mock_transport,
                              "heartbeat",
                              self.heartbeat_interval,
                              self.max_missed_heartbeats)

    @gen_test
    def test_serve(self):
        f = concurrent.Future()
        f.set_result(123)
        self.mock_nats_client.subscribe.return_value = f

        yield self.server.serve()

        self.assertEquals(123, self.server._sid)

    @gen_test
    def test_stop(self):
        mock_heartbeater = mock.Mock()
        mock_heartbeater.is_running.return_value = True
        self.server._heartbeater = mock_heartbeater

        yield self.server.stop()

        mock_heartbeater.stop.assert_called_with()

    def test_set_and_get_high_watermark(self):
        self.server.set_high_watermark(1234)

        self.assertEquals(1234, self.server.get_high_watermark())

    @mock.patch('frugal.server.nats_server.TNatsServiceTransport')
    @gen_test
    def test_accept(self, mock_server_constructor):
        mock_server_transport = mock.Mock()
        mock_processor = mock.Mock()
        mock_protocol = mock.Mock()

        mock_server_constructor.Server.return_value = mock_server_transport

        f = concurrent.Future()
        f.set_result(None)
        self.mock_transport.open.return_value = f

        self.mock_transport_factory.get_transport.return_value = self.mock_transport

        client = yield self.server._accept("listen_to", "reply_to", "heartbeat")

        self.assertEquals(mock_server_transport, client)

        mock_server_constructor.Server.assert_called_with(self.mock_nats_client,
                                                          "listen_to",
                                                          "reply_to")
        self.mock_transport_factory.get_transport.assert_called_with(client)
        self.mock_processor_factory.get_processor.assert_called_with(self.mock_transport)
        self.mock_prot_factory.get_protocol.assert_called_with(self.mock_transport)
        self.mock_transport.open.assert_called_with()

    @mock.patch('frugal.server.nats_server.new_inbox')
    def test_new_frugal_inbox(self, mock_new_inbox):
        mock_new_inbox.return_value = "new_inbox"
        prefix = "frugal._INBOX.d138b9369fa35386624d6ad97"

        result = self.server._new_frugal_inbox(prefix)

        self.assertEquals("frugal._INBOX.new_inbox", result)

    @gen_test
    def test_client_kill(self):
        f = concurrent.Future()
        f.set_result(None)
        f2 = concurrent.Future()
        f2.set_result(None)
        self.mock_transport.close.return_value = f
        self.mock_nats_client.auto_unsubscribe.return_value = f2
        self.client._hb_sub_id = 123
        self.client._heartbeat_timer = ioloop.PeriodicCallback(None, 1)

        yield self.client.kill()

        self.mock_nats_client.auto_unsubscribe.assert_called_with(
            self.client._hb_sub_id,
            ""
        )
        self.mock_transport.close.assert_called_with()

    @gen_test
    def test_client_start(self):
        f = concurrent.Future()
        f.set_result(123)
        self.mock_nats_client.subscribe.return_value = f

        yield self.client.start()

        self.mock_nats_client.subscribe.assert_called_with(
            "heartbeat",
            "",
            self.client._receive_heartbeat
        )

    def test_client_receive_heartbeat(self):
        self.client._missed_heartbeats = 1

        self.client._receive_heartbeat(
            "dont care what this is, heartbeats empty")

        self.assertEquals(0, self.client._missed_heartbeats)

    @gen_test
    def test_client_missed_heartbeat_increments_count(self):

        yield self.client._missed_heartbeat("still dont care")

        self.assertEquals(1, self.client._missed_heartbeats)

    @gen_test
    def test_client_missed_heartbeat_greater_than_max_calls_kill(self):
        f = concurrent.Future()
        f.set_result(None)
        self.mock_transport.close.return_value = f
        f2 = concurrent.Future()
        f2.set_result(123)
        self.mock_nats_client.auto_unsubscribe.return_value = f2

        self.client._missed_heartbeats = 3
        self.client._hb_sub_id = 123
        self.client._heartbeat_timer = ioloop.PeriodicCallback(None, 1)

        self.client._missed_heartbeat("random words: fliggy floo")

        self.mock_nats_client.auto_unsubscribe.assert_called_with(123, "")
        self.mock_transport.close.assert_called_with()


