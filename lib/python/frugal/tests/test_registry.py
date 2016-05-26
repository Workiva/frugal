import unittest

from frugal.registry import FClientRegistry
from frugal.context import FContext
from frugal.exceptions import FException


class TestClientRegistry(unittest.TestCase):

    def test_register(self):
        registry = FClientRegistry()
        context = FContext("fooid")
        context._set_op_id(123)
        callback = self.fake_callback
        registry.register(context, callback)
        self.assertEqual(1, len(registry._handlers))

    def test_register_with_existing_op_id(self):
        registry = FClientRegistry()
        context = FContext("fooid")
        context._set_op_id(0)
        callback = self.fake_callback

        registry.register(context, callback)
        try:
            registry.register(context, callback)
        except FException as ex:
            self.assertEquals("context already registered", ex.message)

    def test_unregister(self):
        registry = FClientRegistry()
        context = FContext("fooid")
        context._set_op_id(1)
        callback = self.fake_callback
        registry.register(context, callback)
        self.assertEqual(1, len(registry._handlers))
        registry.unregister(context)
        self.assertEqual(0, len(registry._handlers))

    def fake_callback():
        pass

