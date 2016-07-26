part of frugal;

/// FMultiplexedTransport is a multiplexed Transport that routes frames to the
/// correct callbacks.
/// Deprecated - Use FAdapterTransport instead
@deprecated
class FMultiplexedTransport extends FTransport {
  final Logger log = new Logger('FMultiplexedTransport');
  _TFramedTransport _transport;
  FRegistry _registry;

  FMultiplexedTransport(TSocketTransport transport)
      : _transport = new _TFramedTransport(transport.socket) {
    super.transport = _transport;
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
  bool get isOpen => _transport.isOpen;

  @override
  Future close() => closeWithException(null);

  Future closeWithException(cause) async {
    await _transport.close();
    await _signalClose(cause);
  }

  @override
  void setRegistry(FRegistry registry) {
    if (registry == null) {
      throw new FError.withMessage('registry cannot be null');
    }
    if (_registry != null) {
      return;
    }

    _registry = registry;
    _transport.onFrame.listen((_FrameWrapper frame) {
      try {
        var dur = new DateTime.now().difference(frame.timestamp);
        if (dur > _highWatermark) {
          log.warning('frame spent ${dur} in the transport buffer, your '
              'consumer might be backed up');
        }
        _registry.execute(frame.frameBytes);
      } catch (e) {
        // TODO: Log the stacktrace
        // Fatal error. Close the transport.
        log.severe('AsyncCallback had a fatal error ${e.toString()}. '
            'Closing transport.');
        closeWithException(e);
      }
    });
  }

  @override
  void register(FContext ctx, FAsyncCallback callback) {
    if (_registry == null) {
      throw new FError.withMessage('transport registry not set');
    }
    _registry.register(ctx, callback);
  }

  @override
  void unregister(FContext ctx) {
    if (_registry == null) {
      throw new FError.withMessage('transport registry not set');
    }
    _registry.unregister(ctx);
  }
}
