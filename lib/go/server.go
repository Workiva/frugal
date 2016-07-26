package frugal

import "time"

// FServer is Frugal's equivalent of Thrift's TServer. It's used to run a Frugal
// RPC service by executing an FProcessor on client connections.
type FServer interface {
	// Serve starts the server.
	Serve() error

	// Stop the server. This is optional on a per-implementation basis. Not all
	// servers are required to be cleanly stoppable.
	Stop() error

	// SetHighWatermark sets the maximum amount of time a frame is allowed to
	// await processing before triggering server overload logic.
	// DEPRECATED - This will be a constructor implementation detail for
	// servers which buffer client requests.
	// TODO: Remove this with 2.0
	SetHighWatermark(watermark time.Duration)
}
