import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:frugal/frugal.dart';
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';
import 'package:mockito/mockito.dart';
import 'f_transport_test.dart' show MockTransportMonitor;

Uint8List mockFrame(FContext ctx, String message) {
  TMemoryOutputBuffer trans = new TMemoryOutputBuffer();
  FProtocol prot = new FProtocol(new TBinaryProtocol(trans));
  prot.writeRequestHeader(ctx);
  prot.writeString(message);
  return trans.writeBytes;
}

void main() {
  group('FAdapterTransport', () {
    StreamController<TSocketState> stateStream;
    StreamController<Object> errorStream;
    StreamController<Uint8List> messageStream;
    MockSocket socket;
    MockSocketTransport socketTransport;
    FAdapterTransport transport;

    setUp(() {
      stateStream = new StreamController.broadcast();
      errorStream = new StreamController.broadcast();
      messageStream = new StreamController.broadcast();

      socket = new MockSocket();
      when(socket.onState).thenReturn(stateStream.stream);
      when(socket.onError).thenReturn(errorStream.stream);
      when(socket.onMessage).thenReturn(messageStream.stream);
      socketTransport = new MockSocketTransport();
      when(socketTransport.socket).thenReturn(socket);
      transport = new FAdapterTransport(socketTransport);
    });

    tearDown(() {
      stateStream.close();
      errorStream.close();
      messageStream.close();
    });

    test('oneway happy path', () async {
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      verify(socket.open()).called(1);

      FContext reqCtx = new FContext();
      var frame = mockFrame(reqCtx, "request");

      await transport.oneway(reqCtx, frame);
      verify(socket.send(frame)).called(1);
    });

    test('requests happy path', () async {
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      verify(socket.open()).called(1);

      FContext reqCtx = new FContext();
      var frame = mockFrame(reqCtx, "request");

      var respFrame = mockFrame(reqCtx, "response");

      when(socket.send(frame)).thenAnswer((_) {
        messageStream.add(respFrame);
      });

      var response = await transport.request(reqCtx, frame) as TMemoryTransport;
      expect(response.buffer, respFrame.sublist(4));
      verify(socket.send(frame)).called(1);
    });

    test('request transport not open', () async {
      try {
        await transport.request(null, null);
        fail('Should have thrown an exception');
      } on TTransportError catch (e) {
        expect(e.type, FrugalTTransportErrorType.NOT_OPEN);
      }
    });

    test('requests time out without a response', () async {
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      verify(socket.open()).called(1);

      FContext ctx = new FContext();
      ctx.timeout = new Duration(milliseconds: 50);
      var frame = mockFrame(ctx, 'request');

      try {
        await transport.request(ctx, frame);
        fail('Should have thrown an exception');
      } on TTransportError catch (e) {
        expect(e.type, FrugalTTransportErrorType.TIMED_OUT);
      }

      verify(socket.send(frame)).called(1);
    });

    test('request is cancelled if the transport is closed', () async {
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      verify(socket.open()).called(1);

      FContext ctx = new FContext();
      var frame = mockFrame(ctx, 'request');
      Future<TTransport> requestFuture = transport.request(ctx, frame);
      await transport.close();

      try {
        await requestFuture;
        fail('Should have thrown an exception');
      } on TTransportError catch (e) {
        expect(e.type, FrugalTTransportErrorType.NOT_OPEN);
      }
    });

    test('test socket error triggers transport close', () async {
      // Open the transport
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      var monitor = new MockTransportMonitor();
      transport.monitor = monitor;
      expect(transport.isOpen, equals(true));

      // Kill the socket with an error
      var err = new StateError('');
      var closeCompleter = new Completer();
      transport.onClose.listen((e) {
        closeCompleter.complete(e);
      });
      var monitorCompleter = new Completer();
      when(monitor.onClosedUncleanly(any))
          .thenAnswer((Invocation realInvocation) {
        monitorCompleter.complete(realInvocation.positionalArguments[0]);
        return -1;
      });
      errorStream.add(err);
      var timeout = new Duration(seconds: 1);
      expect(await closeCompleter.future.timeout(timeout), equals(err));
      expect(await monitorCompleter.future.timeout(timeout), equals(err));
      expect(transport.isOpen, equals(false));

      // Reopen the socket under the hood
      stateStream.add(TSocketState.OPEN);
      monitorCompleter = new Completer();
      when(monitor.onReopenSucceeded()).thenAnswer((Invocation realInvocation) {
        monitorCompleter.complete();
      });
      await monitorCompleter.future.timeout(timeout);
      expect(transport.isOpen, equals(true));
    });
  });
}

/// Mock socket transport.
class MockSocketTransport extends Mock implements TSocketTransport {}

/// Mock socket.
class MockSocket extends Mock implements TSocket {}
