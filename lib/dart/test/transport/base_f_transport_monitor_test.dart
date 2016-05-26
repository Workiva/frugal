import "dart:async";
import "package:test/test.dart";
import "package:frugal/frugal.dart";

void main() {
  test('onClosedUncleanly should return -1 if max attempts is 0', () {
    FTransportMonitor monitor = new BaseFTransportMonitor(
        maxReopenAttempts: 0, initialWait: 0, maxWait: 0);
    expect(-1, monitor.onClosedUncleanly(new Exception('error')));
  });

  test('isConnected', () {
    var futures = [];

    var monitor = new BaseFTransportMonitor();
    expect(monitor.isConnected, equals(true));

    futures.add(monitor.onDisconnect.first);
    monitor.onClosedCleanly();
    expect(monitor.isConnected, isFalse);

    monitor.onReopenFailed(1, 1);
    expect(monitor.isConnected, isFalse);

    futures.add(monitor.onConnect.first);
    monitor.onReopenSucceeded();
    expect(monitor.isConnected, isTrue);

    futures.add(monitor.onDisconnect.first);
    monitor.onClosedUncleanly(new Exception('error'));
    expect(monitor.isConnected, isFalse);

    return Future.wait(futures);
  });

  test(
      'onClosedUncleanly should return expected wait period if max attempts > 0',
      () {
    FTransportMonitor monitor = new BaseFTransportMonitor(
        maxReopenAttempts: 1, initialWait: 1, maxWait: 1);
    expect(1, monitor.onClosedUncleanly(new Exception('error')));
  });

  test('onReopenFailed should return -1 if max attempts is reached', () {
    FTransportMonitor monitor = new BaseFTransportMonitor(
        maxReopenAttempts: 1, initialWait: 0, maxWait: 0);
    expect(-1, monitor.onReopenFailed(1, 0));
  });

  test('onReopenFailed should return double the previous wait', () {
    FTransportMonitor monitor = new BaseFTransportMonitor(
        maxReopenAttempts: 6, initialWait: 1, maxWait: 10);
    expect(2, monitor.onReopenFailed(0, 1));
  });

  test('onReopenFailed should respect the max wait', () {
    FTransportMonitor monitor = new BaseFTransportMonitor(
        maxReopenAttempts: 6, initialWait: 1, maxWait: 1);
    expect(1, monitor.onReopenFailed(0, 1));
  });

  test('close cleanly provides no cause', () async {
    var monitor = new BaseFTransportMonitor();
    monitor.onDisconnect.listen(expectAsync((cause) {
      expect(cause, isNull);
    }));
    monitor.onClosedCleanly();
  });

  test('closeUncleanly provides a cause', () async {
    var monitor =
        new BaseFTransportMonitor(initialWait: 1, maxReopenAttempts: 0);
    var error = new StateError("fake error");
    monitor.onDisconnect.listen(expectAsync((cause) {
      expect(cause, error);
    }));
    monitor.onClosedUncleanly(error);
  });
}
