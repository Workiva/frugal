import unittest

from frugal.exceptions import FException
from frugal.transport.nats_scope_transport import FNatsScopeTransport


class TestFNatsScopeTransport(unittest.TestCase):

    def setUp(self):
        self.transport = FNatsScopeTransport()

    def test_lock_topic_sets_topic(self):
        expected = "topic"

        self.transport.lock_topic(expected)

        self.assertEqual(expected, self.transport._subject)

    def test_unlock_topic_resets_topic(self):

        self.transport.lock_topic("topic")
        self.transport.unlock_topic()

        self.assertEqual("", self.transport._subject)

    def test_subscriber(self):
        expected = "topic"
        self.transport.subscribe(expected)

        self.assertTrue(self.transport._pull)
        self.assertEqual(expected, self.transport._subject)

    def test_subscriber_cannot_lock_topic(self):
        expected = "topic"
        self.transport.subscribe(expected)

        try:
            self.transport.lock_topic(expected)
        except FException as ex:
            self.assertEquals("Subscriber cannot lock topic.", ex.message)

    def test_subscriber_cannot_unlock_topic(self):
        expected = "topic"
        self.transport.subscribe(expected)

        try:
            self.transport.unlock_topic()
        except FException as ex:
            self.assertEquals("Subscriber cannot unlock topic.", ex.message)

