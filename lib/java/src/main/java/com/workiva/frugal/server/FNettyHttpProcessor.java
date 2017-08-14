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

import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;

/**
 * Processes a {@link FullHttpRequest} to return a {@link FullHttpResponse}.
 */
public interface FNettyHttpProcessor {

    /**
     * Process an HTTP request and return an HTTP response.
     *
     * @param request an HTTP request
     * @return a valid Frugal HTTP response
     */
    FullHttpResponse process(FullHttpRequest request);
}
