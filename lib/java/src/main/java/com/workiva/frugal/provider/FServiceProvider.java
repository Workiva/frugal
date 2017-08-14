/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.workiva.frugal.provider;

import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FTransport;

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
                            ServiceMiddleware ...middleware) {
        this.transport = transport;
        this.protocolFactory = protocolFactory;
        this.middleware = Arrays.asList(middleware);
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
