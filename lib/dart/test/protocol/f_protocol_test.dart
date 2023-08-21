import 'package:frugal/frugal.dart';
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';

void main() {
  test("writeRequestHeader writes the request headers and readRequestHeader reads the headers", () {
    var transport = TMemoryTransport();
    var tProtocol = TBinaryProtocol(transport);
    var fProtocol = FProtocol(tProtocol);

    var ctx = FContext(correlationId: "sweet-corid");
    ctx.addRequestHeader("foo", "bar");
    fProtocol.writeRequestHeader(ctx);

    var decodedCtx = fProtocol.readRequestHeader();
    expect(decodedCtx.requestHeaders(), ctx.requestHeaders());
  });
}
