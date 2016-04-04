part of frugal;

/// MonitorRunner runs an FTransportMonitor when a transport is closed.
class MonitorRunner {
  final Logger log = new Logger('FTransportMonitor');
  FTransportMonitor _monitor;
  TTransport _transport;

  MonitorRunner(this._monitor, this._transport);

  Future onClose(cause) async {
    if (cause == null) {
      _handleCleanClose();
    } else {
      await _handleUncleanClose(cause);
    }
  }

  void _handleCleanClose() {
    log.info('transport was closed cleanly');
    _monitor.onClosedCleanly();
  }

  Future _handleUncleanClose(cause) async {
    log.warning('transport was closed uncleanly because: $cause');
    int wait = _monitor.onClosedUncleanly(cause);
    if (wait < 0) {
      log.warning('instructed not to reopen');
      return;
    }
    await _attemptReopen(wait);
  }

  Future _attemptReopen(int initialWait) async {
    int wait = initialWait;
    int prevAttempts = 0;

    while (wait >= 0) {
      log.info('attemptying to reopen after $wait ms');
      await new Future.delayed(new Duration(milliseconds: wait));

      try {
        await _transport.open();
      } on TTransportError catch (e) {
        log.warning('failed to reopen transport due to: $e');
        prevAttempts++;
        wait = _monitor.onReopenFailed(prevAttempts, wait);
        continue;
      }

      log.info('successfully reopened transport');
      _monitor.onReopenSucceeded();
      return;
    }

    log.warning('ReopenFailed callback instructed not to reopen, terminating');
  }
}
