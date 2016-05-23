import unittest
from mock import patch

from frugal.transport.transport import FTransport
from frugal.protocol import FProtocolFactory
from frugal.provider import FScopeProvider


class TestFScopeProvider(unittest.TestCase):

    @patch('frugal.transport.transport_factory.FScopeTransportFactory')
    @patch('thrift.protocol.TProtocol.TProtocolBase')
    def test_new_provider(self, mock_transport_factory, mock_thrift_protocol):

        transport = FTransport()
        protocol_factory = FProtocolFactory(mock_transport_factory)

        mock_transport_factory.get_transport.return_value = transport

        provider = FScopeProvider(mock_transport_factory, protocol_factory)

        trans, prot_factory = provider.new()

        self.assertEqual(transport, trans)
        self.assertEqual(protocol_factory, prot_factory)
