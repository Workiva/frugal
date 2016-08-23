import asyncio
from functools import wraps

from asyncio.test_utils import TestCase


def async_runner(f):
    @wraps(f)
    def wrapper(*args, **kwargs):
        asyncio.get_event_loop().run_until_complete(f(*args, **kwargs))
    return wrapper


class AsyncIOTestCase(TestCase):
    def setUp(self):
        super().setUp()
        self.loop = self.new_test_loop()
        asyncio.set_event_loop(self.loop)

    def tearDown(self):
        super().tearDown()
        asyncio.set_event_loop(None)
