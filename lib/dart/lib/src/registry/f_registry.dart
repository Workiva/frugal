part of frugal;

/// FAsyncCallback is an internal callback which is constructed by generated
/// code and invoked by an FRegistry when a RPC response is received. In other
/// words, it's used to complete RPCs. The operation ID on FContext is used to
/// look up the appropriate callback. FAsyncCallback is passed an in-memory
/// TTransport which wraps the complete message. The callback returns an error
/// or throws an exception if an unrecoverable error occurs and the transport
/// needs to be shutdown.
typedef void FAsyncCallback(TTransport);

/// FRegistry is responsible for multiplexing and handling received messages.
/// Typically there is a client implementation and a server implementation. An
/// FRegistry is used by an FTransport.
///
/// The client implementation is used on the client side, which is making RPCs.
/// When a request is made, an FAsyncCallback is registered to an FContext.
/// When a response for the FContext is received, the FAsyncCallback is looked
/// up, executed, and unregistered.
///
/// The server implementation is used on the server side, which is handling
/// RPCs. It does not actually register FAsyncCallbacks but rather has an
/// FProcessor registered with it. When a message is received, it's buffered
/// and passed to the FProcessor to be handled.
abstract class FRegistry {
  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback);

  /// Unregister a callback for the given Context.
  void unregister(Context);

  /// Dispatch a single Frugal message frame.
  void execute(Uint8List);
}
