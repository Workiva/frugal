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

package com.workiva.frugal.server;

import org.apache.thrift.TException;

/**
 * FServer is Frugal's equivalent of Thrift's TServer. It's used to run a Frugal
 * RPC service by executing an FProcessor on client connections. FServer can
 * optionally support a high-water mark which is the maximum amount of time a
 * request is allowed to be enqueued before triggering server overload logic (e.g.
 * load shedding).
 * <p>
 * Currently, Frugal includes two implementations of FServer: FSimpleServer, which
 * is a basic, accept-loop based server that supports traditional Thrift
 * TServerTransports, and FNatsServer, which is an implementation that uses NATS
 * as the underlying transport.
 */
public interface FServer {

    /**
     * Starts the server.
     *
     * @throws TException
     */
    void serve() throws TException;

    /**
     * Stops the server. This is optional on a per-implementation basis.
     * Not all servers are required to be cleanly stoppable.
     *
     * @throws TException
     */
    void stop() throws TException;
}
