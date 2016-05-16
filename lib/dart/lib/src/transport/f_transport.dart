part of frugal;

/// FTransport is Frugal's equivalent of Thrift's TTransport. FTransport
/// extends TTransport and exposes some additional methods. An FTransport
/// typically has an FRegistry, so it provides methods for setting the
/// FRegistry and registering and unregistering an FAsyncCallback to an
/// FContext. It also allows a way for setting an FTransportMonitor and a
/// high-water mark provided by an FServer.
///
/// FTransport wraps a TTransport, meaning all existing TTransport
/// implementations will work in Frugal. However, all FTransports must used a
/// framed protocol, typically implemented by wrapping a TFramedTransport.
///
/// Most Frugal language libraries include an FMuxTransport implementation,
/// which uses a worker pool to handle messages in parallel.
abstract class FTransport extends TTransport {
  static const REQUEST_TOO_LARGE = 100;
  static const RESPONSE_TOO_LARGE = 101;
  static const DEFAULT_WATERMARK = const Duration(seconds: 5);

  TTransport _transport;
  MonitorRunner _monitor;
  Duration _highWatermark = DEFAULT_WATERMARK;
  StreamController _closeController = new StreamController.broadcast();
  Stream<Object> get onClose => _closeController.stream;

  void set transport(TTransport transport) {
    _transport = transport;
  }

  void set monitor(FTransportMonitor monitor) {
    _monitor = new MonitorRunner(monitor, _transport);
  }

  @override
  bool get isOpen => _transport.isOpen;

  @override
  Future open() => _transport.open();

  @override
  int read(Uint8List buffer, int offset, int length) {
    return _transport.read(buffer, offset, length);
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    _transport.write(buffer, offset, length);
  }

  @override
  Future flush() => _transport.flush();

  /// Set the Registry on the transport.
  void setRegistry(FRegistry registry);

  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback);

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx);

  /// Set the maximum amount of time a frame is allowed to await processing
  /// before triggering transport overload logic. For now, this just consists
  /// of logging a warning. If not set, the default is 5 seconds.
  void setHighWatermark(Duration watermark) {
    this._highWatermark = watermark;
  }

  Future _signalClose(cause) async {
    _closeController.add(cause);
    await _monitor?.onClose(cause);
  }
}
