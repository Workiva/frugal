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

import 'dart:convert';
import 'dart:typed_data';

import 'package:frugal/src/frugal/utils/byte_manipulations.dart';
import 'package:thrift/thrift.dart';


var _encoder = new Utf8Encoder();
var _decoder = new Utf8Decoder();

class _Pair<A, B> {
  A one;
  B two;

  _Pair(this.one, this.two);
}

/// This is an internal-only class. Don't use it!
class Headers {
  /// Encode the headers
  static Uint8List encode(Map<String, String> headers) => encodeHeaders(headers);

  /// Reads the headers from a TTransport
  static Map<String, String> read(TTransport transport) => readHeaders(transport);

  /// Returns the headers from Frugal frame
  static Map<String, String> decodeFromFrame(Uint8List frame) => decodeHeadersFromFrame(frame);
}

/// Frugal protocol version V0
const _v0 = 0x00;

/// Encode the headers
Uint8List encodeHeaders(Map<String, String> headers) {
  var size = 0;
  // Get total frame size headers
  List<_Pair<List<int>, List<int>>> utf8Headers = new List();
  if (headers != null && headers.length > 0) {
    for (var name in headers.keys) {
      List<int> keyBytes = _encoder.convert(name);
      List<int> valueBytes = _encoder.convert(headers[name]);
      utf8Headers.add(new _Pair(keyBytes, valueBytes));

      // 4 bytes each for name, value length
      size += 8 + keyBytes.length + valueBytes.length;
    }
  }

  // Header buff = [version (1 byte), size (4 bytes), headers (size bytes)]
  var buff = new Uint8List(5 + size);

  // Write version
  buff[0] = _v0;

  // Write size
  writeInt(size, buff, 1);

  // Write headers
  if (utf8Headers.length > 0) {
    var i = 5;
    for (var pair in utf8Headers) {
      // Write name length
      var name = pair.one;
      writeInt(name.length, buff, i);
      i += 4;
      // Write name
      writeStringBytes(name, buff, i);
      i += name.length;

      // Write value length
      var value = pair.two;
      writeInt(value.length, buff, i);
      i += 4;
      writeStringBytes(value, buff, i);
      i += value.length;
    }
  }
  return buff;
}

/// Reads the headers from a TTransport
Map<String, String> readHeaders(TTransport transport) {
  // Buffer version
  var buff = new Uint8List(5);
  transport.readAll(buff, 0, 1);

  checkVersion(buff);

  // Read size
  transport.readAll(buff, 1, 4);
  var size = readInt(buff, 1);

  // Read the rest of the header bytes into a buffer
  buff = new Uint8List(size);
  transport.readAll(buff, 0, size);

  return readHeaderPairs(buff, 0, size);
}

/// Returns the headers from Frugal frame
Map<String, String> decodeHeadersFromFrame(Uint8List frame) {
  if (frame.length < 5) {
    throw new TProtocolError(TProtocolErrorType.INVALID_DATA,
        "invalid frame size ${frame.length}");
  }

  checkVersion(frame);

  return readHeaderPairs(frame, 5, readInt(frame, 1) + 5);
}

Map<String, String> readHeaderPairs(Uint8List buff, int start, int end) {
  Map<String, String> headers = {};
  for (var i = start; i < end; i) {
    // Read header name
    var nameSize = readInt(buff, i);
    i += 4;
    if (i > end || i + nameSize > end) {
      throw new TProtocolError(
          TProtocolErrorType.INVALID_DATA, "invalid protocol header name");
    }
    var name = _decoder.convert(buff, i, i + nameSize);
    i += nameSize;

    // Read header value
    var valueSize = readInt(buff, i);
    i += 4;
    if (i > end || i + valueSize > end) {
      throw new TProtocolError(
          TProtocolErrorType.INVALID_DATA, "invalid protocol header value");
    }
    var value = _decoder.convert(buff, i, i + valueSize);
    i += valueSize;

    // Set the pair
    headers[name] = value;
  }
  return headers;
}

// Evaluates the version and throws a TProtocolError if the version is unsupported
// Support more versions when available
void checkVersion(Uint8List frame) {
  if (frame[0] != _v0) {
    throw new TProtocolError(TProtocolErrorType.BAD_VERSION,
        "unsupported header version ${frame[0]}");
  }
}
