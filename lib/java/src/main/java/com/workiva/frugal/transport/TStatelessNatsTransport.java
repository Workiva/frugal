package com.workiva.frugal.transport;

import io.nats.client.Connection;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

/**
 * TStatelessNatsTransport is an extension of thrift.TTransport. This is a "stateless" transport in the sense that there
 * is no connection with a server. A request is simply published to a subject and responses are received on another
 * subject. This assumes requests/responses fit within a single NATS message.
 *
 * @deprecated Use FNatsTransport instead.
 */
@Deprecated
public class TStatelessNatsTransport extends TTransport {

    protected final FNatsTransport fNatsTransport;

    /**
     * Creates a new Thrift TTransport which uses the NATS messaging system as the underlying transport. Unlike
     * TNatsServiceTransport, this TTransport is stateless in that there is no connection maintained between the client
     * and server. A request is simply published to a subject and responses are received on a randomly generated
     * subject. This requires requests to fit within a single NATS message.
     *
     * @param conn    NATS connection
     * @param subject subject to publish requests on
     *
     * @deprecated Use FNatsTransport instead
     */
    @Deprecated
    public TStatelessNatsTransport(Connection conn, String subject) {
        this(conn, subject, conn.newInbox());
    }

    /**
     * Creates a new Thrift TTransport which uses the NATS messaging system as the underlying transport. Unlike
     * TNatsServiceTransport, this TTransport is stateless in that there is no connection maintained between the client
     * and server. A request is simply published to a subject and responses are received on a specified subject. This
     * requires requests to fit within a single NATS message.
     *
     * @param conn    NATS connection
     * @param subject subject to publish requests on
     * @param inbox   subject to receive responses on
     *
     * @deprecated Use FNatsTransport instead
     */
    @Deprecated
    public TStatelessNatsTransport(Connection conn, String subject, String inbox) {
        this.fNatsTransport = new FNatsTransport(conn, subject, inbox, true);
    }

    @Override
    public synchronized boolean isOpen() {
        return fNatsTransport.isOpen();
    }

    /**
     * Subscribes to the configured inbox subject.
     *
     * @throws TTransportException
     */
    @Override
    public synchronized void open() throws TTransportException {
        fNatsTransport.open();
    }

    /**
     * Unsubscribes from the inbox subject and closes the response buffer.
     */
    @Override
    public synchronized void close() {
        fNatsTransport.close();
    }

    /**
     * Reads up to len bytes into the buffer.
     *
     * @throws TTransportException
     */
    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        return fNatsTransport.read(bytes, off, len);
    }

    /**
     * Writes the bytes to a buffer. Throws FMessageSizeException if the buffer exceeds 1MB.
     *
     * @throws TTransportException
     */
    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        fNatsTransport.write(bytes, off, len);
    }

    /**
     * Sends the buffered bytes over NATS.
     *
     * @throws TTransportException
     */
    @Override
    public void flush() throws TTransportException {
        fNatsTransport.flush();
    }
}
