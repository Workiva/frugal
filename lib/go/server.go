package frugal

import "time"

// FServer is Frugal's equivalent of Thrift's TServer. It's used to run a Frugal
// RPC service by executing an FProcessor on client connections. FServer can
// optionally support a high-water mark which is the maximum amount of time a
// request is allowed to be enqueued before triggering server overload logic
// (e.g. load shedding).
//
// Currently, Frugal includes two implementations of FServer: FSimpleServer,
// which is a basic, accept-loop based server that supports traditional Thrift
// TServerTransports, and FNatsServer, which is an implementation that uses
// NATS as the underlying transport.
type FServer interface {
	// Serve starts the server.
	Serve() error

	// Stop the server. This is optional on a per-implementation basis. Not all
	// servers are required to be cleanly stoppable.
	Stop() error

	// SetHighWatermark sets the maximum amount of time a frame is allowed to
	// await processing before triggering server overload logic.
	SetHighWatermark(watermark time.Duration)
}
