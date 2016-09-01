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

  int _capacity;
  List<int> _writeBuffer;
  List<int> get writeBuffer => _writeBuffer;

  FTransport({int capacity: 0}) {
    this._capacity = capacity;
    this._writeBuffer = [];
  }

  void clearWriteBuffer() {
    this._writeBuffer = [];
  }

  /// Set an FTransportMonitor on the transport
  void set monitor(FTransportMonitor monitor) {
    _monitor = new MonitorRunner(monitor, this);
  }

  @override
  Future close() => closeWithException(null);

  /// Close transport with the given exception
  Future closeWithException(cause) async {
    await _signalClose(cause);
  }

  @override
  int read(Uint8List buffer, int offset, int length) {
    throw new UnsupportedError("Cannot call read on FTransport");
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    if (offset + length > buffer.length) {
      throw new ArgumentError('The range exceeds the buffer length');
    }

    if (_capacity > 0 && length + _writeBuffer.length > _capacity) {
      throw new FMessageSizeError.request();
    }

    _writeBuffer.addAll(buffer.sublist(offset, offset + length));
  }

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
