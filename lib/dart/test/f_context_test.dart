import "package:frugal/frugal.dart";
import "package:test/test.dart";

void main() {
  test("FContext.withRequestHeaders", () {
    var context = new FContext.withRequestHeaders({"Something": "Value"});
    expect(context.timeout, isNotNull);
    expect(context.correlationId(), isNotNull);
    expect(context.requestHeaders()["_opid"], equals("0"));
  });

  test("FContext.requestHeaders", () {
    var context = new FContext.withRequestHeaders({"Something": "Value"});
    context.addRequestHeader("foo", "bar");
    expect(context.requestHeader("Something"), equals("Value"));
    expect(context.requestHeader("foo"), equals("bar"));
    var headers = context.requestHeaders();
    expect(headers["Something"], equals("Value"));
    expect(headers["foo"], equals("bar"));
  });

  test("FContext.responseHeaders", () {
    var context = new FContext();
    context.addResponseHeader("foo", "bar");
    expect(context.responseHeader("foo"), equals("bar"));
    var headers = context.responseHeaders();
    expect(headers["foo"], equals("bar"));
  });
}
