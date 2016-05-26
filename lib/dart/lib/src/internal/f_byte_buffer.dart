part of frugal;

/**
 * This is an internal-only class. Don't use it!
 */
class FByteBuffer {
  final Uint8List _buff;
  int _writeIndex = 0;
  int _readIndex = 0;

  FByteBuffer(int capacity) : this._buff = new Uint8List(capacity);

  FByteBuffer.fromUint8List(Uint8List buff) : this._buff = buff;

  int get writeRemaining => _buff.length - _writeIndex;
  int get readRemaining => _buff.length - _readIndex;

  int read(Uint8List buffer, int offset, int length) {
    var n = _transfer(_buff, buffer, _readIndex, offset, length);
    _readIndex += n;
    return n;
  }

  int write(Uint8List buffer, int offset, int length) {
    var n = _transfer(buffer, _buff, offset, _writeIndex, length);
    _writeIndex += n;
    return n;
  }

  int _transfer(Uint8List source, Uint8List dest, int sourceOffset,
      int destOffset, int length) {
    // Can write at most what's left destination buffer
    var bytesInDest = dest.length - destOffset;
    var amtToCopy = length > bytesInDest ? bytesInDest : length;

    if (amtToCopy > 0) {
      // See how many bytes are source buffer
      var bytesInSource = source.length - sourceOffset;
      amtToCopy = amtToCopy > bytesInSource ? bytesInSource : amtToCopy;

      // Write the appropriate range
      if (amtToCopy > 0) {
        var range = source.getRange(sourceOffset, sourceOffset + amtToCopy);
        dest.setRange(destOffset, destOffset + amtToCopy, range);
      }
    }
    return amtToCopy;
  }

  Uint8List asUint8List() {
    return new Uint8List.fromList(_buff.sublist(0, _writeIndex));
  }

  void reset() {
    _writeIndex = 0;
    _readIndex = 0;
  }
}
