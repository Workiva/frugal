part of frugal;

/// FScopeTransportFactory produces FScopeTransports and is typically used by
/// an FScopeProvider.
abstract class FScopeTransportFactory {
  FScopeTransport getTransport();
}
