import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:frugal/frugal.dart';
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';
import 'package:mockito/mockito.dart';

void main() {
  group('FTransport', () {
    int requestSizeLimit = 5;
    _FTransportImpl transport;

    setUp(() {
      transport = new _FTransportImpl(requestSizeLimit);
    });

    test(
        'test closeWithException adds the exeption to the onClose stream and properly triggers the transport monitor',
        () async {
      var monitor = new MockTransportMonitor();
      transport.monitor = monitor;
      transport.errors = [null, new TError(0, 'reopen failed'), null];

      var completer = new Completer<Error>();
      var err = new TTransportError();
      transport.onClose.listen((e) {
        completer.complete(e);
      });

      // Open the transport
      await transport.open();

      // Close the transport with an error
      when(monitor.onClosedUncleanly(any)).thenReturn(1);
      when(monitor.onReopenFailed(any, any)).thenReturn(1);
      await transport.close(err);

      var timeout = new Duration(seconds: 1);
      expect(await completer.future.timeout(timeout), equals(err));
      verify(monitor.onClosedUncleanly(err)).called(1);
      verify(monitor.onReopenFailed(1, 1)).called(1);
      verify(monitor.onReopenSucceeded()).called(1);
    });
  });
}

class _FTransportImpl extends FTransport {
  // Default implementations of non-implemented methods
  List<Error> errors = [];
  int openCalls = 0;

  _FTransportImpl(int requestSizeLimit)
      : super(requestSizeLimit: requestSizeLimit);

  @override
  Future<Null> oneway(FContext ctx, Uint8List payload) => new Future.value();

  @override
  Future<TTransport> request(FContext ctx, Uint8List payload) =>
      new Future.value();

  @override
  Future open() async {
    if (openCalls <= errors.length) {
      if (errors[openCalls] != null) {
        openCalls++;
        throw errors[openCalls];
      }
    }
    openCalls++;
  }

  @override
  bool get isOpen => false;
}

/// Mock transport monitor.
class MockTransportMonitor extends Mock implements FTransportMonitor {}
