part of frugal;

/// Invoked when a message frame is received. Any error thrown by
/// this callback is considered fatal and will result in the
/// shutdown of the transport.
typedef void FAsyncCallback(TTransport);

/// Responsible for multiplexing received messages to the
/// appropriate callback.
abstract class FRegistry {
  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback);

  /// Unregister a callback for the given Context.
  void unregister(Context);

  /// Dispatch a single Frugal message frame.
  void execute(Uint8List);
}
