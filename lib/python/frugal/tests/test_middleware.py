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

import unittest

from frugal import middleware


class TestServiceMiddleware(unittest.TestCase):

    def test_apply_middleware(self):
        call_dict1 = {}
        middleware1 = new_middleware(call_dict1)
        call_dict2 = {}
        middleware2 = new_middleware(call_dict2)
        handler = TestHandler()
        method = middleware.Method(handler.handler_method,
                                   [middleware1, middleware2])
        arg = 42

        ret = method([arg])

        self.assertEqual('foo', ret)
        self.assertEqual(arg+2, handler.called_arg)
        self.assertEqual(arg, call_dict2['called_arg'])
        self.assertEqual(arg + 1, call_dict1['called_arg'])

    def test_no_middleware(self):
        handler = TestHandler()
        method = middleware.Method(handler.handler_method, [])
        arg = 42

        ret = method([arg])

        self.assertEqual('foo', ret)
        self.assertEqual(arg, handler.called_arg)


class TestHandler(object):

    def __init__(self):
        self.called_arg = None

    def handler_method(self, x):
        self.called_arg = x
        return 'foo'


def new_middleware(call_dict):
    def test_middleware(next):
        def invocation_handler(method, args):
            call_dict['called_arg'] = args[0]
            args[0] = int(args[0]) + 1
            return next(method, args)
        return invocation_handler
    return test_middleware


