import 'dart:async';
import 'dart:convert';
import 'dart:typed_data' show Uint8List;

import 'package:frugal/frugal.dart';
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';
import 'package:w_transport/mock.dart';
import 'package:w_transport/w_transport.dart';

void main() {
  configureWTransportForTest();
  const utf8Codec = const Utf8Codec();

  group('FHttpTransport', () {
    HttpClient client;
    FHttpTransport? transport;
    FHttpTransport? transportWithContext;

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
    String transportRequestB64 = base64.encode(transportRequest);
    Uint8List transportResponse = new Uint8List.fromList([6, 7, 8, 9]);
    Uint8List transportResponseFramed =
        new Uint8List.fromList([0, 0, 0, 4, 6, 7, 8, 9]);
    String transportResponseB64 = base64.encode(transportResponseFramed);

    setUp(() {
      client = new HttpClient();
      transport = new FHttpTransport(client, Uri.parse('http://localhost'),
          responseSizeLimit: 10, additionalHeaders: {'foo': 'bar'});
      transportWithContext = new FHttpTransport(
          client, Uri.parse('http://localhost'),
          responseSizeLimit: 10,
          additionalHeaders: {'foo': 'bar'},
          getRequestHeaders: _generateTestHeader);
    });

    test('Test transport sends body and receives response', () async {
      MockTransports.http.when(transport!.uri, (FinalizedRequest request) async {
        if (request.method == 'POST') {
          HttpBody body = request.body as HttpBody;
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

      var response = await transport!.request(new FContext(), transportRequest)
          as TMemoryTransport;
      expect(response.buffer, transportResponse);
    });

    test('Transport times out if request is not received within the timeout',
        () async {
      MockTransports.http.when(transport!.uri, (FinalizedRequest request) async {
        if (request.method == 'POST') {
          throw new TimeoutException("wat");
        }
        return new MockResponse.badRequest();
      });

      try {
        FContext ctx = new FContext()..timeout = new Duration(milliseconds: 20);
        await transport!.request(ctx, transportRequest);
        fail('should have thrown an exception');
      } on TTransportError catch (e) {
        expect(e.type, FrugalTTransportErrorType.TIMED_OUT);
      }
    });

    test('Multiple writes are not coalesced', () async {
      MockTransports.http.when(transport!.uri, (FinalizedRequest request) async {
        if (request.method == 'POST') {
          HttpBody body = request.body as HttpBody;
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

      var first = transport!.request(new FContext(), transportRequest);
      var second = transport!.request(new FContext(), transportRequest);

      var firstResponse = (await first) as TMemoryTransport;
      var secondResponse = (await second) as TMemoryTransport;

      expect(firstResponse.buffer, transportResponse);
      expect(secondResponse.buffer, transportResponse);
    });

    test(
        'Test transport sends body and receives response with FContext function',
        () async {
      FContext newContext = new FContext();
      Map<String, String> tempExpectedHeaders = expectedRequestHeaders;
      tempExpectedHeaders['first-header'] ??= newContext.correlationId ?? '';
      tempExpectedHeaders['second-header'] = 'yup';

      MockTransports.http.when(transportWithContext!.uri,
          (FinalizedRequest request) async {
        if (request.method == 'POST') {
          HttpBody body = request.body as HttpBody;
          if (body == null || body.asString() != transportRequestB64)
            return new MockResponse.badRequest();
          for (var key in tempExpectedHeaders.keys) {
            if (request.headers[key] != tempExpectedHeaders[key]) {
              return new MockResponse.badRequest();
            }
          }
          return new MockResponse.ok(
              body: transportResponseB64, headers: responseHeaders);
        } else {
          return new MockResponse.badRequest();
        }
      });

      var response = await transportWithContext!.request(
          newContext, transportRequest) as TMemoryTransport;
      expect(response.buffer, transportResponse);
    });

    test('Test transport does not execute frame on oneway requests', () async {
      Uint8List responseBytes = new Uint8List.fromList([0, 0, 0, 0]);
      Response response =
          new MockResponse.ok(body: base64.encode(responseBytes));
      MockTransports.http.expect('POST', transport!.uri, respondWith: response);
      var result = await transport!.request(new FContext(), transportRequest);
      expect(result, null);
    });

    test('Test transport throws TransportError on bad oneway requests',
        () async {
      Uint8List responseBytes = new Uint8List.fromList([0, 0, 0, 1]);
      Response response =
          new MockResponse.ok(body: base64.encode(responseBytes));
      MockTransports.http.expect('POST', transport!.uri, respondWith: response);
      expect(transport!.request(new FContext(), transportRequest),
          throwsA(new isInstanceOf<TTransportError>()));
    });

    test('Test transport receives non-base64 payload', () async {
      Response response = new MockResponse.ok(body: '`');
      MockTransports.http.expect('POST', transport!.uri, respondWith: response);
      expect(transport!.request(new FContext(), transportRequest),
          throwsA(new isInstanceOf<TProtocolError>()));
    });

    test('Test transport receives unframed frugal payload', () async {
      Response response = new MockResponse.ok();
      MockTransports.http.expect('POST', transport!.uri, respondWith: response);
      expect(transport!.request(new FContext(), transportRequest),
          throwsA(new isInstanceOf<TProtocolError>()));
    });
  });

  group('FHttpTransport request size too large', () {
    HttpClient client;
    FHttpTransport? transport;

    setUp(() {
      client = new HttpClient();
      transport = new FHttpTransport(client, Uri.parse('http://localhost'),
          requestSizeLimit: 10);
    });

    test('Test transport receives error', () {

      List<int> requestData = utf8Codec.encode('my really long request');
      Uint8List requestDataUint8 = Uint8List.fromList(requestData);

      expect(
          transport?.request(new FContext(), requestDataUint8),
          throwsA(new isInstanceOf<TTransportError>()));
    });
  });

  group('FHttpTransport http post failed', () {
    FHttpTransport? transport;

    setUp(() {
      transport =
          new FHttpTransport(new HttpClient(), Uri.parse('http://localhost'));
    });

    test('Test transport receives error on 401 response', () async {
      Response response = new MockResponse.unauthorized();
      MockTransports.http.expect('POST', transport!.uri, respondWith: response);
      List<int> requestData = utf8Codec.encode('my request');
      Uint8List requestDataUint8 = Uint8List.fromList(requestData);
      expect(transport!.request(new FContext(), requestDataUint8),
          throwsA(new isInstanceOf<TTransportError>()));
    });

    test('Test transport receives response too large error on 413 response',
        () async {
      Response response =
          new MockResponse(FHttpTransport.REQUEST_ENTITY_TOO_LARGE);
      MockTransports.http.expect('POST', transport!.uri, respondWith: response);
      List<int> requestData = utf8Codec.encode('my request');
      Uint8List requestDataUint8 = Uint8List.fromList(requestData);
      expect(transport!.request(new FContext(), requestDataUint8),
          throwsA(new isInstanceOf<TTransportError>()));
    });

    test('Test transport receives error on 404 response', () async {
      Response response = new MockResponse.badRequest();
      MockTransports.http.expect('POST', transport!.uri, respondWith: response);
      List<int> requestData = utf8Codec.encode('my request');
      Uint8List requestDataUint8 = Uint8List.fromList(requestData);
      expect(transport!.request(new FContext(), requestDataUint8),
          throwsA(new isInstanceOf<TTransportError>()));
    });

    test('Test transport receives error on no response', () async {
      Response response = new MockResponse.badRequest();
      MockTransports.http.expect('POST', transport!.uri, respondWith: response);
      List<int> requestData = utf8Codec.encode('my request');
      Uint8List requestDataUint8 = Uint8List.fromList(requestData);
      expect(transport!.request(new FContext(), requestDataUint8),
          throwsA(new isInstanceOf<TTransportError>()));
    });
  });
}

Map<String, String> _generateTestHeader(FContext? ctx) {
  return {
    "first-header": ctx!.correlationId ?? '',
    "second-header": "yup",
    "x-frugal-payload-limit": "these headers",
    "content-transfer-encoding": "will be",
    "accept": "overwritten!"
  };
}
