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

import com.workiva.frugal.FContext;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TException;

/**
 * FProcessorFunction is used internally by generated code. An FProcessor
 * registers an FProcessorFunction for each service method. Like FProcessor, an
 * FProcessorFunction exposes a single process call, which is used to handle a
 * method invocation.
 */
public interface FProcessorFunction {

    /**
     * Processes the request from the input protocol and write the response to the output protocol.
     *
     * @param ctx FContext
     * @param in  input FProtocol
     * @param out output FProtocol
     * @throws TException
     */
    void process(FContext ctx, FProtocol in, FProtocol out) throws TException;

}
