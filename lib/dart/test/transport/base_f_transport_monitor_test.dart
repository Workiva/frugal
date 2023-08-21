import 'dart:async';

import 'package:frugal/frugal.dart';
import 'package:test/test.dart';

void main() {
  test('onClosedUncleanly should return -1 if max attempts is 0', () {
    FTransportMonitor monitor = BaseFTransportMonitor(maxReopenAttempts: 0, initialWait: 0, maxWait: 0);
    expect(-1, monitor.onClosedUncleanly(Exception('error')));
  });

  test('isConnected', () {
    var monitor = BaseFTransportMonitor();
    expect(monitor.isConnected, equals(true));

    var one = monitor.onDisconnect.first;
    monitor.onClosedCleanly();
    expect(monitor.isConnected, isFalse);

    monitor.onReopenFailed(1, 1);
    expect(monitor.isConnected, isFalse);

    var two = monitor.onConnect.first;
    monitor.onReopenSucceeded();
    expect(monitor.isConnected, isTrue);

    var three = monitor.onDisconnect.first;
    monitor.onClosedUncleanly(Exception('error'));
    expect(monitor.isConnected, isFalse);

    return Future.wait([one, two, three]);
  });

  test('onClosedUncleanly should return expected wait period if max attempts > 0', () {
    FTransportMonitor monitor = BaseFTransportMonitor(maxReopenAttempts: 1, initialWait: 1, maxWait: 1);
    expect(1, monitor.onClosedUncleanly(Exception('error')));
  });

  test('onReopenFailed should return -1 if max attempts is reached', () {
    FTransportMonitor monitor = BaseFTransportMonitor(maxReopenAttempts: 1, initialWait: 0, maxWait: 0);
    expect(-1, monitor.onReopenFailed(1, 0));
  });

  test('onReopenFailed should return double the previous wait', () {
    FTransportMonitor monitor = BaseFTransportMonitor(maxReopenAttempts: 6, initialWait: 1, maxWait: 10);
    expect(2, monitor.onReopenFailed(0, 1));
  });

  test('onReopenFailed should respect the max wait', () {
    FTransportMonitor monitor = BaseFTransportMonitor(maxReopenAttempts: 6, initialWait: 1, maxWait: 1);
    expect(1, monitor.onReopenFailed(0, 1));
  });

  test('close cleanly provides no cause', () async {
    var monitor = BaseFTransportMonitor();
    // ignore: strong_mode_down_cast_composite
    monitor.onDisconnect.listen(expectAsync1((cause) {
      expect(cause, isNull);
    }));
    monitor.onClosedCleanly();
  });

  test('closeUncleanly provides a cause', () async {
    var monitor = BaseFTransportMonitor(initialWait: 1, maxReopenAttempts: 0);
    var error = StateError("fake error");
    // ignore: strong_mode_down_cast_composite
    monitor.onDisconnect.listen(expectAsync1((cause) {
      expect(cause, error);
    }));
    monitor.onClosedUncleanly(error);
  });
}
