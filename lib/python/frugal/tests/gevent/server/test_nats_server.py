from io import BytesIO
import mock
import struct
import unittest

from gnats.client.client import Subscription

from frugal.gevent.server import FNatsServer


class testFNatsGeventServer(unittest.TestCase):

    def setUp(self):
        super(testFNatsGeventServer, self).setUp()

        self.subject = "foo"
        self.mock_nats_client = mock.Mock()
        self.mock_processor = mock.Mock()
        self.mock_transport_factory = mock.Mock()
        self.mock_prot_factory = mock.Mock()

        self.server = FNatsServer(
            self.mock_nats_client,
            self.subject,
            self.mock_processor,
            self.mock_prot_factory
        )
        self.server._iprot_factory = mock.Mock()
        self.server._oprot_factory = mock.Mock()

    def test_serve(self):
        sub = Subscription(sid=123, subject='foo')
        self.mock_nats_client.subscribe.return_value = sub

        with mock.patch("frugal.gevent.server.nats_server.Event") as m:
            self.server.blocking_event = m
            self.server.serve()

        self.assertEquals([sub], self.server._subs)

    def test_stop(self):
        self.server._sub_ids = [123]
        self.mock_nats_client.unsubscribe.return_value = None

        self.server.stop()

        self.mock_nats_client.unsubscribe.assert_called_with(123)

    def test_on_message_callback_no_reply_returns_early(self):
        data = b'asdf'
        frame_size = struct.pack('!I', len(data))
        msg = TestMsg(subject='test', reply='', data=frame_size + data)

        self.server._on_message_callback(msg)

        assert not self.server._iprot_factory.get_protocol.called
        assert not self.server._oprot_factory.get_protocol.called
        assert not self.server._processor.process.called

    @mock.patch('frugal.server.server._NATS_MAX_MESSAGE_SIZE', 6)
    def test_on_message_callback_bad_framesize_returns_early(self):
        data = b'asdf'
        frame_size = struct.pack('!I', len(data))
        msg = TestMsg(subject='test', reply='reply', data=frame_size + data)

        self.server._on_message_callback(msg)

        assert not self.server._iprot_factory.get_protocol.called
        assert not self.server._oprot_factory.get_protocol.called
        assert not self.server._processor.process.called

    def test_on_message_callback_calls_process(self):
        iprot = BytesIO()
        oprot = BytesIO()
        self.server._iprot_factory.get_protocol.return_value = iprot
        self.server._oprot_factory.get_protocol.return_value = oprot

        data = b'asdf'
        frame_size = struct.pack('!I', len(data))

        msg = TestMsg(subject="foo", reply="inbox", data=frame_size + data)
        self.mock_nats_client.publish.return_value = None


        self.mock_processor.process.return_value = None

        self.server._on_message_callback(msg)

        self.server._processor.process.assert_called_with(iprot, oprot)


class TestMsg(object):
    def __init__(self, subject='', reply='', data=b'', sid=0,):
        self.subject = subject
        self.reply = reply
        self.data = data
        self.sid = sid
