package com.workiva.frugal.transport;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;

public abstract class FScopeTransport extends TTransport {
    /**
     * Sets the publish topic and locks the transport for exclusive access.
     *
     * @param topic the pub/sub topic to publish on.
     * @throws TException
     */
    public abstract void lockTopic(String topic) throws TException;

    /**
     * Unsets the publish topic and unlocks the transport.
     *
     * @throws TException
     */
    public abstract void unlockTopic() throws TException;

    /**
     * Opens the Transport to receive messages on the subscription.
     *
     * @param topic the pub/sub topic to subscribe to.
     * @throws TException
     */
    public abstract void subscribe(String topic) throws TException;

    /**
     * Discards the current message frame the transport is reading, if any. After calling this, a subsequent call to
     * read will read from the next frame. This must be called from the same thread as the thread calling read.
     */
    public abstract void discardFrame();
}
