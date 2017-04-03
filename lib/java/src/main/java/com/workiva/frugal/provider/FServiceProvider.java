package com.workiva.frugal.provider;

import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FTransport;
import org.apache.thrift.transport.TTransportException;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * FServiceProvider is the service equivalent of FScopeProvider. It produces
 * FTransports and FProtocols for use by RPC service clients. The main
 * purpose of this is to provide a shim for adding middleware to a client.
 */
public class FServiceProvider {

    private FTransport transport;
    private FProtocolFactory protocolFactory;
    private List<ServiceMiddleware> middleware = new ArrayList<>();

    public FServiceProvider(FTransport transport, FProtocolFactory protocolFactory) {
        this.transport = transport;
        this.protocolFactory = protocolFactory;
    }

    public FServiceProvider(FTransport transport, FProtocolFactory protocolFactory,
                            ServiceMiddleware... middleware) {
        this.transport = transport;
        this.protocolFactory = protocolFactory;
        this.middleware = Arrays.asList(middleware);
    }

    /**
     * Open the FServiceProvider so it can be used by a client.
     *
     * @throws TTransportException if opening the transport failed
     */
    public void open() throws TTransportException {
        transport.open();
    }

    /**
     * Close the FServiceProvider and dispose the transport.
     */
    public void close() {
        transport.close();
    }

    /**
     * Get the FTransport from the provider.
     *
     * @return FTransport instance stored on the provider.
     */
    public FTransport getTransport() {
        return transport;
    }

    /**
     * Get the FProtocolFactory from the provider.
     *
     * @return FProtocolFactory stored on the provider.
     */
    public FProtocolFactory getProtocolFactory() {
        return protocolFactory;
    }

    public List<ServiceMiddleware> getMiddleware() {
        return new ArrayList<>(middleware);
    }
}
