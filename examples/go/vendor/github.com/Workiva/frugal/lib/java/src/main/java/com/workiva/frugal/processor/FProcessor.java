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

package com.workiva.frugal.processor;

import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TException;

import java.util.Map;

/**
 * FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic object
 * which operates upon an input stream and writes to an output stream.
 * Specifically, an FProcessor is provided to an FServer in order to wire up a
 * service handler to process requests.
 */
public interface FProcessor {

    /**
     * Processes the request from the input protocol and write the response to the output protocol.
     *
     * @param in  input FProtocol
     * @param out output FProtocol
     * @throws TException if issues processing requests occur
     */
    void process(FProtocol in, FProtocol out) throws TException;

    /**
     * Adds the given ServiceMiddleware to the FProcessor. This should only be called before the server is started.
     *
     * @param middleware the ServiceMiddleware to add
     */
    void addMiddleware(ServiceMiddleware middleware);

    /**
     * Returns a map of method name to annotations as defined in the service IDL that is serviced by this processor.
     *
     * @return Map of method name to annotations as defined by the IDL
     */
    Map<String, Map<String, String>> getAnnotations();
}
