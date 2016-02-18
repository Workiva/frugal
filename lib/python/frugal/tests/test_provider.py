import unittest
from mock import patch

from frugal.transport.transport import FTransport
from frugal.protocol.protocol import FProtocol
from frugal.provider import FScopeProvider


class TestFScopeProvider(unittest.TestCase):

    @patch('frugal.transport.scope_transport_factory.FScopeTransportFactory')
    @patch('frugal.protocol.protocol_factory.FProtocolFactory')
    def test_build_client(self, mock_transport_factory, mock_protocol_factory):
        transport = FTransport()
        protocol = FProtocol(transport)

        mock_transport_factory.get_transport.return_value = transport
        mock_protocol_factory.get_protocol.return_value = protocol

        provider = FScopeProvider(mock_transport_factory, mock_protocol_factory)

        client = provider.build_client()

        self.assertEqual(transport, client.get_transport())
        self.assertEqual(protocol, client.get_protocol())
