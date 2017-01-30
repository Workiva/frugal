import "dart:typed_data";
import "package:test/test.dart";
import "package:thrift/thrift.dart";

import "../../lib/src/frugal.dart";

var _headers = {"foo": "bar", "blah": "baz"};
var _list = [
  0,
  0,
  0,
  0,
  29,
  0,
  0,
  0,
  3,
  102,
  111,
  111,
  0,
  0,
  0,
  3,
  98,
  97,
  114,
  0,
  0,
  0,
  4,
  98,
  108,
  97,
  104,
  0,
  0,
  0,
  3,
  98,
  97,
  122
];

void main() {
  test("read reads the headers out of the transport", () {
    var encodedHeaders = new Uint8List.fromList(_list);
    var transport = new TMemoryTransport.fromUint8List(encodedHeaders);
    var decodedHeaders = Headers.read(transport);
    expect(decodedHeaders, _headers);
  });

  test("read throws error for unsupported version", () {
    var encodedHeaders = new Uint8List.fromList([0x01]);
    var transport = new TMemoryTransport.fromUint8List(encodedHeaders);
    expect(() => Headers.read(transport),
        throwsA(new isInstanceOf<TProtocolError>()));
  });

  test("decodeFromFrame decodes headers from a fixed frame", () {
    var encodedHeaders = new Uint8List.fromList(_list);
    var decodedHeaders = Headers.decodeFromFrame(encodedHeaders);
    expect(decodedHeaders, _headers);
  });

  test("encode encodes headers and decodeFromFrame decodes them", () {
    var encodedHeaders = Headers.encode(_headers);
    var decodedHeaders = Headers.decodeFromFrame(encodedHeaders);
    expect(decodedHeaders, _headers);
  });

  test("encode encodes null headers and decodeFromFrame decodes them", () {
    var encodedHeaders = Headers.encode(null);
    var decodedHeaders = Headers.decodeFromFrame(encodedHeaders);
    expect(decodedHeaders, {});
  });

  test("encode encodes empty headers and decodeFromFrame decodes them", () {
    var encodedHeaders = Headers.encode({});
    var decodedHeaders = Headers.decodeFromFrame(encodedHeaders);
    expect(decodedHeaders, {});
  });

  test('encode encodes utf-8 headers and decodeFromFrame decodes them', () {
    Map<String, String> headers = {
      'Đ¥ÑØ': 'δάüΓ',
      'good\u00F1ight': 'moo\u00F1',
    };
    var encodedHeaders = Headers.encode(headers);
    var decodedHeaders = Headers.decodeFromFrame(encodedHeaders);
    expect(headers, decodedHeaders);
  });

  test("decodFromFrame throws error for bad frame", () {
    expect(() => Headers.decodeFromFrame(new Uint8List(3)),
        throwsA(new isInstanceOf<TProtocolError>()));
  });

  test("decodeHeadersFromeFrame throws error for unsupported version", () {
    var encodedHeaders = new Uint8List.fromList([0x01, 0, 0, 0, 0]);
    expect(() => Headers.decodeFromFrame(encodedHeaders),
        throwsA(new isInstanceOf<TProtocolError>()));
  });
}
