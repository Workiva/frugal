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
        FContext clone = (FContext) super.clone();
        clone.requestHeaders = this.getRequestHeaders();
        clone.responseHeaders = this.getResponseHeaders();
        clone.addRequestHeader(OPID_HEADER, getNextOpId());
        return clone;
    }
}
