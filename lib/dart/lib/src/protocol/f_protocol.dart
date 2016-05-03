part of frugal;

/// FProtocol is Frugal's equivalent of Thrift's TProtocol. It defines the
/// serialization protocol used for messages, such as JSON, binary, etc.
/// FProtocol actually extends TProtocol and adds support for serializing
/// FContext. In practice, FProtocol simply wraps a TProtocol and uses Thrift's
/// built-in serialization. FContext is encoded before the TProtocol
/// serialization of the message using a simple binary protocol. See the
/// protocol documentation for more details.
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
