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
import com.workiva.frugal.protocol.FAsyncCallback;
import io.nats.client.Connection;

import io.nats.client.Connection.Status;
import io.nats.client.Dispatcher;
import org.apache.thrift.TConfiguration;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import static com.workiva.frugal.transport.FNatsTransport.FRUGAL_PREFIX;

/**
 * FNatsSubscriberTransport implements FSubscriberTransport by using NATS as the pub/sub message broker.
 * Messages are limited to 1MB in size.
 */
public class FNatsSubscriberTransport implements FSubscriberTransport {

    private static final Logger LOGGER = LoggerFactory.getLogger(FNatsSubscriberTransport.class);

    private final Connection conn;
    protected String subject;
    protected final String queue;
    protected Dispatcher dispatcher;
    private final TConfiguration tconfig;

    /**
     * Creates a new FNatsScopeTransport which is used for subscribing. Subscribers using this transport will subscribe
     * to the provided queue, forming a queue group. When a queue group is formed, only one member receives the message.
     * If the queue is null, then the subscriber does not join a queue group.
     *
     * @param conn  NATS connection
     * @param queue subscription queue
     */
    protected FNatsSubscriberTransport(Connection conn, String queue) {
        this.conn = conn;
        this.queue = queue;
        this.tconfig = new TConfiguration(
                Integer.MAX_VALUE,
                TConfiguration.DEFAULT_MAX_FRAME_SIZE,
                TConfiguration.DEFAULT_RECURSION_DEPTH
        );
    }

    /**
     * An FSubscriberTransportFactory implementation which creates FSubscriberTransports backed by NATS.
     */
    public static class Factory implements FSubscriberTransportFactory {

        private final Connection conn;
        private final String queue;

        /**
         * Creates a NATS FSubscriberTransportFactory using the provided NATS connection.
         *
         * @param conn NATS connection
         */
        public Factory(Connection conn) {
            this(conn, null);
        }

        /**
         * Creates a NATS FSubscriberTransportFactory using the provided NATS connection. Subscribers using this
         * transport will subscribe to the provided queue, forming a queue group. When a queue group is formed,
         * only one member receives the message.
         *
         * @param conn  NATS connection
         * @param queue subscription queue
         */
        public Factory(Connection conn, String queue) {
            this.conn = conn;
            this.queue = queue;
        }

        /**
         * Get a new FSubscriberTransport instance.
         *
         * @return A new FSubscriberTransport instance.
         */
        public FNatsSubscriberTransport getTransport() {
            return new FNatsSubscriberTransport(conn, queue);
        }
    }

    @Override
    public boolean isSubscribed() {
        return conn.getStatus() == Status.CONNECTED && dispatcher != null && dispatcher.isActive();
    }

    @Override
    public void subscribe(String topic, FAsyncCallback callback) throws TException {
        if (conn.getStatus() != Status.CONNECTED) {
            throw new TTransportException(TTransportExceptionType.NOT_OPEN,
                    "NATS not connected, has status " + conn.getStatus());
        }

        subject = topic;
        if ("".equals(subject)) {
            throw new TTransportException("Subject cannot be empty.");
        }

        dispatcher = conn.createDispatcher(msg -> {
            if (msg.getData().length < 4) {
                LOGGER.warn("discarding invalid scope message frame");
                return;
            }
            try {
                callback.onMessage(
                        new TMemoryInputTransport(tconfig, msg.getData(), 4, msg.getData().length - 4)
                );
            } catch (Throwable t) {
                LOGGER.error("error executing user provided callback", t);
            }
        });

        if (queue != null && !queue.isEmpty()) {
            dispatcher.subscribe(getFormattedSubject(), queue);
        } else {
            dispatcher.subscribe(getFormattedSubject());
        }
    }

    @Override
    public synchronized void unsubscribe() {
        if (dispatcher == null) {
            LOGGER.warn("attempted to unsubscribe with a null internal " +
                    "subscription - possibly unsubscribing more than once - subject: " + subject);
            return;
        }
        dispatcher.unsubscribe(getFormattedSubject());
        conn.closeDispatcher(dispatcher);
        dispatcher = null;
    }

    private String getFormattedSubject() {
        return FRUGAL_PREFIX + subject;
    }

}

