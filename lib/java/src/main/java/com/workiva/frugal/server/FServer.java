package com.workiva.frugal.server;

import org.apache.thrift.TException;

/**
 * FServer is a Frugal service server.
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
     */
    void setHighWatermark(long watermark);
}
