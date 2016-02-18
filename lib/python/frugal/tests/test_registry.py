import unittest

from frugal.registry import FClientRegistry, FAsyncCallback
from frugal.context import FContext
from frugal.exceptions import FException


class TestClientRegistry(unittest.TestCase):

    def test_register(self):
        registry = FClientRegistry()
        context = FContext("fooid")
        callback = FAsyncCallback()
        registry.register(context, callback)
        self.assertEqual(1, len(registry._handlers))

    def test_register_with_existing_op_id(self):
        registry = FClientRegistry()
        context = FContext("fooid")
        callback = FAsyncCallback()
        registry.register(context, callback)
        self.assertRaises(FException,
                          registry.register, context, callback)

    def test_unregister(self):
        registry = FClientRegistry()
        context = FContext("fooid")
        callback = FAsyncCallback()
        registry.register(context, callback)
        self.assertEqual(1, len(registry._handlers))
        registry.unregister(context)
        self.assertEqual(0, len(registry._handlers))

