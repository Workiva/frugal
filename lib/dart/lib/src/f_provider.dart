part of frugal;

/// FScopeProvider produces FScopeTransports and FProtocols for use by pub/sub
/// scopes. It does this by wrapping an FScopeTransportFactory and
/// FProtocolFactory.
class FScopeProvider {
  final FScopeTransportFactory fTransportFactory;
  final FProtocolFactory fProtocolFactory;

  FScopeProvider(this.fTransportFactory, this.fProtocolFactory);
}

/// FServiceProvider is the service equivalent of FScopeProvider. It produces
/// FTransports and FProtocols for use by RPC service clients.
class FServiceProvider {
  final FTransport fTransport;
  final FProtocolFactory fProtocolFactory;

  FServiceProvider(this.fTransport, this.fProtocolFactory);
}
