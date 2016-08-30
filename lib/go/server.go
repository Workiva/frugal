package frugal

// FServer is Frugal's equivalent of Thrift's TServer. It's used to run a Frugal
// RPC service by executing an FProcessor on client connections.
type FServer interface {
	// Serve starts the server.
	Serve() error

	// Stop the server. This is optional on a per-implementation basis. Not all
	// servers are required to be cleanly stoppable.
	Stop() error
}
