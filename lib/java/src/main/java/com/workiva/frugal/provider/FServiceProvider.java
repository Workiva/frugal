package com.workiva.frugal.provider;

import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FTransport;

/**
 * FServiceProvider is the service equivalent of FScopeProvider. It produces
 * FTransports and FProtocols for use by RPC service clients.
 */
public class FServiceProvider {

    private FTransport transport;
    private FProtocolFactory protocolFactory;

    public FServiceProvider(FTransport transport, FProtocolFactory protocolFactory) {
        this.transport = transport;
        this.protocolFactory = protocolFactory;
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
}
