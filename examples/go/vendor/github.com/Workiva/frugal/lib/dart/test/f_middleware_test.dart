import "dart:async";

import "package:frugal/frugal.dart";
import "package:test/test.dart";

/// Test class.
class MiddlewareTestingService {
  /// A field.
  String str;

  /// Handle something.
  Future<int> handleSomething(int first, int second, String str) {
    this.str = str;
    return new Future.value(first + second);
  }
}

/// Test struct.
class MiddlewareDataStruct {
  /// An arg.
  int arg;

  /// A method name.
  String methodName;

  /// A service name.
  String serviceName;
}

void main() {
  Middleware newMiddleware(MiddlewareDataStruct mds) {
    return (InvocationHandler next) {
      return (String serviceName, String methodName, List<Object> args) {
        mds.arg = args[0];
        mds.serviceName = serviceName;
        mds.methodName = methodName;
        args[0] = mds.arg + 1;
        return next(serviceName, methodName, args);
      };
    };
  }

  test('no middleware', () async {
    MiddlewareTestingService service = new MiddlewareTestingService();
    FMethod method = new FMethod(service.handleSomething,
        'MiddlewareTestingService', 'handleSomething', null);
    expect(await method([3, 64, 'foo']), equals(67));
    expect(service.str, equals('foo'));
  });

  test('middleware', () async {
    MiddlewareDataStruct mds1 = new MiddlewareDataStruct();
    MiddlewareDataStruct mds2 = new MiddlewareDataStruct();

    MiddlewareTestingService service = new MiddlewareTestingService();
    Middleware middleware1 = newMiddleware(mds1);
    Middleware middleware2 = newMiddleware(mds2);
    FMethod method = new FMethod(
        service.handleSomething,
        'MiddlewareTestingService',
        'handleSomething',
        [middleware1, middleware2]);
    expect(await method([3, 64, 'foo']), equals(69));
    expect(mds1.arg, equals(4));
    expect(mds2.arg, equals(3));
    expect(mds1.serviceName, equals('MiddlewareTestingService'));
    expect(mds2.serviceName, equals('MiddlewareTestingService'));
    expect(mds1.methodName, equals('handleSomething'));
    expect(mds2.methodName, equals('handleSomething'));
  });

  group('msgDebugMiddleware', () {
    test('Prints method, args, and return value', () async {
      bool handlerRan = false;
      InvocationHandler handler =
          (String serviceName, String methodName, List<Object> args) {
        handlerRan = true;
        print('hello world');
      };
      await debugMiddleware(handler)(
          'Service', 'method', [new FContext(correlationId: 'cid'), 1]);
      // It would be nice to expect that print ...hello world... was called, but that does not seem possible
      // Next best thing is to just see that the handler was called without throwing an error
      expect(handlerRan, true);
    });
  });
}
