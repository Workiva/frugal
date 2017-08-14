/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

part of frugal.src.frugal;

/// Runs an [FTransportMonitor] when a transport is closed.
class MonitorRunner {
  final Logger _log = new Logger('FTransportMonitor');
  FTransportMonitor _monitor;
  FTransport _transport;
  int _attempts = 0;
  int _wait = 0;
  bool _failed = false;
  Completer _reopenCompleter;
  Timer _reopenTimer;

  /// Create a new [MonitorRunner] with the given [FTransportMonitor] and
  /// [FTransport].
  MonitorRunner(this._monitor, this._transport);

  /// Indicates if the monitor is waiting to run or gave up.
  bool get _sleeping => (_reopenTimer != null || _failed);

  /// Handle close event.
  Future onClose(Object cause) async {
    if (cause == null) {
      _handleCleanClose();
    } else {
      await _handleUncleanClose(cause);
    }
  }

  /// Signal that the transport is now open.
  void signalOpen() {
    if (_sleeping) _signalOpen();
  }

  void _signalOpen() {
    _log.log(Level.INFO, 'successfully reopened transport');
    _stop();
    _monitor.onReopenSucceeded();
    return;
  }

  void _stop({bool failed: false}) {
    _attempts = 0;
    _wait = 0;
    _failed = failed;
    _reopenCompleter?.complete();
    _reopenCompleter = null;
    _reopenTimer?.cancel();
    _reopenTimer = null;
  }

  void _handleCleanClose() {
    _log.log(Level.INFO, 'transport was closed cleanly');
    _monitor.onClosedCleanly();
  }

  Future _handleUncleanClose(cause) async {
    if (_reopenCompleter != null) {
      // TODO: Should we reset _attemps/_wait? Or does this indicate something
      // bigger is wrong?
      _log.log(Level.WARNING, 'received multiple unclean close calls!');
      return;
    }

    _log.log(Level.WARNING, 'transport was closed uncleanly because: $cause');
    _wait = _monitor.onClosedUncleanly(cause);
    if (_wait < 0) {
      _log.log(Level.WARNING, 'instructed not to reopen');
      _stop(failed: true);
      return;
    }
    _reopenCompleter = new Completer();
    _startReopenTimer();
    await _reopenCompleter.future;
  }

  void _startReopenTimer() {
    _log.log(Level.INFO, 'attempting to reopen after $_wait ms');
    _reopenTimer = new Timer(new Duration(milliseconds: _wait), _attemptReopen);
  }

  Future _attemptReopen() async {
    // Not sleeping anymore.
    _reopenTimer = null;
    try {
      await _transport.open();
      _signalOpen();
    } catch (e) {
      _log.log(Level.WARNING, 'failed to reopen transport due to: $e');
      _attempts++;
      _wait = _monitor.onReopenFailed(_attempts, _wait);
      if (_wait >= 0) {
        _startReopenTimer();
        return;
      }
      _stop(failed: true);
      _log.log(Level.WARNING,
          'ReopenFailed callback instructed not to reopen, terminating');
    }
  }
}
