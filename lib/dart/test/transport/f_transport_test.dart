import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:frugal/frugal.dart';
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';
import 'package:mockito/mockito.dart';

void main() {
  group('FTransport', () {
    MockFRegistry registry;
    FTransport transport;
    FContext context;

    setUp(() {
      transport = new FTransportImpl();
      registry = new MockFRegistry();
      context = new FContext();
    });

    test('test setting null registry throws FError', () async {
      expect(() => transport.setRegistry(null),
          throwsA(new isInstanceOf<FError>()));
    });

    test('test registering/unregistering beforw setting registry throws FError',
        () async {
      expect(() => transport.register(context, callback),
          throwsA(new isInstanceOf<FError>()));
      expect(() => transport.unregister(context),
          throwsA(new isInstanceOf<FError>()));
    });

    test('test register/unregister call through to registry', () async {
      transport.setRegistry(registry);
      expect(registry.context, isNull);
      expect(registry.callback, isNull);
      transport.register(context, callback);
      expect(registry.context, equals(context));
      expect(registry.callback, equals(callback));
      transport.unregister(context);
      expect(registry.context, isNull);
      expect(registry.callback, isNull);
    });

    test(
        'test closeWithException add the exeption to the onClose stream '
        'and properly triggers the transport monitor', () async {
      var monitor = new MockTransportMonitor();
      transport.monitor = monitor;

      var completer = new Completer<Error>();
      var err = new TTransportError();
      transport.onClose.listen((e) {
        completer.complete(e);
      });

      when(monitor.onClosedUncleanly(any)).thenReturn(-1);
      await transport.closeWithException(err);

      var timeout = new Duration(seconds: 1);
      expect(await completer.future.timeout(timeout), equals(err));
      verify(monitor.onClosedUncleanly(err)).called(1);
    });
  });
}

void callback(TTransport transport) {
  return;
}

class FTransportImpl extends FTransport {}

class MockFRegistry extends FRegistry {
  List<Uint8List> data;
  FContext context;
  FAsyncCallback callback;
  Completer executeCompleter;
  Error executeError;

  MockFRegistry() {
    data = new List();
  }

  void initCompleter() {
    executeCompleter = new Completer();
  }

  void register(FContext ctx, FAsyncCallback callback) {
    this.context = ctx;
    this.callback = callback;
  }

  void unregister(FContext ctx) {
    if (this.context == ctx) {
      this.context = null;
      this.callback = null;
    }
  }

  void execute(Uint8List data) {
    this.data.add(data);
    if (executeCompleter != null && !executeCompleter.isCompleted) {
      executeCompleter.complete();
    }
    if (executeError != null) {
      throw executeError;
    }
  }
}

class MockTransportMonitor extends Mock implements FTransportMonitor {
  noSuchMethod(i) => super.noSuchMethod(i);
}
