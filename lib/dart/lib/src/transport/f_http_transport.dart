part of frugal;

/// FHttpClientTransport is a client FTransport that makes frugal requests via http.
class FHttpClientTransport extends FTransport {
  static const int UNAUTHORIZED = 401;
  static const int REQUEST_ENTITY_TOO_LARGE = 413;

  final Logger log = new Logger('FHttpTransport');
  final List<int> _writeBuffer = [];
  final wt.Client client;
  final FHttpConfig config;

  FHttpClientTransport(this.client, this.config) {}

  @override
  bool get isOpen => true;

  @override
  Future open() => new Future.value();

  @override
  Future close() => new Future.value();

  @override
  void write(Uint8List buffer, int offset, int length) {
    if (buffer == null) {
      throw new ArgumentError.notNull('buffer');
    }

    if (offset + length > buffer.length) {
      throw new ArgumentError('The range exceeds the buffer length');
    }

    if (config.requestSizeLimit > 0 &&
        length + _writeBuffer.length > config.requestSizeLimit) {
      throw new FMessageSizeError.request();
    }

    _writeBuffer.addAll(buffer.sublist(offset, offset + length));
  }

  Future _flushData(Uint8List bytes) async {
    // Encode request payload
    var requestBody = BASE64.encode(bytes);

    // Configure the request
    wt.Request request = client.newRequest()
      ..headers = config.headers
      ..uri = config.uri
      ..body = requestBody;

    // Attempt the request
    wt.Response response;
    try {
      response = await request.post();
    } on StateError catch (ex) {
      throw new TTransportError(
          TTransportErrorType.UNKNOWN, 'Malformed request ${ex.toString()}');
    } on wt.RequestException catch (ex) {
      if (ex.response == null) {
        throw new TTransportError(TTransportErrorType.UNKNOWN, ex.message);
      }
      if (ex.response.status == UNAUTHORIZED) {
        throw new TTransportError(TTransportErrorType.UNKNOWN,
            'Frugal http request failed - unauthorized ${ex.message}');
      }
      if (ex.response.status == REQUEST_ENTITY_TOO_LARGE) {
        throw new FMessageSizeError.response();
      }
      throw new TTransportError(TTransportErrorType.UNKNOWN, ex.message);
    }

    // Attempt to decode the response payload
    Uint8List data;
    try {
      data = new Uint8List.fromList(BASE64.decode(response.body.asString()));
    } on FormatException catch (_) {
      throw new TProtocolError(TProtocolErrorType.INVALID_DATA,
          'Expected a Base 64 encoded string.');
    }

    // If not enough data, throw a protocol error
    if (data.length < 4) {
      throw new TProtocolError(
          TProtocolErrorType.INVALID_DATA, 'Expected frugal data to be framed');
    }

    // Process the request, but drop the frame size
    executeFrame(data);
  }

  @override
  Future flush() {
    // Swap out the write buffer before yielding the thread via a future,
    // otherwise other writes could get into the buffer before this one is sent.

    // Frame the request body per frugal spec
    Uint8List bytes = new Uint8List(4 + _writeBuffer.length);
    bytes.buffer.asByteData().setUint32(0, _writeBuffer.length);
    bytes.setAll(4, _writeBuffer);
    _writeBuffer.clear();

    return _flushData(bytes);
  }
}

/// FHttpConfig wraps request configuration information,
/// such as server URL and request headers.
class FHttpConfig {
  final Uri uri;

  /// Limits the size of Frugal frame size for requests to the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int requestSizeLimit;

  /// Limits the size of Frugal frame size for responses from the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int responseSizeLimit;

  Map<String, String> _headers;
  get headers => _headers;

  FHttpConfig(this.uri, Map<String, String> headers,
      {this.requestSizeLimit: 0, this.responseSizeLimit: 0}) {
    if (uri == null || !uri.hasAuthority) {
      throw new ArgumentError('Invalid uri');
    }

    _initHeaders(headers);
  }

  void _initHeaders(Map<String, String> initial) {
    var h = {};

    if (initial != null) {
      h.addAll(initial);
    }

    h['content-type'] = 'application/x-frugal';
    h['content-transfer-encoding'] = 'base64';
    h['accept'] = 'application/x-frugal';

    if (responseSizeLimit > 0) {
      h['x-frugal-payload-limit'] = responseSizeLimit.toString();
    }

    _headers = new Map.unmodifiable(h);
  }
}
