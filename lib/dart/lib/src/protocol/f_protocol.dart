part of frugal;

/// Extension of TProtocol with the addition of headers
class FProtocol extends TProtocolDecorator {
  final TTransport _transport;

  FProtocol(TProtocol protocol)
      : this._transport = protocol.transport,
        super(protocol);

  /// write the request headers on the given Context
  void writeRequestHeader(FContext ctx) {
    transport.writeAll(Headers.encode(ctx.requestHeaders()));
  }

  /// read the requests headers into a new Context
  FContext readRequestHeader() {
    return new FContext.withRequestHeaders(Headers.read(transport));
  }

  /// write the response headers on the given Context
  void writeResponseHeader(FContext ctx) {
    transport.writeAll(Headers.encode(ctx.responseHeaders()));
  }

  /// read the requests headers into the given Context
  void readResponseHeader(FContext ctx) {
    ctx.addResponseHeaders(Headers.read(transport));
  }
}
