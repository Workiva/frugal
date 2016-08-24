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


