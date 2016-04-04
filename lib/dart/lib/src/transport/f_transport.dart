part of frugal;

/// FTransport is a TTransport for services.
abstract class FTransport extends TTransport {
  static const REQUEST_TOO_LARGE = 100;
  static const RESPONSE_TOO_LARGE = 101;
  static const DEFAULT_WATERMARK = const Duration(seconds: 5);

  TTransport _transport;
  MonitorRunner _monitor;
  Duration _highWatermark = DEFAULT_WATERMARK;

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
    if (_monitor != null) await _monitor.onClose(cause);
  }
}
