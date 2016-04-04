part of frugal;

String _cid = "_cid";
String _opid = "_opid";

/// Context for Frugal request.
class FContext {
  Map<String, String> _requestHeaders;
  Map<String, String> _responseHeaders;

  // Default timeout to 1 minute
  Duration _timeout = new Duration(minutes: 1);

  Duration get timeout => _timeout;
  void set timeout(timeout) {
    _timeout = timeout;
  }

  FContext({String correlationId: ""}) {
    if (correlationId == "") {
      correlationId = _generateCorrelationId();
    }
    _requestHeaders = {_cid: correlationId, _opid: "0",};
    _responseHeaders = {};
  }

  FContext.withRequestHeaders(Map<String, String> headers) {
    if (!headers.containsKey(_cid) || headers[_cid] == "") {
      headers[_cid] = _generateCorrelationId();
    }
    if (!headers.containsKey(_opid) || headers[_opid] == "") {
      headers[_opid] = "0";
    }
    _requestHeaders = headers;
    _responseHeaders = {};
  }

  /// Correlation Id for the context
  String correlationId() => _requestHeaders[_cid];

  /// Get the operation id for the context
  int _opId() {
    var opIdStr = _requestHeaders[_opid];
    return int.parse(opIdStr);
  }

  /// Set the operation id for the context
  void _setOpId(int id) {
    _requestHeaders[_opid] = "$id";
  }

  /// Add a request header to the context for the given name.
  /// Will overwrite existing header of the same name.
  void addRequestHeader(String name, value) {
    _requestHeaders[name] = value;
  }

  /// Add given request headers to the context. Will overwrite
  /// existing pre-existing headers with the same names as the
  /// given headers.
  void addRequestsHeaders(Map<String, String> headers) {
    if (headers == null || headers.length == 0) {
      return;
    }
    for (var name in headers.keys) {
      _requestHeaders[name] = headers[name];
    }
  }

  /// Get the named request header
  String requestHeader(String name) {
    return _requestHeaders[name];
  }

  /// Get requests headers map
  Map<String, String> requestHeaders() {
    return new UnmodifiableMapView(_requestHeaders);
  }

  /// Add a response header to the context for the given name
  /// Will overwrite existing header of the same name.
  void addResponseHeader(String name, value) {
    _responseHeaders[name] = value;
  }

  /// Add given response headers to the context. Will overwrite
  /// existing pre-existing headers with the same names as the
  /// given headers.
  void addResponseHeaders(Map<String, String> headers) {
    if (headers == null || headers.length == 0) {
      return;
    }
    for (var name in headers.keys) {
      _responseHeaders[name] = headers[name];
    }
  }

  /// Get the named response header
  String responseHeader(String name) {
    return _requestHeaders[name];
  }

  /// Get response headers map
  Map<String, String> responseHeaders() {
    return new UnmodifiableMapView(_responseHeaders);
  }

  static String _generateCorrelationId() =>
      new Uuid().v4().toString().replaceAll('-', '');
}
