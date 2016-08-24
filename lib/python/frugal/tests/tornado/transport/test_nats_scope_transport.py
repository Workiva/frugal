import mock
from io import BytesIO
import struct

from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase
from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FException, FMessageSizeException
from frugal.tornado.transport.nats_scope_transport import FNatsScopeTransport


class TestFNatsScopeTransport(AsyncTestCase):

    def setUp(self):
        super(TestFNatsScopeTransport, self).setUp()

        self.nats_client = mock.Mock()

        self.publisher_transport = FNatsScopeTransport(self.nats_client)
        self.subscriber_transport = FNatsScopeTransport(self.nats_client, "Q")

    def test_lock_topic_sets_topic(self):
        expected = "topic"

        self.publisher_transport.lock_topic(expected)

        self.assertEqual(expected, self.publisher_transport._subject)

    def test_unlock_topic_resets_topic(self):

        self.publisher_transport.lock_topic("topic")
        self.publisher_transport.unlock_topic()

        self.assertEqual("", self.publisher_transport._subject)

    def test_subscriber(self):
        expected = "topic"
        self.subscriber_transport.subscribe(expected)

        self.assertTrue(self.subscriber_transport._pull)
        self.assertEqual(expected, self.subscriber_transport._subject)

    def test_subscriber_cannot_lock_topic(self):
        expected = "topic"
        self.publisher_transport.subscribe(expected)

        with self.assertRaises(FException) as cm:
            self.publisher_transport.lock_topic(expected)

        self.assertEquals("Subscriber cannot lock topic.",
                          cm.exception.message)

    def test_subscriber_cannot_unlock_topic(self):
        expected = "topic"
        self.publisher_transport.subscribe(expected)

        with self.assertRaises(FException) as cm:
            self.publisher_transport.unlock_topic()

        self.assertEquals("Subscriber cannot unlock topic.",
                          cm.exception.message)

    @gen_test
    def test_open_throws_exception_if_nats_not_connected(self):
        mock_conn = mock.Mock()
        mock_conn.is_connected.return_value = False

        self.publisher_transport = FNatsScopeTransport(mock_conn)

        with self.assertRaises(TTransportException) as cm:
            yield self.publisher_transport.open()

        self.assertEquals(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEquals("Nats not connected!", cm.exception.message)

    @gen_test
    def test_open_throws_exception_if_nats_already_open(self):
        self.nats_client.is_connected.return_value = True
        self.publisher_transport._is_open = True

        with self.assertRaises(TTransportException) as cm:
            yield self.publisher_transport.open()

        self.assertEquals(TTransportException.ALREADY_OPEN, cm.exception.type)
        self.assertEquals("Nats is already open!", cm.exception.message)

    @gen_test
    def test_open_when_subscriber_throws_if_empty_subject(self):
        self.nats_client.is_connected.return_value = True
        self.subscriber_transport._pull = True

        with self.assertRaises(TTransportException) as cm:
            yield self.subscriber_transport.open()

        self.assertEquals("Subject cannot be empty.", cm.exception.message)

    @gen_test
    def test_open_when_subscriber_calls_subscribe(self):
        self.nats_client.is_connected.return_value = True
        self.subscriber_transport._pull = True
        self.subscriber_transport._subject = "foo"

        f = concurrent.Future()
        f.set_result(1)
        self.nats_client.subscribe.return_value = f

        yield self.subscriber_transport.open()

        self.nats_client.subscribe.assert_called()
        self.assertTrue(self.subscriber_transport.isOpen())

    @gen_test
    def test_write_throws_if_max_message_size_exceeded(self):
        self.nats_client.is_connected.return_value = True
        self.publisher_transport._is_open = True
        self.publisher_transport._write_buffer = BytesIO()

        buff = bytearray(1024 * 2048)
        with self.assertRaises(FMessageSizeException) as cm:
            yield self.publisher_transport.write(buff)

        self.assertEquals("Message exceeds NATS max message size",
                          cm.exception.message)

    def test_write_writes_to_write_buffer(self):
        self.nats_client.is_connected.return_value = True
        self.publisher_transport._is_open = True
        self.publisher_transport._write_buffer = BytesIO()
        buff = bytearray(b'\x00\x00\x00\x00\x01')

        self.publisher_transport.write(buff)

        self.assertEquals(buff,
                          self.publisher_transport._write_buffer.getvalue())

    @gen_test
    def test_flush_throws_if_transport_not_open(self):
        with self.assertRaises(TTransportException) as cm:
            yield self.publisher_transport.flush()

        self.assertEquals(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEquals("Nats not connected!", cm.exception.message)

    @gen_test
    def test_flush_publishes_to_formatted_subject(self):
        self.nats_client.is_connected.return_value = True
        self.publisher_transport._is_open = True
        self.publisher_transport._subject = "batman"
        self.publisher_transport._write_buffer = BytesIO()
        buff = bytearray(b'\x00\x00\x00\x00\x01')
        size = struct.pack('!I', len(buff))

        f = concurrent.Future()
        f.set_result("")
        self.nats_client.publish.return_value = f

        self.publisher_transport.write(buff)
        yield self.publisher_transport.flush()

        self.nats_client.publish.assert_called_with("frugal.batman",
                                                    "{0}{1}".format(size, buff)
                                                    )

    def test_get_formatted_subject(self):
        self.publisher_transport._subject = "robin"
        self.assertEquals("frugal.robin",
                          self.publisher_transport._get_formatted_subject())

