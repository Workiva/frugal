package com.workiva.frugal.transport;

/**
 * FScopeTransportFactory produces FScopeTransports and is typically used by an
 * FScopeProvider.
 */
public interface FScopeTransportFactory {

    /**
     * Get a new FScopeTransport instance.
     *
     * @return A new FScopeTransport instance.
     */
    FScopeTransport getTransport();

}
