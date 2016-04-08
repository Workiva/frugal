import "package:frugal/frugal.dart";
import "package:test/test.dart";

void main() {
  test("FContext.withRequestHeaders", () {
    var context = new FContext.withRequestHeaders({"Something": "Value"});
    expect(context.timeout, isNotNull);
    expect(context.correlationId(), isNotNull);
    expect(context.requestHeaders()["_opid"], equals("0"));
  });
}
