from io import BytesIO
import mock
import struct
from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase

from frugal.tornado.server import FNatsTornadoServer


class TestFNatsTornadoServer(AsyncTestCase):

    def setUp(self):
        super(TestFNatsTornadoServer, self).setUp()

        self.subject = "foo"
        self.mock_nats_client = mock.Mock()
        self.mock_processor = mock.Mock()
        self.mock_transport_factory = mock.Mock()
        self.mock_prot_factory = mock.Mock()

        self.server = FNatsTornadoServer(
            self.mock_nats_client,
            self.subject,
            self.mock_processor,
            self.mock_prot_factory
        )
        self.server._iprot_factory = mock.Mock()
        self.server._oprot_factory = mock.Mock()

    @gen_test
    def test_serve(self):
        f = concurrent.Future()
        f.set_result(123)
        self.mock_nats_client.subscribe.return_value = f

        yield self.server.serve()

        self.assertEquals([123], self.server._sids)

    @gen_test
    def test_stop(self):
        self.server._sids = [123]
        f = concurrent.Future()
        f.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = f

        yield self.server.stop()

        self.mock_nats_client.unsubscribe.assert_called_with(123)

    @gen_test
    def test_on_message_callback_no_reply_returns_early(self):
        data = b'asdf'
        frame_size = struct.pack('!I', len(data))
        msg = TestMsg(subject='test', reply='', data=frame_size + data)

        yield self.server._on_message_callback(msg)

        assert not self.server._iprot_factory.get_protocol.called
        assert not self.server._oprot_factory.get_protocol.called
        assert not self.server._processor.process.called

    @mock.patch('frugal.tornado.server.nats_server.MAX_MESSAGE_SIZE', 6)
    @gen_test
    def test_on_message_callback_bad_framesize_returns_early(self):
        data = b'asdf'
        frame_size = struct.pack('!I', len(data))
        msg = TestMsg(subject='test', reply='reply', data=frame_size + data)

        yield self.server._on_message_callback(msg)

        assert not self.server._iprot_factory.get_protocol.called
        assert not self.server._oprot_factory.get_protocol.called
        assert not self.server._processor.process.called

    @gen_test
    def test_on_message_callback_calls_process(self):
        iprot = BytesIO()
        oprot = BytesIO()
        self.server._iprot_factory.get_protocol.return_value = iprot
        self.server._oprot_factory.get_protocol.return_value = oprot

        data = b'asdf'
        frame_size = struct.pack('!I', len(data))

        msg = TestMsg(subject="foo", reply="inbox", data=frame_size + data)

        yield self.server._on_message_callback(msg)

        self.server._processor.process.assert_called_with(iprot, oprot)


class TestMsg(object):
    def __init__(self, subject='', reply='', data=b'', sid=0,):
        self.subject = subject
        self.reply = reply
        self.data = data
        self.sid = sid
