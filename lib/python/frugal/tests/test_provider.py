import unittest
from mock import patch

from frugal.transport.transport import FTransport
from frugal.protocol.protocol import FProtocol
from frugal.provider import FScopeProvider


class TestFScopeProvider(unittest.TestCase):

    @patch('frugal.transport.scope_transport_factory.FScopeTransportFactory')
    @patch('frugal.protocol.protocol_factory.FProtocolFactory')
    @patch('thrift.protocol.TProtocol.TProtocolBase')
    def test_new_provider(self, mock_transport_factory,
                          mock_protocol_factory, mock_thrift_protocol):
        transport = FTransport()
        protocol = FProtocol(mock_thrift_protocol)

        mock_transport_factory.get_transport.return_value = transport
        mock_protocol_factory.get_protocol.return_value = protocol

        provider = FScopeProvider(mock_transport_factory, mock_protocol_factory)

        trans, prot = provider.new()

        self.assertEqual(transport, trans)
        self.assertEqual(protocol, prot)
