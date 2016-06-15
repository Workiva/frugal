import json
import mock

from tornado import concurrent, ioloop
from tornado.testing import gen_test, AsyncTestCase

from frugal.server import FNatsTornadoServer
from frugal.server.nats_server import _Client

_NATS_PROTOCOL_V0 = 0


class TestFNatsTornadoServer(AsyncTestCase):

    def setUp(self):
        patcher = mock.patch('frugal.server.nats_server.new_inbox')
        self.mock_new_inbox = patcher.start()
        self.addCleanup(patcher.stop)

        super(TestFNatsTornadoServer, self).setUp()
        self.mock_new_inbox.return_value = "new_inbox"

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
            self.max_missed_heartbeats,
            self.mock_processor_factory,
            self.mock_transport_factory,
            self.mock_prot_factory,
            self.heartbeat_interval
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
        # Set heartbeat_interval to 0 so we dont start the PeriodicCallback
        self.server._heartbeat_interval = 0

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

    @gen_test
    def test_remove(self):
        mock_client = mock.Mock()
        f = concurrent.Future()
        f.set_result(None)
        mock_client.kill.return_value = f

        self.server._clients = {"heartbeat": mock_client}

        yield self.server._remove("heartbeat")

        mock_client.kill.assert_called_with()

    @gen_test
    def test_send_heartbeat(self):
        yield self.server._send_heartbeat()

        # Don't publish heartbeats if no clients connected.
        self.mock_nats_client.publish.assert_not_called()

        f = concurrent.Future()
        f.set_result(None)
        self.mock_nats_client.publish.return_value = f
        mock_client = mock.Mock()
        self.server._clients = {"hearbeat": mock_client}

        yield self.server._send_heartbeat()

        self.mock_nats_client.publish.assert_called_with(
            self.server._heartbeat_subject,
            ""
        )

    @mock.patch('frugal.server.nats_server.TNatsServiceTransport')
    @gen_test
    def test_on_message_callback(self, mock_server_constructor):
        mock_server_transport = mock.Mock()

        mock_server_constructor.Server.return_value = mock_server_transport

        f = concurrent.Future()
        f.set_result(None)
        self.mock_transport.open.return_value = f

        self.mock_transport_factory.get_transport.return_value = self.mock_transport
        self.mock_nats_client.publish_request.return_value = f

        msg = TestMsg()

        yield self.server._on_message_callback(msg)

        conn_prot = json.dumps({"version": _NATS_PROTOCOL_V0})

        msg = TestMsg("subject", "reply", conn_prot)

        yield self.server._on_message_callback(msg)

        expected_listen = "new_inbox"
        expected_connect = "{} {} {}".format(self.server._heartbeat_subject,
                                             "new_inbox",
                                             self.server._heartbeat_interval)

        self.mock_nats_client.publish_request.assert_called_with(
            "reply",
            expected_listen,
            expected_connect
        )

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
        self.mock_nats_client.unsubscribe.return_value = f2
        self.client._hb_sub_id = 123
        self.client._heartbeat_timer = ioloop.PeriodicCallback(None, 1)

        yield self.client.kill()

        self.mock_nats_client.unsubscribe.assert_called_with(
            self.client._hb_sub_id
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
        self.mock_nats_client.unsubscribe.return_value = f2

        self.client._missed_heartbeats = 3
        self.client._hb_sub_id = 123
        self.client._heartbeat_timer = ioloop.PeriodicCallback(None, 1)

        self.client._missed_heartbeat("random words: fliggy floo")

        self.mock_nats_client.unsubscribe.assert_called_with(123)
        self.mock_transport.close.assert_called_with()


class TestMsg(object):
    def __init__(self, subject='', reply='', data=b'', sid=0,):
        self.subject = subject
        self.reply   = reply
        self.data    = data
        self.sid     = sid
