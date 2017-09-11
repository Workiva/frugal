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

package com.workiva.frugal.transport;

import com.workiva.frugal.exception.TTransportExceptionType;
import io.nats.client.Connection;
import io.nats.client.Nats;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;

import static com.workiva.frugal.transport.FNatsTransport.FRUGAL_PREFIX;
import static com.workiva.frugal.transport.FNatsTransport.NATS_MAX_MESSAGE_SIZE;
import static com.workiva.frugal.transport.FNatsTransport.getClosedConditionException;

/**
 * FNatsPublisherTransport implements FPublisherTransport by using NATS as the pub/sub message broker.
 * Messages are limited to 1MB in size.
 */
public class FNatsPublisherTransport implements FPublisherTransport {
    private static final Logger LOGGER = LoggerFactory.getLogger(FNatsPublisherTransport.class);

    private final Connection conn;

    /**
     * Creates a new FNatsPublisherTransport which is used for publishing.
     *
     * @param conn NATS connection
     */
    protected FNatsPublisherTransport(Connection conn) {
        this.conn = conn;
    }

    /**
     * An FPublisherTransportFactory implementation which creates FPublisherTransports backed by NATS.
     */
    public static class Factory implements FPublisherTransportFactory {

        private final Connection conn;

        /**
         * Creates a NATS FPublisherTransportFactory using the provided NATS connection.
         *
         * @param conn NATS connection
         */
        public Factory(Connection conn) {
            this.conn = conn;
        }

        /**
         * Get a new FPublisherTransport instance.
         *
         * @return A new FPublisherTransport instance.
         */
        public FPublisherTransport getTransport() {
            return new FNatsPublisherTransport(this.conn);
        }
    }

    @Override
    public synchronized boolean isOpen() {
        return conn.getState() == Nats.ConnState.CONNECTED;
    }

    @Override
    public synchronized void open() throws TTransportException {
        // We only need to check that the NATS client is connected
        if (conn.getState() != Nats.ConnState.CONNECTED) {
            throw new TTransportException(TTransportExceptionType.NOT_OPEN,
                    "NATS not connected, has status " + conn.getState());
        }
    }

    @Override
    public synchronized void close() {
        /* Do nothing */
    }

    @Override
    public int getPublishSizeLimit() {
        return NATS_MAX_MESSAGE_SIZE;
    }

    @Override
    public void publish(String topic, byte[] payload) throws TTransportException {
        if (!isOpen()) {
            throw getClosedConditionException(conn.getState(), "publish:");
        }

        if ("".equals(topic)) {
            throw new TTransportException("Subject cannot be empty.");
        }

        if (payload.length > NATS_MAX_MESSAGE_SIZE) {
            throw new TTransportException(TTransportExceptionType.REQUEST_TOO_LARGE,
                    String.format("Message exceeds %d bytes, was %d bytes",
                            NATS_MAX_MESSAGE_SIZE, payload.length));
        }

        try {
            conn.publish(getFormattedSubject(topic), payload);
        } catch (IOException e) {
            throw new TTransportException("publish: unable to publish data: " + e.getMessage());
        }
    }

    private String getFormattedSubject(String topic) {
        return FRUGAL_PREFIX + topic;
    }
}
