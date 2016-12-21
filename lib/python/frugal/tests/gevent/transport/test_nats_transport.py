import mock
import struct
import unittest

from thrift.transport.TTransport import TTransportException

from frugal.gevent.transport import FNatsTransport
from gnats.client.client import Subscription


class TestFNatsTransport(unittest.TestCase):

    def setUp(self):
        self.mock_nats_client = mock.Mock()
        self.subject = "foo"
        self.inbox = "new_inbox"
        super(TestFNatsTransport, self).setUp()

        self.transport = FNatsTransport(self.mock_nats_client,
                                        self.subject,
                                        self.inbox)

    def test_is_open_returns_true_when_nats_connected(self):
        self.transport._is_open = True
        self.mock_nats_client.is_connected.return_value = True

        self.assertTrue(self.transport.is_open())

    def test_is_open_returns_false_when_nats_not_connected(self):
        self.mock_nats_client.is_connected.return_value = True

        self.assertFalse(self.transport.is_open())

    @mock.patch('frugal.gevent.transport.nats_transport.new_inbox')
    def test_init(self, mock_new_inbox):
        self.assertEquals(self.mock_nats_client, self.transport._nats_client)
        self.assertEquals(self.subject, self.transport._subject)
        self.assertEquals(self.inbox, self.transport._inbox)

        mock_new_inbox.return_value = "asdf"

        transport = FNatsTransport(self.mock_nats_client, self.subject)

        mock_new_inbox.assert_called_with()
        self.assertEquals("asdf", transport._inbox)

    def test_open_throws_nats_not_connected_exception(self):
        self.mock_nats_client.is_connected = False

        with self.assertRaises(TTransportException) as cm:
            self.transport.open()

        self.assertEqual(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEqual("NATS not connected.", cm.exception.message)

    def test_open_throws_transport_already_open_exception(self):
        self.mock_nats_client.is_connected = True
        self.transport._is_open = True

        with self.assertRaises(TTransportException) as cm:
            self.transport.open()

        self.assertEqual(TTransportException.ALREADY_OPEN, cm.exception.type)
        self.assertEqual("NATS transport already open.", cm.exception.message)

    def test_open_subscribes_to_new_inbox(self):

        sub = Subscription(sid=123, subject='foo')
        self.mock_nats_client.subscribe.return_value = sub

        self.transport.open()

        self.assertEquals(sub, self.transport._sub)
        self.mock_nats_client.subscribe.assert_called_with(
            "new_inbox", cb=self.transport._on_message_callback)

    def test_on_message_callback(self):
        registry_mock = mock.Mock()
        self.transport._registry = registry_mock

        data = b'fooobar'
        msg_mock = mock.Mock(data=data)
        self.transport._on_message_callback(msg_mock)
        registry_mock.execute.assert_called_once_with(data[4:])

    def test_close_calls_unsubscribe_and_sets_is_open_to_false(self):
        self.transport._sub = Subscription(sid=123, subject='foo')

        self.mock_nats_client.unsubscribe.return_value = None
        self.mock_nats_client.flush.return_value = None

        self.transport.close()

        self.mock_nats_client.unsubscribe.assert_called_with(
            self.transport._sub)
        self.mock_nats_client.flush.assert_called_with()

        self.assertFalse(self.transport._is_open)

    def test_close_with_no_sub_id_returns_early(self):
        self.transport._sub_id = 1

        self.mock_nats_client.unsubscribe.return_value = None
        self.mock_nats_client.flush.return_value = None

        self.transport.close()

        self.mock_nats_client.unsubscribe.assert_not_called()

    def test_send_not_open_raises_exception(self):
        with self.assertRaises(TTransportException) as cm:
            self.transport.send([])

        self.assertEquals(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEquals("NATS not connected.", cm.exception.message)

    def test_send_publishes_request_to_inbox(self):
        self.mock_nats_client.is_connected.return_value = True
        self.transport._is_open = True

        data = bytearray('test')
        frame_length = struct.pack('!I', len(data))

        self.mock_nats_client.publish.return_value = None
        self.mock_nats_client.flush.return_value = None

        self.transport.send(frame_length + data)

        self.mock_nats_client.publish.assert_called_with(
            self.subject,
            frame_length + data,
            reply=self.inbox
        )

        # self.mock_nats_client.flush.assert_called_with()
