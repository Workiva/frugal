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

package com.workiva.frugal;

import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

/**
 * FContext is the context for a Frugal message. Every RPC has an FContext, which
 * can be used to set request headers, response headers, and the request timeout.
 * The default timeout is five seconds. An FContext is also sent with every publish
 * message which is then received by subscribers.
 * An FContext can also contain a map of properties that won't be serialized,
 * called ephemeralProperties.
 * <p>
 * As a best practice, the request headers of an inbound FContext should not be
 * modified, and outbound FContext instances should not be reused.  Instead,
 * FContext instances should be cloned before each outbound call.
 * <p>
 * In addition to headers, the FContext also contains a correlation ID which can
 * be used for distributed tracing purposes. A random correlation ID is generated
 * for each FContext if one is not provided.
 * <p>
 * FContext also plays a key role in Frugal's multiplexing support. A unique,
 * per-request operation ID is set on every FContext before a request is made.
 * This operation ID is sent in the request and included in the response, which is
 * then used to correlate a response to a request. The operation ID is an internal
 * implementation detail and is not exposed to the user.
 * <p>
 * This object is not thread-safe.
 */
public class FContext implements Cloneable {

    /**
     * To ensure every new FContext gets a unique opid, use an atomic, incrementing integer.
     */
    private static final AtomicLong NEXT_OP_ID = new AtomicLong(0);

    /**
     * Header containing correlation id.
     */
    public static final String CID_HEADER = "_cid";

    /**
     * Header containing op id (uint64 as string).
     */
    public static final String OPID_HEADER = "_opid";

    /**
     * Header containing request timeout (milliseconds as string).
     */
    protected static final String TIMEOUT_HEADER = "_timeout";

    /**
     * Default request timeout.
     */
    protected static final long DEFAULT_TIMEOUT = 5 * 1000;

    private Map<String, String> requestHeaders = new ConcurrentHashMap<>();
    private Map<String, String> responseHeaders = new ConcurrentHashMap<>();
    private Map<Object, Object> ephemeralProperties = new HashMap<>();

    private FContext(Map<String, String> requestHeaders, Map<String, String> responseHeaders) {
        this.requestHeaders = requestHeaders;
        this.responseHeaders = responseHeaders;
    }

    /**
     * Creates a new FContext with a randomly generated correlation id for tracing purposes.
     */
    public FContext() {
        this(generateCorrelationId());
    }

    /**
     * Creates a new FContext with the given correlation id for tracing purposes.
     *
     * @param correlationId unique tracing identifier
     */
    public FContext(String correlationId) {
        requestHeaders.put(CID_HEADER, correlationId);
        requestHeaders.put(OPID_HEADER, getNextOpId());
        requestHeaders.put(TIMEOUT_HEADER, Long.toString(DEFAULT_TIMEOUT));
    }

    /**
     * Creates a new FContext with the given request headers.
     *
     * @param headers request headers
     * @return FContext
     */
    public static FContext withRequestHeaders(Map<String, String> headers) {
        headers.computeIfAbsent(CID_HEADER, k -> generateCorrelationId());
        headers.computeIfAbsent(TIMEOUT_HEADER, k -> Long.toString(DEFAULT_TIMEOUT));
        // Always generate a new opid as it has to be unique to the context
        headers.put(OPID_HEADER, getNextOpId());
        return new FContext(headers, new HashMap<>());
    }

    /**
     * Returns a new unique opid. Should not be used by consumers outside of
     * frugal.
     *
     * @return A new unique opid.
     */
    public static String getNextOpId() {
        return Long.toString(NEXT_OP_ID.getAndIncrement());
    }

    private static String generateCorrelationId() {
        return UUID.randomUUID().toString().replace("-", "");
    }

    /**
     * Returns the correlation id for the FContext. This is used for distributed-tracing purposes.
     *
     * @return correlation id
     */
    public String getCorrelationId() {
        return requestHeaders.get(CID_HEADER);
    }

    /**
     * Adds a request header to the FContext for the given name. A header is a key-value pair. If a header with the name
     * is already present on the FContext, it will be replaced. The _opid header is reserved and setting it could
     * interfere with asynchronous transports that rely upon it. Calls to set the _cid header will be ignored.
     * Returns the same FContext to allow for call chaining.
     *
     * @param name  header name
     * @param value header value
     * @return FContext
     */
    public FContext addRequestHeader(String name, String value) {
        requestHeaders.put(name, value);
        return this;
    }

    /**
     * Adds request headers to the FContext for the given headers map. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced.
     * The _opid header is reserved and setting it could interfere with asynchronous transports that rely upon it.
     * Calls to set the _cid header will be ignored. Returns the same FContext to allow for call chaining.
     *
     * @param headers headers to add to request headers
     * @return FContext
     */
    public FContext addRequestHeaders(Map<String, String> headers) {
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            addRequestHeader(pair.getKey(), pair.getValue());
        }
        return this;
    }

    /**
     * Removes a request header from the FContext for the given name.
     * The _opid header is reserved and removing it could interfere with asynchronous transports that rely upon it.
     * Returns the same FContext to allow for call chaining.
     *
     * @param name header name
     * @return FContext
     */
    public FContext removeRequestHeader(String name) {
        requestHeaders.remove(name);
        return this;
    }

    /**
     * Adds a response header to the FContext for the given name. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced.
     * The _opid header is reserved and setting it could interfere with asynchronous transports that rely upon it.
     * Returns the same FContext to allow for call chaining.
     *
     * @param name  header name
     * @param value header value
     * @return FContext
     */
    public FContext addResponseHeader(String name, String value) {
        responseHeaders.put(name, value);
        return this;
    }

    /**
     * Adds response headers to the FContext for the given headers map. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced. The _opid
     * header is reserved. Returns the same FContext to allow for call chaining.
     *
     * @param headers headers to add to request headers
     * @return FContext
     */
    public FContext addResponseHeaders(Map<String, String> headers) {
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            addResponseHeader(pair.getKey(), pair.getValue());
        }
        return this;
    }

    /**
     * Removes a response header from the FContext for the given name.
     * The _opid header is reserved and removing it could interfere with asynchronous transports that rely upon it.
     * Returns the same FContext to allow for call chaining.
     *
     * @param name header name
     * @return FContext
     */
    public FContext removeResponseHeader(String name) {
        responseHeaders.remove(name);
        return this;
    }

    /**
     * Returns the request header with the given name. If no such header exists, null is returned.
     *
     * @param name header name
     * @return header value or null if it doesn't exist
     */
    public String getRequestHeader(String name) {
        return requestHeaders.get(name);
    }

    /**
     * Returns the response header with the given name. If no such header exists, null is returned.
     *
     * @param name header name
     * @return header value or null if it doesn't exist
     */
    public String getResponseHeader(String name) {
        return responseHeaders.get(name);
    }

    /**
     * Returns the request headers on the FContext.
     *
     * @return request headers map
     */
    public Map<String, String> getRequestHeaders() {
        return new HashMap<>(requestHeaders);
    }

    /**
     * Returns the response headers on the FContext.
     *
     * @return response headers map
     */
    public Map<String, String> getResponseHeaders() {
        return new HashMap<>(responseHeaders);
    }

    /**
     * Sets the given map as the ephemeral properties for the FContext.
     * A reference to the map will be kept, so modifications made to the map
     * outside this object will affect the ephemeral properties in this object.
     *
     * @param ephemeralProperties map to set as ephemeral properties
     * @return FContext
     */
    public FContext setEphemeralProperties(Map<Object, Object> ephemeralProperties) {
        this.ephemeralProperties = ephemeralProperties;
        return this;
    }

    /**
     * Adds an ephemeral property to the FContext. An ephemeral property is a key-value
     * pair that will not be serialized with the FContext. If an ephemeral property
     * with the same name is already present, it will be replaced. The same FContext
     * is returned to allow for call chaining.
     *
     * @param key ephemeral property name
     * @param value ephemeral property value
     * @return FContext
     */
    public FContext addEphemeralProperty(Object key, Object value) {
        ephemeralProperties.put(key, value);
        return this;
    }

    /**
     * Adds the given ephemeral properties to the FContext. An ephemeral property is
     * a key-value pair that will not be serialized with the FContext. If an ephemeral
     * property with the same name is already present, it will be replaced. The same
     * FContext is returned to allow for call chaining.
     *
     * @param properties ephemeral properties to add
     * @return FContext
     */
    public FContext addEphemeralProperties(Map<Object, Object> properties) {
        ephemeralProperties.putAll(properties);
        return this;
    }

    /**
     * Returns the ephemeral properties on the FContext. Ephemeral properties are
     * key-value pairs that will not be serialized with the FContext.
     *
     * @return ephemeral properties
     */
    public Map<Object, Object> getEphemeralProperties() {
        return new HashMap<>(ephemeralProperties);
    }

    /**
     * Returns the ephemeral property from the FContext for the corresponding key,
     * Or null if it doesn't exist. An ephemeral property is a key-value pair that
     * will not be serialized with the FContext.
     *
     * @param key ephemeral property name
     * @return ephemeral property value
     */
    public Object getEphemeralProperty(Object key) {
        return ephemeralProperties.get(key);
    }

    /**
     * Get the request timeout.
     *
     * @return the request timeout in milliseconds.
     */
    public long getTimeout() {
        return Long.parseLong(requestHeaders.getOrDefault(TIMEOUT_HEADER, Long.toString(DEFAULT_TIMEOUT)));
    }

    /**
     * Set the request timeout. Default is 5 seconds.
     *
     * @param timeout timeout for the request in milliseconds.
     */
    public void setTimeout(long timeout) {
        requestHeaders.put(TIMEOUT_HEADER, Long.toString(timeout));
    }

    @Override
    public FContext clone() throws CloneNotSupportedException {
        FContext cloned = (FContext) super.clone();
        cloned.requestHeaders = this.getRequestHeaders();
        cloned.responseHeaders = this.getResponseHeaders();
        cloned.ephemeralProperties = this.getEphemeralProperties();
        cloned.addRequestHeader(OPID_HEADER, getNextOpId());
        return cloned;
    }
}
