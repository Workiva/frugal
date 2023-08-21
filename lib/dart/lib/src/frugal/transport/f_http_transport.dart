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

/// typedef for passed in function that generates headers given an [FContext]
typedef GetHeadersWithContext = Map<String, String> Function(FContext ctx);

/// An [FTransport] that makes frugal requests via HTTP.
class FHttpTransport extends FTransport {
  /// HTTP status code for an unauthorized reqeuest.
  static const int UNAUTHORIZED = 401;

  /// HTTP status code for requesting too much data.
  static const int REQUEST_ENTITY_TOO_LARGE = 413;

  final Logger _log = Logger('FHttpTransport');

  /// Client used by the transport to make HTTP requests.
  final wt.HttpClient client;

  /// URI of the frugal HTTP server.
  final Uri uri;

  /// Limits the size of responses from the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int responseSizeLimit;

  Map<String, String> _headers = {};

  /// Function that accepts an FContext that should return a Map<String, String>
  /// of headers to be added to every request
  final GetHeadersWithContext _getRequestHeaders;

  /// Create an [FHttpTransport] instance with the given w_transport [Client],
  /// uri, and optional size restrictions, and headers.
  ///
  /// If specifying headers, note that the
  ///   * content-type
  ///   * content-transfer-encoding
  ///   * accept
  ///   * x-frugal-payload-limit
  /// headers will be overwritten.
  ///
  /// Additionally, a function that accepts an FContext can be passed in
  /// that should return additional headers to be appended to each request
  /// using the getRequestHeaders param.
  FHttpTransport(this.client, this.uri,
      {int requestSizeLimit = 0,
      this.responseSizeLimit = 0,
      Map<String, String>? additionalHeaders,
      GetHeadersWithContext? getRequestHeaders})
      : _getRequestHeaders = getRequestHeaders ?? ((_) => {}),
        super(requestSizeLimit: requestSizeLimit) {
    _headers = additionalHeaders ?? {};
    // add and potentially overwrite with default headers
    _headers.addAll({
      'content-type': 'application/x-frugal',
      'content-transfer-encoding': 'base64',
      'accept': 'application/x-frugal'
    });
    if (responseSizeLimit > 0) {
      _headers['x-frugal-payload-limit'] = responseSizeLimit.toString();
    }
  }

  @override
  bool get isOpen => true;

  @override
  Future open() => Future.value();

  @override
  Future close([Error? error]) => Future.value();

  @override
  Future<Null> oneway(FContext ctx, Uint8List payload) async {
    await request(ctx, payload);
  }

  @override
  Future<TTransport?> request(FContext ctx, Uint8List payload) async {
    _preflightRequestCheck(payload);

    // append dynamic headers first
    Map<String, String> requestHeaders = _getRequestHeaders(ctx);
    // add and potentially overwrite with default headers
    requestHeaders.addAll(_headers);

    // Encode request payload
    var requestBody = base64.encode(payload);

    // Configure the request
    wt.Request request = client.newRequest()
      ..headers = requestHeaders
      ..uri = uri
      ..body = requestBody
      ..timeoutThreshold = ctx.timeout;

    // Attempt the request
    wt.Response response;
    try {
      response = await request.post();
    } on StateError catch (ex) {
      throw TTransportError(FrugalTTransportErrorType.UNKNOWN, 'Malformed request ${ex.toString()}');
    } on wt.RequestException catch (ex) {
      if (ex.error != null && ex.error.runtimeType == TimeoutException) {
        throw TTransportError(FrugalTTransportErrorType.TIMED_OUT, "http request timed out after ${ctx.timeout}");
      }
      if (ex.response == null) {
        throw TTransportError(FrugalTTransportErrorType.UNKNOWN, ex.message);
      }
      if (ex.response?.status == UNAUTHORIZED) {
        throw TTransportError(
            FrugalTTransportErrorType.UNKNOWN, 'Frugal http request failed - unauthorized ${ex.message}');
      }
      if (ex.response?.status == REQUEST_ENTITY_TOO_LARGE) {
        throw TTransportError(FrugalTTransportErrorType.RESPONSE_TOO_LARGE);
      }
      throw TTransportError(FrugalTTransportErrorType.UNKNOWN, ex.message);
    }

    // Attempt to decode the response payload
    Uint8List data;
    try {
      data = Uint8List.fromList(base64.decode(response.body.asString()));
    } on FormatException catch (_) {
      throw TProtocolError(TProtocolErrorType.INVALID_DATA, 'Expected a Base 64 encoded string.');
    }

    // If not enough data, throw a protocol error
    if (data.length < 4) {
      throw TProtocolError(TProtocolErrorType.INVALID_DATA, 'Expected frugal data to be framed');
    }

    // If there are only 4 bytes, this is a one-way request
    if (data.length == 4) {
      var bData = ByteData.view(data.buffer);
      if (bData.getUint32(0) != 0) {
        throw TTransportError(FrugalTTransportErrorType.UNKNOWN, "invalid frame size");
      }
      return null;
    }

    return TMemoryTransport.fromUint8List(data.sublist(4));
  }
}
