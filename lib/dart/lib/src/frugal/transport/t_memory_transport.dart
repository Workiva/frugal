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

/// An [FByteBuffer]-backed implementation of [TTransport].
class TMemoryTransport extends TTransport {
  static const _defaultBufferLength = 1024;

  final FByteBuffer _buff;

  /// Create a new [TMemoryTransport] instance with the optional size capacity.
  TMemoryTransport([int capacity])
      : _buff = new FByteBuffer(capacity ?? _defaultBufferLength);

  /// Create a new [TMemoryTransport] instance from the given buffer.
  TMemoryTransport.fromUint8List(Uint8List buffer)
      : _buff = new FByteBuffer.fromUint8List(buffer);

  @override
  bool get isOpen => true;

  @override
  Future open() async {}

  @override
  Future close() async {}

  @override
  int read(Uint8List buffer, int offset, int length) {
    return _buff.read(buffer, offset, length);
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    _buff.write(buffer, offset, length);
  }

  @override
  Future flush() async {}

  /// The bytes currently stored in the transport.
  Uint8List get buffer => _buff.asUint8List();
}
