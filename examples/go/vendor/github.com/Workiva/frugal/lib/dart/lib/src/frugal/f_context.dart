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

/// Header containing correlation id.
const String _cidHeader = "_cid";

/// Header containing op id (uint64 as string).
const String _opidHeader = "_opid";

/// Header containing request timeout (milliseconds as string).
const String _timeoutHeader = "_timeout";

/// Default request timeout duration.
final Duration _defaultTimeout = new Duration(seconds: 5);

/// The context for a Frugal message. Every RPC has an [FContext], which can be
/// used to set request headers, response headers, and the request timeout.
/// The default timeout is five seconds. An [FContext] is also sent with every
/// publish message which is then received by subscribers.
///
/// In addition to headers, the [FContext] also contains a correlation ID which
/// can be used for distributed tracing purposes. A random correlation ID is
/// generated for each [FContext] if one is not provided.
///
/// [FContext] also plays a key role in Frugal's multiplexing support. A unique,
/// per-request operation ID is set on every [FContext] upon instantiation.
/// This operation ID is sent in the request and included in the response,
/// which is then used to correlate a response to a request. The operation ID
/// is an internal implementation detail and is not exposed to the user.
///
/// An [FContext] should belong to a single request for the lifetime of that
/// request. It can be reused once the request has completed, though they
/// should generally not be reused.
class FContext {
  static int _globalOpId = 0;

  Map<String, String> _requestHeaders;
  Map<String, String> _responseHeaders;

  /// Create a new [FContext] with the optionally specified [correlationId].
  FContext({String correlationId: ""}) {
    if (correlationId == "") {
      correlationId = _generateCorrelationId();
    }
    _globalOpId++;
    _requestHeaders = {
      _cidHeader: correlationId,
      _opidHeader: _globalOpId.toString(),
      _timeoutHeader: _defaultTimeout.inMilliseconds.toString(),
    };
    _responseHeaders = {};
  }

  /// Create a new [FContext] with the given request headers.
  FContext.withRequestHeaders(Map<String, String> headers) {
    if (!headers.containsKey(_cidHeader) || headers[_cidHeader] == "") {
      headers[_cidHeader] = _generateCorrelationId();
    }
    if (!headers.containsKey(_opidHeader) || headers[_opidHeader] == "") {
      _globalOpId++;
      headers[_opidHeader] = _globalOpId.toString();
    }
    if (!headers.containsKey(_timeoutHeader) || headers[_timeoutHeader] == "") {
      headers[_timeoutHeader] = _defaultTimeout.inMilliseconds.toString();
    }
    _requestHeaders = headers;
    _responseHeaders = {};
  }

  /// The request timeout for any method call using this context.
  /// The default is 5 seconds.
  Duration get timeout {
    return new Duration(
        milliseconds: int.parse(_requestHeaders[_timeoutHeader]));
  }

  /// Set the request timeout for any method call using this context.
  set timeout(Duration timeout) {
    _requestHeaders[_timeoutHeader] = timeout.inMilliseconds.toString();
  }

  /// Correlation id for the context.
  String get correlationId => _requestHeaders[_cidHeader];

  /// The operation id for the context.
  int get _opId {
    var opIdStr = _requestHeaders[_opidHeader];
    return int.parse(opIdStr);
  }

  /// Set the operation id for the context.
  set _opId(int id) {
    _requestHeaders[_opidHeader] = "$id";
  }

  /// Add a request header to the context for the given name.
  /// Will overwrite existing header of the same name.
  void addRequestHeader(String name, String value) {
    _requestHeaders[name] = value;
  }

  /// Add given request headers to the context. Will overwrite existing
  /// pre-existing headers with the same names as the given headers.
  void addRequestsHeaders(Map<String, String> headers) {
    if (headers == null || headers.length == 0) {
      return;
    }
    for (var name in headers.keys) {
      _requestHeaders[name] = headers[name];
    }
  }

  /// Get the named request header.
  String requestHeader(String name) {
    return _requestHeaders[name];
  }

  /// Get requests headers map.
  Map<String, String> requestHeaders() {
    return new UnmodifiableMapView(_requestHeaders);
  }

  /// Add a response header to the context for the given name
  /// Will overwrite existing header of the same name.
  void addResponseHeader(String name, String value) {
    _responseHeaders[name] = value;
  }

  /// Add given response headers to the context. Will overwrite existing
  /// pre-existing headers with the same names as the given headers.
  void addResponseHeaders(Map<String, String> headers) {
    if (headers == null || headers.length == 0) {
      return;
    }
    for (var name in headers.keys) {
      _responseHeaders[name] = headers[name];
    }
  }

  /// Get the named response header.
  String responseHeader(String name) {
    return _responseHeaders[name];
  }

  /// Get response headers map.
  Map<String, String> responseHeaders() {
    return new UnmodifiableMapView(_responseHeaders);
  }

  static String _generateCorrelationId() =>
      new Uuid().v4().toString().replaceAll('-', '');
}
