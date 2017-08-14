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

from frugal.aio.transport import FTransportBase
from frugal.tests.aio import utils as aio_utils


class TestFTransportBase(aio_utils.AsyncIOTestCase):

    def setUp(self):
        self.transport = FTransportBase()

        super().setUp()

    def test_is_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            self.transport.is_open()

    @aio_utils.async_runner
    async def test_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            await self.transport.open()

    @aio_utils.async_runner
    async def test_close_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            await self.transport.close()

    @aio_utils.async_runner
    async def test_oneway_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            await self.transport.oneway(None, [])

    @aio_utils.async_runner
    async def test_request_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError):
            await self.transport.request(None, [])

