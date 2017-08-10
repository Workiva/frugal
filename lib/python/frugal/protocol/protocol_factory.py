from thrift.protocol.TProtocol import TProtocolFactory

from frugal.protocol import FProtocol, FUniversalProtocol


class FProtocolFactory(object):
    """FProtocolFactory creates FProtocols."""

    def __init__(self, t_protocol_factory):
        """Initialize FProtocolFactory.

        Args:
            t_protocol_factory: Thrift TProtocolFactory.
        """
        self._t_protocol_factory = t_protocol_factory

    def get_protocol(self, transport):
        """Return FProtocol for the given TTransport.

        Args:
            transport: Thrift TTransport.

        Returns:
            FProtocol wrapping the TTransport.
        """
        return FProtocol(self._t_protocol_factory.getProtocol(transport))


class FUniversalProtocolFactory(TProtocolFactory):
    def getProtocol(self, trans, **kwargs):
        return FUniversalProtocol(trans, **kwargs)
