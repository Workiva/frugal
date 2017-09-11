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

from tornado import gen

from frugal.transport import FTransport


class FTransportBase(FTransport):
    """
    FTransportBase extends FTransport using the coroutine decorators used by
    all tornado FTransports.
    """
    def is_open(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def open(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def close(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def oneway(self, context, payload):
        raise NotImplementedError('You must override this.')

    @gen.coroutine
    def request(self, context, payload):
        raise NotImplementedError('You must override this.')
