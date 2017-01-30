part of frugal.src.frugal;

var _encoder = new Utf8Encoder();
var _decoder = new Utf8Decoder();

class _Pair<A, B> {
  A one;
  B two;

  _Pair(this.one, this.two);
}

/// This is an internal-only class. Don't use it!
class Headers {
  /// Frugal protocol version V0
  static const _v0 = 0x00;

  /// Encode the headers
  static Uint8List encode(Map<String, String> headers) {
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
    _writeInt(size, buff, 1);

    // Write headers
    if (utf8Headers.length > 0) {
      var i = 5;
      for (var pair in utf8Headers) {
        // Write name length
        var name = pair.one;
        _writeInt(name.length, buff, i);
        i += 4;
        // Write name
        _writeStringBytes(name, buff, i);
        i += name.length;

        // Write value length
        var value = pair.two;
        _writeInt(value.length, buff, i);
        i += 4;
        _writeStringBytes(value, buff, i);
        i += value.length;
      }
    }
    return buff;
  }

  /// Reads the headers from a TTransport
  static Map<String, String> read(TTransport transport) {
    // Buffer version
    var buff = new Uint8List(5);
    transport.readAll(buff, 0, 1);

    _checkVersion(buff);

    // Read size
    transport.readAll(buff, 1, 4);
    var size = _readInt(buff, 1);

    // Read the rest of the header bytes into a buffer
    buff = new Uint8List(size);
    transport.readAll(buff, 0, size);

    return _readPairs(buff, 0, size);
  }

  /// Returns the headers from Frugal frame
  static Map<String, String> decodeFromFrame(Uint8List frame) {
    if (frame.length < 5) {
      throw new TProtocolError(TProtocolErrorType.INVALID_DATA,
          "invalid frame size ${frame.length}");
    }

    _checkVersion(frame);

    return _readPairs(frame, 5, _readInt(frame, 1) + 5);
  }

  static Map<String, String> _readPairs(Uint8List buff, int start, int end) {
    Map<String, String> headers = {};
    for (var i = start; i < end; i) {
      // Read header name
      var nameSize = _readInt(buff, i);
      i += 4;
      if (i > end || i + nameSize > end) {
        throw new TProtocolError(
            TProtocolErrorType.INVALID_DATA, "invalid protocol header name");
      }
      var name = _decoder.convert(buff, i, i + nameSize);
      i += nameSize;

      // Read header value
      var valueSize = _readInt(buff, i);
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

  static int _readInt(Uint8List buff, int i) {
    return ((buff[i] & 0xff) << 24) |
        ((buff[i + 1] & 0xff) << 16) |
        ((buff[i + 2] & 0xff) << 8) |
        (buff[i + 3] & 0xff);
  }

  static void _writeInt(int i, Uint8List buff, int i1) {
    buff[i1] = (0xff & (i >> 24));
    buff[i1 + 1] = (0xff & (i >> 16));
    buff[i1 + 2] = (0xff & (i >> 8));
    buff[i1 + 3] = (0xff & (i));
  }

  static void _writeStringBytes(List<int> strBytes, Uint8List buff, int i) {
    buff.setRange(i, i + strBytes.length, strBytes);
  }

  // Evaluates the version and throws a TProtocolError if the version is unsupported
  // Support more versions when available
  static void _checkVersion(Uint8List frame) {
    if (frame[0] != _v0) {
      throw new TProtocolError(TProtocolErrorType.BAD_VERSION,
          "unsupported header version ${frame[0]}");
    }
  }
}
