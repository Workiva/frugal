part of frugal;

/// Wraps a Thrift TTransport. Used for frugal Scopes.
abstract class FScopeTransport extends TTransport {
  @override
  int read(Uint8List buffer, int offset, int length) {
    throw new TTransportError(TTransportErrorType.UNKNOWN,
        'Cannot read directly from FScopeTransport');
  }

  /// set the publish topic.
  void setTopic(String topic);

  /// set the subscribe topic and opens the transport.
  Future subscribe(String topic, FAsyncCallback callback);

  /// Signals errors that occur on the transport. An error here
  /// indicates that the transport is now closed.
  Stream<Error> get error;
}
