package com.workiva.frugal.protocol;

import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;


/**
 * FContext is the context for a Frugal message. Every RPC has an FContext, which
 * can be used to set request headers, response headers, and the request timeout.
 * The default timeout is one minute. An FContext is also sent with every publish
 * message which is then received by subscribers.
 * <p/>
 * In addition to headers, the FContext also contains a correlation ID which can
 * be used for distributed tracing purposes. A random correlation ID is generated
 * for each FContext if one is not provided.
 * <p/>
 * FContext also plays a key role in Frugal's multiplexing support. A unique,
 * per-request operation ID is set on every FContext before a request is made.
 * This operation ID is sent in the request and included in the response, which is
 * then used to correlate a response to a request. The operation ID is an internal
 * implementation detail and is not exposed to the user.
 * <p/>
 * This object is not thread-safe.
 */
public class FContext {

    protected static final String CID = "_cid";
    protected static final String OP_ID = "_opid";
    protected static final long DEFAULT_TIMEOUT = 60 * 1000;

    private Map<String, String> requestHeaders = new ConcurrentHashMap<>();
    private Map<String, String> responseHeaders = new ConcurrentHashMap<>();

    private volatile long timeout = DEFAULT_TIMEOUT;

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
        requestHeaders.put(CID, correlationId);
        requestHeaders.put(OP_ID, "0");

    }

    /**
     * Creates a new FContext with the given request headers.
     *
     * @param headers request headers
     * @return FContext
     */
    protected static FContext withRequestHeaders(Map<String, String> headers) {
        if (headers.get(CID) == null) {
            headers.put(CID, generateCorrelationId());
        }
        return new FContext(headers, new HashMap<String, String>());
    }

    /**
     * Returns the correlation id for the FContext. This is used for distributed-tracing purposes.
     *
     * @return correlation id
     */
    public String getCorrelationId() {
        return requestHeaders.get(CID);
    }

    /**
     * Returns the operation id for the FContext. This is a unique long per operation. This is protected as operation
     * ids are an internal implementation detail.
     *
     * @return operation id
     */
    protected long getOpId() {
        String opIdStr = requestHeaders.get(OP_ID);
        if (opIdStr == null) {
            return 0;
        }
        return Long.valueOf(opIdStr);
    }

    /**
     * Sets the operation id on the FContext. The operation id is used to map responses to requests. This is protected
     * as operation ids are an internal implementation detail.
     *
     * @param opId the operation id to set
     */
    protected void setOpId(long opId) {
        requestHeaders.put(OP_ID, Long.toString(opId));
    }

    /**
     * Adds a request header to the FContext for the given name. A header is a key-value pair. If a header with the name
     * is already present on the FContext, it will be replaced. The _opid and _cid headers are reserved.
     *
     * @param name  header name
     * @param value header value
     */
    public void addRequestHeader(String name, String value) {
        if (OP_ID.equals(name) || CID.equals(name)) {
            return;
        }
        requestHeaders.put(name, value);
    }

    /**
     * Adds request headers to the FContext for the given headers map. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced. The _opid
     * and _cid headers are reserved.
     *
     * @param headers headers to add to request headers
     */
    public void addRequestHeaders(Map<String, String> headers) {
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            addRequestHeader(pair.getKey(), pair.getValue());
        }
    }

    /**
     * Adds a response header to the FContext for the given name. A header is a key-value pair. If a header with the name
     * is already present on the FContext, it will be replaced. The _opid header is reserved.
     *
     * @param name  header name
     * @param value header value
     */
    public void addResponseHeader(String name, String value) {
        if (OP_ID.equals(name)) {
            return;
        }
        responseHeaders.put(name, value);
    }

    /**
     * Adds response headers to the FContext for the given headers map. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced. The _opid
     * header is reserved.
     *
     * @param headers headers to add to request headers
     */
    public void addResponseHeaders(Map<String, String> headers) {
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            addResponseHeader(pair.getKey(), pair.getValue());
        }
    }

    /**
     * Adds response headers to the FContext for the given headers map. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced.
     *
     * @param headers headers to add to request headers
     */
    protected void forceAddResponseHeaders(Map<String, String> headers) {
        responseHeaders.putAll(headers);
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
     * Set the request timeout. Default is 1 minute.
     *
     * @param timeout timeout for the request in milliseconds.
     */
    public void setTimeout(long timeout) {
        this.timeout = timeout;
    }

    /**
     * Get the request timeout.
     *
     * @return the request timeout in milliseconds.
     */
    public long getTimeout() {
        return this.timeout;
    }

    protected void setResponseOpId(String opId) {
        responseHeaders.put(OP_ID, opId);
    }

    private static String generateCorrelationId() {
        return UUID.randomUUID().toString().replace("-", "");
    }
}
