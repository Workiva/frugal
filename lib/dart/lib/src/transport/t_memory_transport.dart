part of frugal;

/// A ReadByteBuffer-backed implementation of TTransport.
class TMemoryTransport extends TTransport {
  static const _DEFAULT_BUFFER_LENGTH = 1024;

  final FByteBuffer _buff;

  TMemoryTransport([int capacity])
      : _buff = new FByteBuffer(capacity ?? _DEFAULT_BUFFER_LENGTH);

  TMemoryTransport.fromUint8List(Uint8List buffer)
      : _buff = new FByteBuffer.fromUint8List(buffer);

  @override
  bool get isOpen => true;

  @override
  Future open() {
    return new Future.value();
  }

  @override
  Future close() {
    return new Future.value();
  }

  @override
  int read(Uint8List buffer, int offset, int length) {
    return _buff.read(buffer, offset, length);
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    _buff.write(buffer, offset, length);
  }

  @override
  Future flush() {
    return new Future.value();
  }
}
