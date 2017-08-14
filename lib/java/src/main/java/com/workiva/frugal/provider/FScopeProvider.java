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
import com.workiva.frugal.transport.FPublisherTransport;
import com.workiva.frugal.transport.FPublisherTransportFactory;
import com.workiva.frugal.transport.FSubscriberTransport;
import com.workiva.frugal.transport.FSubscriberTransportFactory;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * FScopeProvider produces FPublisherTransports, FSubscriberTransports, and
 * FProtocols for use by pub/sub scopes. It does this by wrapping an
 * FPublisherTransportFactory, an FSubscriberTransportFactory, and an
 * FProtocolFactory. This also provides a shim for adding middleware to a
 * publisher or subscriber.
 */
public class FScopeProvider {

    /**
     * Publisher of this scope.
     */
    public static class Publisher {
        private FPublisherTransport transport;
        private FProtocolFactory protocolFactory;

        private Publisher(FPublisherTransport t, FProtocolFactory pf) {
            transport = t;
            protocolFactory = pf;
        }

        public FPublisherTransport getTransport() {
            return transport;
        }

        public FProtocolFactory getProtocolFactory() {
            return protocolFactory;
        }
    }

    /**
     * Subscriber of this scope.
     */
    public static class Subscriber {
        private FSubscriberTransport transport;
        private FProtocolFactory protocolFactory;

        private Subscriber(FSubscriberTransport t, FProtocolFactory pf) {
            transport = t;
            protocolFactory = pf;
        }

        public FSubscriberTransport getTransport() {
            return transport;
        }

        public FProtocolFactory getProtocolFactory() {
            return protocolFactory;
        }
    }

    private FPublisherTransportFactory publisherTransportFactory;
    private FSubscriberTransportFactory subscriberTransportFactory;
    private FProtocolFactory protocolFactory;
    private List<ServiceMiddleware> middleware;

    public FScopeProvider(FPublisherTransportFactory ptf, FSubscriberTransportFactory stf,
                          FProtocolFactory pf, ServiceMiddleware ...middleware) {
        publisherTransportFactory = ptf;
        subscriberTransportFactory = stf;
        protocolFactory = pf;
        this.middleware = Arrays.asList(middleware);
    }

    /**
     * Returns a new Publisher containing an FPublisherTransport and FProtocolFactory
     * used for publishing.
     *
     * @return Publisher with FPublisherTransport and FProtocol.
     */
    public Publisher buildPublisher() {
        FPublisherTransport transport = publisherTransportFactory.getTransport();
        return new Publisher(transport, protocolFactory);
    }

    /**
     * Returns a new Subscriber containing an FSubscriberTransport and FProtocolFactory
     * used for subscribing.
     *
     * @return SubscriberClient with FSubscriberTransport and FProtocol.
     */
    public Subscriber buildSubscriber() {
        FSubscriberTransport transport = subscriberTransportFactory.getTransport();
        return new Subscriber(transport, protocolFactory);
    }

    public List<ServiceMiddleware> getMiddleware() {
        return new ArrayList<>(middleware);
    }
}
