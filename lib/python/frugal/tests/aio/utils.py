import asyncio
from functools import wraps

from asyncio.test_utils import TestCase


def async_runner(f):
    @wraps(f)
    def wrapper(*args, **kwargs):
        asyncio.get_event_loop().run_until_complete(f(*args, **kwargs))
    return wrapper


def default_gen():
    """Advance test time by the time requested to allow timeouts to expire."""
    time_requested = 1
    while time_requested != 0:
        time_requested = yield time_requested


class AsyncIOTestCase(TestCase):
    def setUp(self, gen=None):
        super().setUp()
        if gen is None:
            gen = default_gen
        self.loop = self.new_test_loop(gen=gen)
        asyncio.set_event_loop(self.loop)

    def tearDown(self):
        super().tearDown()
        asyncio.set_event_loop(None)
