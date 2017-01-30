import mock
from io import BytesIO

from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase
from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FrugalTTransportExceptionType
from frugal.tornado.transport import FNatsPublisherTransport
from frugal.tornado.transport import FNatsSubscriberTransport


class TestFNatsScopeTransport(AsyncTestCase):

    def setUp(self):
        super(TestFNatsScopeTransport, self).setUp()

        self.nats_client = mock.Mock()

        self.publisher_transport = FNatsPublisherTransport(self.nats_client)
        self.subscriber_transport = FNatsSubscriberTransport(
            self.nats_client, "Q")

    @gen_test
    def test_subscriber(self):
        future = concurrent.Future()
        future.set_result(235)
        self.nats_client.subscribe_async.return_value = future
        topic = 'bar'

        yield self.subscriber_transport.subscribe(topic, None)
        self.nats_client.subscribe_async.assert_called_once_with(
            'frugal.bar',
            queue='Q',
            cb=mock.ANY,
        )
        self.assertEqual(self.subscriber_transport._sub_id, 235)

    @gen_test
    def test_open_throws_exception_if_nats_not_connected(self):
        self.nats_client.is_connected = False

        with self.assertRaises(TTransportException) as cm:
            yield self.publisher_transport.open()

        self.assertEquals(
            FrugalTTransportExceptionType.NOT_OPEN, cm.exception.type)
        self.assertEquals("Nats not connected!", cm.exception.message)

    @gen_test
    def test_open_when_subscriber_calls_subscribe(self):
        self.nats_client.is_connected = True

        f = concurrent.Future()
        f.set_result(1)
        self.nats_client.subscribe_async.return_value = f

        yield self.subscriber_transport.subscribe('foo', None)

        self.nats_client.subscribe_async.assert_called()
        self.assertTrue(self.subscriber_transport.is_subscribed())

    @gen_test
    def test_publish_throws_if_max_message_size_exceeded(self):
        self.nats_client.is_connected = True
        self.publisher_transport._is_open = True

        buff = bytearray(1024 * 2048)
        with self.assertRaises(TTransportException) as cm:
            yield self.publisher_transport.publish('foo', buff)

        self.assertEqual(FrugalTTransportExceptionType.REQUEST_TOO_LARGE,
                         cm.exception.type)
        self.assertEqual("Message exceeds NATS max message size",
                         cm.exception.message)

    @gen_test
    def test_publish_throws_if_transport_not_open(self):
        self.nats_client.is_connected = False

        with self.assertRaises(TTransportException) as cm:
            yield self.publisher_transport.publish('foo', [])

        self.assertEquals(
            FrugalTTransportExceptionType.NOT_OPEN, cm.exception.type)
        self.assertEquals("Nats not connected!", cm.exception.message)

    @gen_test
    def test_flush_publishes_to_formatted_subject(self):
        self.nats_client.is_connected = True
        self.publisher_transport._is_open = True
        self.publisher_transport._subject = "batman"
        self.publisher_transport._write_buffer = BytesIO()
        buff = bytearray(b'\x00\x00\x00\x00\x01')

        f = concurrent.Future()
        f.set_result("")
        self.nats_client.publish.return_value = f

        yield self.publisher_transport.publish('batman', buff)

        self.nats_client.publish.assert_called_with("frugal.batman", buff)
