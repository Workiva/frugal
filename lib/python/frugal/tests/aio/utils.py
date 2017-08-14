# Copyright 2017 Workiva
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
