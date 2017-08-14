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

import com.workiva.frugal.protocol.FAsyncCallback;
import org.apache.thrift.TException;

/**
 * FSubscriberTransport is used exclusively for scope publishers.
 */
public interface FSubscriberTransport {

    /**
     * Queries whether the transport is subscribed to a topic.
     *
     * @return True if the transport is subscribed to a topic.
     */
    boolean isSubscribed();

    /**
     * Opens the Transport to receive messages on the subscription.
     *
     * @param topic the pub/sub topic to subscribe to.
     * @throws TException if there was a problem subscribing.
     */
    void subscribe(String topic, FAsyncCallback callback) throws TException;

    /**
     * Closes the transport by unsubscribing from the set topic.
     */
    void unsubscribe();

    /**
     * Remove unsubscribes and removes durably stored information on the broker, if applicable.
     */
    default void remove() throws TException {
        unsubscribe();
    }
}
