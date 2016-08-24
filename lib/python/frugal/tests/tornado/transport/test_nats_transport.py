import mock
import struct

from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase
from thrift.transport.TTransport import TTransportException

from frugal.tornado.transport import FNatsTransport


class TestFNatsTransport(AsyncTestCase):

    def setUp(self):
        self.mock_nats_client = mock.Mock()
        self.subject = "foo"
        self.inbox = "new_inbox"
        super(TestFNatsTransport, self).setUp()

        self.transport = FNatsTransport(self.mock_nats_client,
                                        self.subject,
                                        self.inbox)

    @mock.patch('frugal.tornado.transport.nats_transport.new_inbox')
    def test_init(self, mock_new_inbox):
        self.assertEquals(self.mock_nats_client, self.transport._nats_client)
        self.assertEquals(self.subject, self.transport._subject)
        self.assertEquals(self.inbox, self.transport._inbox)

        mock_new_inbox.return_value = "asdf"

        transport = FNatsTransport(self.mock_nats_client, self.subject)

        mock_new_inbox.assert_called_with()
        self.assertEquals("asdf", transport._inbox)

    @gen_test
    def test_open_throws_nats_not_connected_exception(self):
        self.mock_nats_client.is_connected.return_value = False

        with self.assertRaises(TTransportException) as cm:
            yield self.transport.open()

        self.assertEqual(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEqual("NATS not connected.", cm.exception.message)

    @gen_test
    def test_open_throws_transport_already_open_exception(self):
        self.mock_nats_client.is_connected.return_value = True
        self.transport._is_open = True

        with self.assertRaises(TTransportException) as cm:
            yield self.transport.open()

        self.assertEqual(TTransportException.ALREADY_OPEN, cm.exception.type)
        self.assertEqual("NATS transport already open", cm.exception.message)

    @gen_test
    def test_open_subscribes_to_new_inbox(self):
        f = concurrent.Future()
        f.set_result(1)
        self.mock_nats_client.subscribe.return_value = f

        yield self.transport.open()

        self.assertEquals(1, self.transport._sub_id)
        self.mock_nats_client.subscribe.assert_called_with(
            "new_inbox", "", self.transport._on_message_callback)

    @gen_test
    def test_on_message_callback(self):
        registry_mock = mock.Mock()
        self.transport.set_registry(registry_mock)

        data = b'fooobar'
        msg_mock = mock.Mock(data=data)
        self.transport._on_message_callback(msg_mock)
        registry_mock.execute.assert_called_once_with(data[4:])

    @gen_test
    def test_close_calls_unsubscribe_and_sets_is_open_to_false(self):
        self.transport._sub_id = 1
        f = concurrent.Future()
        f.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = f

        yield self.transport.close()

        self.mock_nats_client.unsubscribe.assert_called_with(
            self.transport._sub_id)

        self.assertFalse(self.transport._is_open)

    @gen_test
    def test_close_with_no_sub_id_returns_early(self):
        self.transport._sub_id = 1
        f = concurrent.Future()
        f.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = f
        yield self.transport.close()
        self.mock_nats_client.unsubscribe.assert_not_called()

    def test_read_throws_exception(self):
        with self.assertRaises(NotImplementedError):
            self.transport.read(2)

    def test_write_adds_to_write_buffer(self):
        b = bytearray('writetest')
        self.mock_nats_client.is_connected.return_value = True
        self.transport._is_open = True

        self.transport.write(b)

        self.assertEquals(b, self.transport._wbuf.getvalue())

    @gen_test
    def test_flush_not_open_raises_exception(self):
        with self.assertRaises(TTransportException) as cm:
            yield self.transport.flush()

        self.assertEquals(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEquals("Nats not connected!", cm.exception.message)

    @gen_test
    def test_flush_publishes_request_to_inbox(self):
        self.mock_nats_client.is_connected.return_value = True
        self.transport._is_open = True

        b = bytearray('test')
        self.transport._wbuf.write(b)
        frame_length = struct.pack('!I', len(b))

        f = concurrent.Future()
        f.set_result("")
        self.mock_nats_client.publish_request.return_value = f

        yield self.transport.flush()

        self.mock_nats_client.publish_request.assert_called_with(
            self.subject,
            self.inbox,
            frame_length + b
        )
