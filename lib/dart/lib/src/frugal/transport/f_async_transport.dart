part of frugal.src.frugal;

/// FAsyncTransport is an extension of FTransport that asynchronous frameworks
/// can implement. Implementations need only implement [flush] to send request
/// data and call [handleResponse] when asynchronous responses are received.
abstract class FAsyncTransport extends FTransport {
  final Logger _log = new Logger('FAsyncTransport');
  Map<int, Completer<Uint8List>> _handlers = {};

  /// Instantiate an [FAsyncTransport].
  FAsyncTransport({int requestSizeLimit})
      : super(requestSizeLimit: requestSizeLimit);

  /// Flush the payload to the server. Implementations must be threadsafe.
  Future<Null> flush(Uint8List payload);

  @override
  Future<Null> oneway(FContext ctx, Uint8List payload) async {
    _preflightRequestCheck(payload);
    await flush(payload).timeout(ctx.timeout, onTimeout: () {
      throw new TTransportError(FrugalTTransportErrorType.TIMED_OUT,
          'request timed out after ${ctx.timeout}');
    });
  }

  @override
  Future<TTransport> request(FContext ctx, Uint8List payload) async {
    _preflightRequestCheck(payload);

    Completer<Uint8List> resultCompleter = new Completer();

    if (_handlers.containsKey(ctx._opId)) {
      throw new StateError("frugal: context already registered");
    }
    _handlers[ctx._opId] = resultCompleter;
    Completer<Uint8List> closedCompleter = new Completer();
    StreamSubscription<Object> closedSub = onClose.listen((_) {
      closedCompleter.completeError(
          new TTransportError(FrugalTTransportErrorType.NOT_OPEN));
    });

    try {
      Future<Uint8List> resultFuture =
          _request(payload, resultCompleter).timeout(ctx.timeout);

      // Bail early if the transport is closed
      Uint8List response =
          await Future.any([resultFuture, closedCompleter.future]);
      return new TMemoryTransport.fromUint8List(response);
    } on TimeoutException catch (_) {
      throw new TTransportError(FrugalTTransportErrorType.TIMED_OUT,
          "request timed out after ${ctx.timeout}");
    } finally {
      _handlers.remove(ctx._opId);
      await closedSub.cancel();
    }
  }

  /// Flushes the payload to the server and waits for the response to be
  /// received.
  Future<Uint8List> _request(
      Uint8List payload, Completer<Uint8List> completer) async {
    await flush(payload);
    return await completer.future;
  }

  /// Handles a frugal frame response. NOTE: this frame must NOT include the
  /// frame size. Implementations should call this when asynchronous responses
  /// are received from the server.
  void handleResponse(Uint8List frame) {
    var headers = Headers.decodeFromFrame(frame);
    var opId;
    try {
      opId = int.parse(headers[_opidHeader]);
    } catch (e) {
      _log.warning("frugal: invalid protocol frame: op id not a uint64", e);
      return;
    }

    if (!_handlers.containsKey(opId)) {
      return;
    }

    _handlers[opId].complete(frame);
  }
}
