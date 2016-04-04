part of frugal;

/// Factory for FProtocol
class FProtocolFactory {
  TProtocolFactory _tProtocolFactory;

  FProtocolFactory(this._tProtocolFactory) {}

  FProtocol getProtocol(TTransport transport) {
    return new FProtocol(_tProtocolFactory.getProtocol(transport));
  }
}
