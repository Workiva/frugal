part of frugal;

/// FAdapterTransport returns an FTransport which uses the given
/// TSocketTransport for write/callback operations in a way that is compatible
/// with Frugal. Messages received on the TSocket (i.e. Frugal frames) are
/// routed to the FRegistry's execute method.
class FAdapterTransport extends FTransport {
  final Logger log = new Logger('FAdapterTransport');
  _TFramedTransport _framedTransport;

  FAdapterTransport(TSocketTransport transport)
      : _framedTransport = new _TFramedTransport(transport.socket) {
    // If there is an error on the socket, close the transport pessimistically.
    // This error is already logged upstream in TSocketTransport.
    transport.socket.onError.listen((e) => closeWithException(e));
    // Forward state changes on to the transport monitor.
    // Note: Just forwarding OPEN on for the time-being.
    transport.socket.onState.listen((state) {
      if (state == TSocketState.OPEN) _monitor?.signalOpen();
    });
  }

  @override
  bool get isOpen => _framedTransport.isOpen;

  @override
  Future open() => _framedTransport.open();

  @override
  Future close() => closeWithException(null);

  @override
  void write(Uint8List buffer, int offset, int length) {
    _framedTransport.write(buffer, offset, length);
  }

  @override
  Future flush() {
    if (!isOpen) {
      throw new TTransportError(TTransportErrorType.NOT_OPEN);
    }
    return _framedTransport.flush();
  }

  Future closeWithException(cause) async {
    await _framedTransport.close();
    await super.closeWithException(cause);
  }

  @override
  void setRegistry(FRegistry registry) {
    super.setRegistry(registry);
    _framedTransport.onFrame.listen((_FrameWrapper frame) {
      try {
        _registry.execute(frame.frameBytes);
      } catch (e) {
        // Fatal error. Close the transport.
        log.severe("FAsyncCallback had a fatal error ${e.toString()}." +
            "Closing transport.");
        closeWithException(e);
      }
    });
  }
}
