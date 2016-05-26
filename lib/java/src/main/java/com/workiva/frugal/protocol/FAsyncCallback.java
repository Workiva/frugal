package com.workiva.frugal.protocol;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;

/**
 * FAsyncCallback is an internal callback which is constructed by generated code
 * and invoked by an FRegistry when a RPC response is received. In other words,
 * it's used to complete RPCs. The operation ID on FContext is used to look up the
 * appropriate callback. FAsyncCallback is passed an in-memory TTransport which
 * wraps the complete message. The callback returns an error or throws an
 * exception if an unrecoverable error occurs and the transport needs to be
 * shutdown.
 */
public interface FAsyncCallback {

    void onMessage(TTransport transport) throws TException;

}