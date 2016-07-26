part of frugal;

/// FTransport is Frugal's equivalent of Thrift's TTransport. FTransport
/// extends TTransport and exposes some additional methods. An FTransport
/// has an FRegistry, so it provides methods for setting the FRegistry and
/// registering and unregistering an FAsyncCallback to an FContext.
abstract class FTransport extends TTransport {
  static const REQUEST_TOO_LARGE = 100;
  static const RESPONSE_TOO_LARGE = 101;

  FRegistry _registry;
  MonitorRunner _monitor;
  StreamController _closeController = new StreamController.broadcast();
  Stream<Object> get onClose => _closeController.stream;

  // TODO: Remove with 2.0
  @deprecated
  static const DEFAULT_WATERMARK = const Duration(seconds: 5);
  @deprecated
  TTransport _transport;
  @deprecated
  Duration _highWatermark = DEFAULT_WATERMARK;

  /// With 2.0, implementations of FTransport will not typically
  /// wrap TTransport implementations - except for FAdapterTransport.
  @deprecated
  void set transport(TTransport transport) {
    _transport = transport;
  }

  /// Set an FTransportMonitor on the transport
  void set monitor(FTransportMonitor monitor) {
    _monitor = new MonitorRunner(monitor, this);
  }

  // TODO: Don't implement with 2.0
  @override
  bool get isOpen => _transport.isOpen;

  // TODO: Don't implement with 2.0
  @override
  Future open() => _transport.open();

  @override
  Future close() => closeWithException(null);

  /// Close transport with the given exception
  Future closeWithException(cause) async {
    // TODO: Remove the transport close with 2.0
    await _transport?.close();
    await _signalClose(cause);
  }

  /// TODO: Throw error when reading on FTransport
  @override
  int read(Uint8List buffer, int offset, int length) {
    return _transport.read(buffer, offset, length);
  }

  /// TODO: Implement a write buffer for implementing classes
  /// to use with 2.0.
  @override
  void write(Uint8List buffer, int offset, int length) {
    _transport.write(buffer, offset, length);
  }

  // TODO: Don't implement with 2.0
  @override
  Future flush() => _transport.flush();

  /// Set the Registry on the transport.
  void setRegistry(FRegistry registry) {
    if (registry == null) {
      throw new FError.withMessage('registry cannot be null');
    }
    if (_registry != null) {
      return;
    }
    _registry = registry;
  }

  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback) {
    if (_registry == null) {
      throw new FError.withMessage('transport registry not set');
    }
    _registry.register(ctx, callback);
  }

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx) {
    if (_registry == null) {
      throw new FError.withMessage('transport registry not set');
    }
    _registry.unregister(ctx);
  }

  /// Set the maximum amount of time a frame is allowed to await processing
  /// before triggering transport overload logic. For now, this just consists
  /// of logging a warning. If not set, the default is 5 seconds.
  /// With 2.0, this will be an implementation detail for transports
  /// which buffer read data.
  @deprecated
  void setHighWatermark(Duration watermark) {
    _highWatermark = watermark;
  }

  /// Execute a frugal frame (NOTE: this frame must include the frame size).
  void executeFrame(Uint8List frame) {
    if (_registry == null) {
      throw new FError.withMessage('transport registry not set');
    }
    _registry.execute(frame.sublist(4));
  }

  Future _signalClose(cause) async {
    _closeController.add(cause);
    await _monitor?.onClose(cause);
  }
}
