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

/// A default monitor implementation that attempts to reopen a closed transport
/// with exponential backoff behavior and a capped number of retries. Its
/// behavior can be customized by extending this class and overriding desired
/// callbacks.
class BaseFTransportMonitor extends FTransportMonitor {
  /// Default maximum reopen attempts.
  static const int DEFAULT_MAX_REOPEN_ATTEMPTS = 60;

  /// Default number of milliseconds to wait before reopening.
  static const int DEFAULT_INITIAL_WAIT = 2000;

  /// Default maximum amount of milliseconds to wait between reopen attempts.
  static const int DEFAULT_MAX_WAIT = 2000;

  int _maxReopenAttempts;
  int _initialWait;
  int _maxWait;

  StreamController _onConnectController = new StreamController.broadcast();
  StreamController _onDisconnectController = new StreamController.broadcast();

  bool _isConnected = true;

  /// Create a [BaseFTransportMonitor] with default parameters.
  BaseFTransportMonitor(
      {maxReopenAttempts: DEFAULT_MAX_REOPEN_ATTEMPTS,
      initialWait: DEFAULT_INITIAL_WAIT,
      maxWait: DEFAULT_MAX_WAIT}) {
    this._maxReopenAttempts = maxReopenAttempts;
    this._initialWait = initialWait;
    this._maxWait = maxWait;
  }

  /// Listen to connect events.
  Stream get onConnect => _onConnectController.stream;

  /// Listen to disconnect events.
  Stream get onDisconnect => _onDisconnectController.stream;

  /// Queries the state of the [FTransport].
  bool get isConnected => _isConnected;

  @override
  void onClosedCleanly() {
    _isConnected = false;
    _onDisconnectController.add(null);
  }

  @override
  int onClosedUncleanly(Object cause) {
    _isConnected = false;
    _onDisconnectController.add(cause);

    return _maxReopenAttempts > 0 ? _initialWait : -1;
  }

  @override
  int onReopenFailed(int prevAttempts, int prevWait) {
    if (prevAttempts >= _maxReopenAttempts) {
      return -1;
    }

    return (prevWait * 2).clamp(0, _maxWait);
  }

  @override
  void onReopenSucceeded() {
    _isConnected = true;
    _onConnectController.add(null);
  }
}
