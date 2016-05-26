part of frugal;

/// FScopeTransport extends Thrift's TTransport and is used exclusively for
/// pub/sub scopes. Subscribers use an FScopeTransport to subscribe to a
/// pub/sub topic. Publishers use it to publish to a topic.
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
