import unittest
from mock import patch

from frugal.transport.transport import FTransport
from frugal.registry import FRegistry

class TestFTransport(unittest.TestCase):

    @patch.object(FRegistry, 'register')
    def test_register(self, mock_registry):

        transprt = FTransport(mock_registry)
