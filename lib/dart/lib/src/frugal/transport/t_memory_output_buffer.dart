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

/// An implementation of a framed [TTransport] using a memory buffer and is used
/// exclusively for writing. The size of this buffer is optionally limited. If
/// limited, writes which cause the buffer to exceed its size limit throw an
/// FMessageSizeError.
class TMemoryOutputBuffer extends TTransport {
  final List<int> _writeBuffer = [];
  final int _limit;

  /// Create an TMemoryOutputBuffer with a buffer size limit.
  ///
  /// [size] is size limit of the buffer. Note: If [size] is non-positive,
  /// no limit will be enforced on the buffer.
  TMemoryOutputBuffer([int size]) : _limit = size ?? 0;

  @override
  bool get isOpen => true;

  @override
  Future open() => new Future.value();

  @override
  Future close() => new Future.value();

  @override
  int read(Uint8List buffer, int offset, int length) {
    throw new UnsupportedError('Cannot read from ${this.toString()}');
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    // Leave room for the frame size
    if (_limit > 0 && _writeBuffer.length + length + 4 > _limit) {
      _writeBuffer.clear();
      throw new TTransportError(FrugalTTransportErrorType.REQUEST_TOO_LARGE);
    }
    _writeBuffer.addAll(buffer.sublist(offset, offset + length));
  }

  @override
  Future flush() => new Future.value();

  /// Query if data has been written to the transport.
  bool get hasWriteData => _writeBuffer.length > 0;

  /// The number of bytes that have been written to the transport.
  int get size => _writeBuffer.length;

  /// Get the framed bytes that have been written to the transport.
  Uint8List get writeBytes {
    var bytes = new Uint8List(4 + _writeBuffer.length);
    bytes.buffer.asByteData().setUint32(0, _writeBuffer.length);
    bytes.setAll(4, _writeBuffer);
    return bytes;
  }

  /// Clear the write buffer
  void reset() {
    _writeBuffer.clear();
  }
}
