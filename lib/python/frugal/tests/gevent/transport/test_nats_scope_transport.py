import mock
from io import BytesIO
import unittest

from thrift.transport.TTransport import TTransportException

from gnats.client.client import Subscription
from gnats.client.errors import ErrConnectionClosed, ErrTimeout, ErrMaxPayload

from frugal.exceptions import FMessageSizeException
from frugal.gevent.transport import FNatsPublisherTransport
from frugal.gevent.transport import FNatsSubscriberTransport


class TestFNatsScopeTransport(unittest.TestCase):

    def setUp(self):
        super(TestFNatsScopeTransport, self).setUp()

        self.nats_client = mock.Mock()

        self.publisher_transport = FNatsPublisherTransport(self.nats_client)
        self.subscriber_transport = FNatsSubscriberTransport(
            self.nats_client, "Q")

    def test_subscriber(self):
        sub = Subscription(sid=123, subject='foo')
        self.nats_client.subscribe.return_value = sub

        topic = 'bar'

        self.subscriber_transport.subscribe(topic, None)
        self.nats_client.subscribe.assert_called_once_with(
            'frugal.bar',
            queue='Q',
            cb=mock.ANY,
        )
        self.assertEqual(self.subscriber_transport._sub, sub)
        self.nats_client.flush.assert_called_once_with()

    def test_open_throws_exception_if_nats_not_connected(self):
        self.nats_client.is_connected = False

        with self.assertRaises(TTransportException) as cm:
            self.publisher_transport.open()

        self.assertEquals(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEquals("Nats not connected!", cm.exception.message)

    def test_open_when_subscriber_calls_subscribe(self):
        self.nats_client.is_connected = True

        sub = Subscription(sid=123, subject='foo')
        self.nats_client.subscribe.return_value = sub

        self.subscriber_transport.subscribe('foo', None)

        self.assertTrue(self.nats_client.subscribe.called)
        self.assertTrue(self.subscriber_transport.is_subscribed())

    def test_publish_throws_if_max_message_size_exceeded(self):
        self.nats_client.is_connected = True
        self.publisher_transport._is_open = True

        buff = bytearray(1024 * 2048)
        with self.assertRaises(FMessageSizeException) as cm:
            self.publisher_transport.publish('foo', buff)

        self.assertEquals("Message exceeds NATS max message size",
                          cm.exception.message)

    def test_publish_throws_if_transport_not_open(self):
        self.nats_client.is_connected = False

        with self.assertRaises(TTransportException) as cm:
            self.publisher_transport.publish('foo', [])

        self.assertEquals(TTransportException.NOT_OPEN, cm.exception.type)
        self.assertEquals("Nats not connected!", cm.exception.message)

    def test_publisher_raises_exception_on_nats_publish_error(self):
        def nats_publish_raise_maxpayload(topic, data):
            raise ErrMaxPayload()

        def nats_publish_raise_connection_closed(topic, data):
            raise ErrConnectionClosed()

        self.nats_client.publish = nats_publish_raise_maxpayload
        with self.assertRaises(TTransportException):
            self.publisher_transport.publish('foo', 'bar')

        self.nats_client.publish = nats_publish_raise_connection_closed
        with self.assertRaises(TTransportException):
            self.publisher_transport.publish('foo', 'bar')

    def test_publish_raises_exception_if_nats_flush_fails(self):

        def nats_flush_raise_closed():
            raise ErrConnectionClosed()

        def nats_flush_raise_timeout():
            raise ErrTimeout()

        self.nats_client.flush = nats_flush_raise_closed

        with self.assertRaises(TTransportException):
            self.subscriber_transport.subscribe('foo', None)

        self.nats_client.flush = nats_flush_raise_timeout
        with self.assertRaises(TTransportException):
            self.subscriber_transport.subscribe('bar', None)

    def test_flush_publishes_to_formatted_subject(self):
        self.nats_client.is_connected = True
        self.publisher_transport._is_open = True
        self.publisher_transport._subject = "batman"
        self.publisher_transport._write_buffer = BytesIO()
        buff = bytearray(b'\x00\x00\x00\x00\x01')

        self.nats_client.publish.return_value = None

        self.publisher_transport.publish('batman', buff)

        self.nats_client.publish.assert_called_with("frugal.batman", buff)
