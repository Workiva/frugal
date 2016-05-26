import unittest
import mock

from frugal.subscription import FSubscription


class TestFSubscription(unittest.TestCase):

    def test_get_topic(self):
        trans = mock.Mock()

        sub = FSubscription("topic", trans)

        self.assertEquals("topic", sub.get_topic())

    def test_unsubscribe_closes_transport(self):
        trans = mock.Mock()

        sub = FSubscription("topic", trans)

        sub.unsubscribe()

        trans.close.assert_called_with()
