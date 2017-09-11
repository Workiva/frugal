import 'dart:async';
import 'dart:convert' show BASE64;
import 'dart:convert' show Utf8Codec;
import 'dart:typed_data' show Uint8List;

import 'package:frugal/frugal.dart';
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';
import 'package:w_transport/w_transport.dart';
import 'package:w_transport/w_transport_mock.dart';

void main() {
  configureWTransportForTest();
  const utf8Codec = const Utf8Codec();

  group('FHttpTransport', () {
    Client client;
    FHttpTransport transport;

    Map<String, String> expectedRequestHeaders = {
      'x-frugal-payload-limit': '10',
      // TODO: When w_transport supports content-type overrides, enable this.
      // 'content-type': 'application/x-frugal',
      'content-transfer-encoding': 'base64',
      'accept': 'application/x-frugal',
      'foo': 'bar'
    };
    Map<String, String> responseHeaders = {
      'content-type': 'application/x-frugal',
      'content-transfer-encoding': 'base64'
    };
    Uint8List transportRequest =
        new Uint8List.fromList([0, 0, 0, 5, 1, 2, 3, 4, 5]);
    String transportRequestB64 = BASE64.encode(transportRequest);
    Uint8List transportResponse = new Uint8List.fromList([6, 7, 8, 9]);
    Uint8List transportResponseFramed =
        new Uint8List.fromList([0, 0, 0, 4, 6, 7, 8, 9]);
    String transportResponseB64 = BASE64.encode(transportResponseFramed);

    setUp(() {
      client = new Client();
      transport = new FHttpTransport(client, Uri.parse('http://localhost'),
          responseSizeLimit: 10, additionalHeaders: {'foo': 'bar'});
    });

    test('Test transport sends body and receives response', () async {
      MockTransports.http.when(transport.uri, (FinalizedRequest request) async {
        if (request.method == 'POST') {
          HttpBody body = request.body;
          if (body == null || body.asString() != transportRequestB64)
            return new MockResponse.badRequest();
          for (var key in expectedRequestHeaders.keys) {
            if (request.headers[key] != expectedRequestHeaders[key]) {
              return new MockResponse.badRequest();
            }
          }
          return new MockResponse.ok(
              body: transportResponseB64, headers: responseHeaders);
        } else {
          return new MockResponse.badRequest();
        }
      });

      var response = await transport.request(new FContext(), transportRequest)
          as TMemoryTransport;
      expect(response.buffer, transportResponse);
    });

    test('Transport times out if request is not received within the timeout',
        () async {
      MockTransports.http.when(transport.uri, (FinalizedRequest request) async {
        if (request.method == 'POST') {
          await new Future.delayed(new Duration(milliseconds: 100));
        }
      });

      try {
        FContext ctx = new FContext()..timeout = new Duration(milliseconds: 20);
        await transport.request(ctx, transportRequest);
        fail('should have thrown an exception');
      } on TTransportError catch (e) {
        expect(e.type, FrugalTTransportErrorType.TIMED_OUT);
      }
    });

    test('Multiple writes are not coalesced', () async {
      MockTransports.http.when(transport.uri, (FinalizedRequest request) async {
        if (request.method == 'POST') {
          HttpBody body = request.body;
          if (body == null || body.asString() != transportRequestB64)
            return new MockResponse.badRequest();
          for (var key in expectedRequestHeaders.keys) {
            if (request.headers[key] != expectedRequestHeaders[key]) {
              return new MockResponse.badRequest();
            }
          }
          return new MockResponse.ok(
              body: transportResponseB64, headers: responseHeaders);
        } else {
          return new MockResponse.badRequest();
        }
      });

      var first = transport.request(new FContext(), transportRequest);
      var second = transport.request(new FContext(), transportRequest);

      var firstResponse = (await first) as TMemoryTransport;
      var secondResponse = (await second) as TMemoryTransport;

      expect(firstResponse.buffer, transportResponse);
      expect(secondResponse.buffer, transportResponse);
    });

    test('Test transport does not execute frame on oneway requests', () async {
      Uint8List responseBytes = new Uint8List.fromList([0, 0, 0, 0]);
      Response response =
          new MockResponse.ok(body: BASE64.encode(responseBytes));
      MockTransports.http.expect('POST', transport.uri, respondWith: response);
      var result = await transport.request(new FContext(), transportRequest);
      expect(result, null);
    });

    test('Test transport throws TransportError on bad oneway requests',
        () async {
      Uint8List responseBytes = new Uint8List.fromList([0, 0, 0, 1]);
      Response response =
          new MockResponse.ok(body: BASE64.encode(responseBytes));
      MockTransports.http.expect('POST', transport.uri, respondWith: response);
      expect(transport.request(new FContext(), transportRequest),
          throwsA(new isInstanceOf<TTransportError>()));
    });

    test('Test transport receives non-base64 payload', () async {
      Response response = new MockResponse.ok(body: '`');
      MockTransports.http.expect('POST', transport.uri, respondWith: response);
      expect(transport.request(new FContext(), transportRequest),
          throwsA(new isInstanceOf<TProtocolError>()));
    });

    test('Test transport receives unframed frugal payload', () async {
      Response response = new MockResponse.ok();
      MockTransports.http.expect('POST', transport.uri, respondWith: response);
      expect(transport.request(new FContext(), transportRequest),
          throwsA(new isInstanceOf<TProtocolError>()));
    });
  });

  group('FHttpTransport request size too large', () {
    Client client;
    FHttpTransport transport;

    setUp(() {
      client = new Client();
      transport = new FHttpTransport(client, Uri.parse('http://localhost'),
          requestSizeLimit: 10);
    });

    test('Test transport receives error', () {
      expect(
          transport.request(
              new FContext(), utf8Codec.encode('my really long request')),
          throwsA(new isInstanceOf<TTransportError>()));
    });
  });

  group('FHttpTransport http post failed', () {
    FHttpTransport transport;

    setUp(() {
      transport =
          new FHttpTransport(new Client(), Uri.parse('http://localhost'));
    });

    test('Test transport receives error on 401 response', () async {
      Response response = new MockResponse.unauthorized();
      MockTransports.http.expect('POST', transport.uri, respondWith: response);
      expect(transport.request(new FContext(), utf8Codec.encode('my request')),
          throwsA(new isInstanceOf<TTransportError>()));
    });

    test('Test transport receives response too large error on 413 response',
        () async {
      Response response =
          new MockResponse(FHttpTransport.REQUEST_ENTITY_TOO_LARGE);
      MockTransports.http.expect('POST', transport.uri, respondWith: response);
      expect(transport.request(new FContext(), utf8Codec.encode('my request')),
          throwsA(new isInstanceOf<TTransportError>()));
    });

    test('Test transport receives error on 404 response', () async {
      Response response = new MockResponse.badRequest();
      MockTransports.http.expect('POST', transport.uri, respondWith: response);
      expect(transport.request(new FContext(), utf8Codec.encode('my request')),
          throwsA(new isInstanceOf<TTransportError>()));
    });

    test('Test transport receives error on no response', () async {
      Response response = new MockResponse.badRequest();
      MockTransports.http.expect('POST', transport.uri, respondWith: response);
      expect(transport.request(new FContext(), utf8Codec.encode('my request')),
          throwsA(new isInstanceOf<TTransportError>()));
    });
  });
}
