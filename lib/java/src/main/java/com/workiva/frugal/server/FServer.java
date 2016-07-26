package com.workiva.frugal.server;

import org.apache.thrift.TException;

/**
 * FServer is Frugal's equivalent of Thrift's TServer. It's used to run a Frugal
 * RPC service by executing an FProcessor on client connections. FServer can
 * optionally support a high-water mark which is the maximum amount of time a
 * request is allowed to be enqueued before triggering server overload logic (e.g.
 * load shedding).
 * <p/>
 * Currently, Frugal includes two implementations of FServer: FSimpleServer, which
 * is a basic, accept-loop based server that supports traditional Thrift
 * TServerTransports, and FNatsServer, which is an implementation that uses NATS
 * as the underlying transport.
 */
public interface FServer {

    /**
     * Starts the server.
     *
     * @throws TException
     */
    void serve() throws TException;

    /**
     * Stops the server. This is optional on a per-implementation basis.
     * Not all servers are required to be cleanly stoppable.
     *
     * @throws TException
     */
    void stop() throws TException;

    /**
     * Sets the maximum amount of time a frame is allowed to await processing
     * before triggering transport overload logic.
     *
     * @param watermark the watermark time in milliseconds.
     *
     * @deprecated This will be a constructor implementation detail for
     * servers which buffer client requests.
     */
    @Deprecated
    void setHighWatermark(long watermark);
}
