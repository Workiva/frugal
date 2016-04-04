part of frugal;

/// Provider for Frugal Scopes
class FScopeProvider {
  final FScopeTransportFactory fTransportFactory;
  final FProtocolFactory fProtocolFactory;

  FScopeProvider(this.fTransportFactory, this.fProtocolFactory);
}

/// Provider for Frugal Services
class FServiceProvider {
  final FTransport fTransport;
  final FProtocolFactory fProtocolFactory;

  FServiceProvider(this.fTransport, this.fProtocolFactory);
}
