part of frugal;

/// BaseFTransportMonitor is a default monitor implementation that attempts
/// to reopen a closed transport with exponential backoff behavior
/// and a capped number of retries. Its behavior can be customized by extending
/// this class and overriding desired callbacks.
class BaseFTransportMonitor extends FTransportMonitor {
  static const int DEFAULT_MAX_REOPEN_ATTEMPTS = 60;
  static const int DEFAULT_INITIAL_WAIT = 2000;
  static const int DEFAULT_MAX_WAIT = 2000;

  int _maxReopenAttempts;
  int _initialWait;
  int _maxWait;

  StreamController _onConnectController = new StreamController.broadcast();
  StreamController _onDisconnectController = new StreamController.broadcast();

  Stream get onConnect => _onConnectController.stream;
  Stream get onDisconnect => _onDisconnectController.stream;

  bool _isConnected = true;
  bool get isConnected => _isConnected;

  BaseFTransportMonitor(
      {maxReopenAttempts: DEFAULT_MAX_REOPEN_ATTEMPTS,
      initialWait: DEFAULT_INITIAL_WAIT,
      maxWait: DEFAULT_MAX_WAIT}) {
    this._maxReopenAttempts = maxReopenAttempts;
    this._initialWait = initialWait;
    this._maxWait = maxWait;
  }

  @override
  void onClosedCleanly() {
    _isConnected = false;
    _onDisconnectController.add(null);
  }

  @override
  int onClosedUncleanly(cause) {
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
