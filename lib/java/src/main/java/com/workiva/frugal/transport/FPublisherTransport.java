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

import org.apache.thrift.transport.TTransportException;

/**
 * FPublisherTransport is used exclusively for scope publishers.
 */
public interface FPublisherTransport {

    /**
     * Queries whether the transport is open.
     *
     * @return True if the transport is open.
     */
    boolean isOpen();

    /**
     * Opens the transport.
     *
     * @throws TTransportException if the transport could not be opened.
     */
    void open() throws TTransportException;

    /**
     * Closes the transport.
     */
    void close();

    /**
     * Get the maximum publish size permitted by the transport. If <code>getPublishSizeLimit</code>
     * returns a non-positive number, the transport is assumed to have no publish size limit.
     *
     * @return the publish size limit
     */
    int getPublishSizeLimit();

    /**
     * Publish the given framed frugal payload over the transport. Implementations of <code>publish</code>
     * should be thread-safe.
     *
     * @param topic the topic on which to publish the payload
     * @param payload framed frugal bytes
     * @throws TTransportException if publishing the payload failed
     */
    void publish(String topic, byte[] payload) throws TTransportException;
}
