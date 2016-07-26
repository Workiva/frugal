part of frugal;

/// A framed implementation of TTransport. Has stream for consuming
/// entire frames. Disallows direct reads.
class _TFramedTransport extends TTransport {
  final Logger log = new Logger('frugal.transport._TFramedTransport');
  static const int headerByteCount = 4;
  final Uint8List writeHeaderBytes = new Uint8List(headerByteCount);

  final TSocket socket;
  final List<int> _writeBuffer = [];
  final List<int> _readBuffer = [];
  final List<int> _readHeaderBytes = [];
  int _frameSize;

  StreamController<_FrameWrapper> _frameStream = new StreamController();

  bool _isOpen = false;

  final Uint8List headerBytes = new Uint8List(headerByteCount);

  StreamSubscription _messageSub;

  /// Instantiate new TFramedTransport for the given TSocket.
  /// Add a listener to the socket state that opens/closes the
  /// transport in response to socket state changes.
  _TFramedTransport(this.socket) {
    if (socket == null) {
      throw new ArgumentError.notNull('socket');
    }
    // Listen and react to state changes on the TSocket
    socket.onState.listen((state) {
      switch (state) {
        case TSocketState.OPEN:
          open();
          break;
        case TSocketState.CLOSED:
          close();
          break;
        default:
          // Should not happen.
          log.log(Level.WARNING, 'Unhandled TSocketState $state');
      }
    });
  }

  void _reset({bool isOpen: false}) {
    _isOpen = isOpen;
    _writeBuffer.clear();
    _readBuffer.clear();
    _messageSub?.cancel();
  }

  /// Stream for getting frame data.
  Stream<_FrameWrapper> get onFrame => _frameStream.stream;

  @override
  bool get isOpen => _isOpen;

  /// Opens the transport.
  /// Will reset the write/read buffers and socket onMessage listener.
  /// Will also open the underlying TSocket (if not already open).
  @override
  Future open() async {
    _reset(isOpen: true);
    // TODO: With 2.0, don't check socket state before
    // calling through to socket open.
    if (socket.isClosed) {
      await socket.open();
    }
    _messageSub = socket.onMessage.listen(messageHandler);
  }

  /// Closes the transport.
  /// Will reset the write/read buffers and socket onMessage listener.
  /// Will also close the underlying TSocket (if not already closed).
  @override
  Future close() async {
    _reset(isOpen: false);
    // TODO: With 2.0, don't check socket state before
    // calling through to socket close.
    if (socket.isOpen) {
      await socket.close();
    }
  }

  /// Direct reading is not allowed. To consume read data listen
  /// to onFrame.
  int read(Uint8List buffer, int offset, int length) {
    throw new TTransportError(TTransportErrorType.UNKNOWN,
        'frugal: cannot read directly from _TFramedSocket.');
  }

  /// Handler for messages received on the TSocket.
  void messageHandler(Uint8List list) {
    var offset = 0;
    if (_frameSize == null) {
      // Not enough bytes to get the frame length. Add these and move on.
      if ((_readHeaderBytes.length + list.length) < headerByteCount) {
        _readHeaderBytes.addAll(list);
        return;
      }

      // Get the frame size
      var headerBytesToGet = headerByteCount - _readHeaderBytes.length;
      _readHeaderBytes.addAll(list.getRange(0, headerBytesToGet));
      var frameBuffer = new Uint8List.fromList(_readHeaderBytes).buffer;
      _frameSize = frameBuffer.asByteData().getInt32(0);
      _readHeaderBytes.clear();
      offset += headerBytesToGet;
    }

    if (_frameSize < 0) {
      // TODO: Put this error on an error stream and bubble it up.
      throw new TTransportError(TTransportErrorType.UNKNOWN,
          'Read a negative frame size: $_frameSize');
    }

    // Grab up to the frame size in bytes
    var bytesToGet = min(_frameSize - _readBuffer.length, list.length - offset);
    _readBuffer.addAll(list.getRange(offset, offset + bytesToGet));

    // Have an entire frame. Fire it off and reset.
    if (_readBuffer.length == _frameSize) {
      _frameStream.add(new _FrameWrapper(
          new Uint8List.fromList(_readBuffer), new DateTime.now()));
      _readBuffer.clear();
      _frameSize = null;
    }

    // More bytes to get. Run through the handler again.
    if ((bytesToGet + offset < list.length)) {
      messageHandler(new Uint8List.fromList(list.sublist(bytesToGet + offset)));
      return;
    }
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    if (buffer == null) {
      throw new ArgumentError.notNull('buffer');
    }

    if (offset + length > buffer.length) {
      throw new ArgumentError('The range exceeds the buffer length');
    }

    _writeBuffer.addAll(buffer.sublist(offset, offset + length));
  }

  @override
  Future flush() {
    int length = _writeBuffer.length;
    headerBytes.buffer.asByteData().setUint32(0, length);
    _writeBuffer.insertAll(0, headerBytes);
    var buff = new Uint8List.fromList(_writeBuffer);
    _writeBuffer.clear();
    return new Future.value(socket.send(buff));
  }
}

/// _FrameWrapper wraps a _TFramedTransport frame with a timestamp indicating
/// when it was placed in the frame buffer.
class _FrameWrapper {
  Uint8List frameBytes;
  DateTime timestamp;

  _FrameWrapper(this.frameBytes, this.timestamp) {}
}
