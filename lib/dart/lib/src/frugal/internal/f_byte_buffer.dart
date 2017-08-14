/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

part of frugal.src.frugal;

/// This is an internal-only class. Don't use it!
class FByteBuffer {
  final Uint8List _buff;
  int _writeIndex = 0;
  int _readIndex = 0;

  /// Creates an [FByteBuffer] with the given capacity.
  FByteBuffer(int capacity) : this._buff = new Uint8List(capacity);

  /// Creates an [FByteBuffer] from the given buffer.
  FByteBuffer.fromUint8List(Uint8List buff)
      : this._buff = buff,
        _writeIndex = buff.lengthInBytes;

  /// Number of write bytes remaining in the buffer.
  int get writeRemaining => _buff.length - _writeIndex;

  /// Number of read bytes remaining in the buffer.
  int get readRemaining => _buff.length - _readIndex;

  /// Reads up to [length] bytes into [buffer], starting at [offset].
  /// Returns the number of bytes actually read.
  int read(Uint8List buffer, int offset, int length) {
    var n = _transfer(_buff, buffer, _readIndex, offset, length);
    _readIndex += n;
    return n;
  }

  /// Writes up to [length] bytes from the buffer.
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

  /// Returns the buffer as a [Uint8List].
  Uint8List asUint8List() {
    return new Uint8List.fromList(_buff.sublist(0, _writeIndex));
  }

  /// Resets the read/write indices.
  void reset() {
    _writeIndex = 0;
    _readIndex = 0;
  }
}
