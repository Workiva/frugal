package com.workiva.frugal;

import org.jose4j.jwk.HttpsJwks;
import org.jose4j.jwt.JwtClaims;
import org.jose4j.jwt.consumer.JwtConsumer;
import org.jose4j.jwt.consumer.JwtConsumerBuilder;
import org.jose4j.jwt.consumer.JwtContext;
import org.jose4j.keys.resolvers.HttpsJwksVerificationKeyResolver;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;


/**
 * FContext is the message context for a frugal message. This object is not thread-safe.
 */
public class FContext {

    private static final String CID = "_cid";
    private static final String OP_ID = "_opid";
    private static final AtomicLong NEXT_OP_ID = new AtomicLong();
    private static final long DEFAULT_TIMEOUT = 60 * 1000;

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
        requestHeaders.put(OP_ID, Long.toString(NEXT_OP_ID.getAndIncrement()));
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
        if (headers.get(OP_ID) == null) {
            headers.put(OP_ID, Long.toString(NEXT_OP_ID.getAndIncrement()));
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
     * Returns the operation id for the FContext. This is a unique long per operation.
     *
     * @return operation id
     */
    public long getOpId() {
        String opIdStr = requestHeaders.get(OP_ID);
        if (opIdStr == null) {
            throw new RuntimeException("opId is null!");
        }
        return Long.valueOf(opIdStr);
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

    // Auth related stuff

    private static JwtConsumer jwtConsumer;
    private Long membershipId = null;
    private String membershipRid = null;
    private Long userId = null;
    private String userRid = null;
    private Long accountId = null;
    private String accountRid = null;
    final private static String
            AUTH_HEADER_NAME = "Authorization",
            AUTH_HEADER_PREFIX = "Bearer ";


    private String getAuthToken() throws Exception {
        String authToken = requestHeaders.get(AUTH_HEADER_NAME);
        if (authToken == null || authToken.equals("")) {
            throw new Exception(AUTH_HEADER_NAME + " header not present");
        }
        if (authToken.equals("")) {
            throw new Exception(AUTH_HEADER_NAME + " header is empty");
        }
        return authToken.replaceAll("^" + AUTH_HEADER_PREFIX, "");
    }

    /**
     * Sets the auth token on a context and returns itself
     */
    public FContext setAuthToken(String authToken) {
        this.requestHeaders.put(AUTH_HEADER_NAME, AUTH_HEADER_PREFIX + authToken);
        return this;
    }

    private static synchronized JwtConsumer getJwtConsumer() throws Exception {
        if (jwtConsumer == null) {
            String authHost = System.getenv("OAUTH2_HOST");
            if (authHost == null || authHost.equals("")) {
                throw new Exception("Env OAUTH2_HOST must be declared, e.g. https://wk-dev.wdesk.org");
            }
            authHost = authHost.replaceAll("/$", ""); // Remove trailing slash
            jwtConsumer = new JwtConsumerBuilder()
                    .setExpectedIssuer(authHost)
                    .setVerificationKeyResolver(new HttpsJwksVerificationKeyResolver(new HttpsJwks(authHost + "/iam/oauth2/v2.0/certs")))
                    .setRequireExpirationTime()
                    .setAllowedClockSkewInSeconds(10)
                    .setExpectedAudience(false)
                    .setSkipDefaultAudienceValidation()
                    .build();
        }
        return jwtConsumer;
    }

    /**
     * Verifies that a context's access token has a valid signature, issuer, timestamp, scopes, etc.
     */
    public FContext verifyAuthToken(String... requiredScopes) throws Exception {
        // Check the signature, issuer, expiration
        JwtContext jwtContext = getJwtConsumer().process(getAuthToken());
        JwtClaims jwtClaims = jwtContext.getJwtClaims();

        // Check the version
        if (jwtClaims.getClaimValue("ver", Long.class) < 2) {
            throw new Exception("Token version must be at least 2");
        }

        // Check scopes
        List<String> possessedScopes = (ArrayList<String>) jwtClaims.getClaimValue("scp");
        for (String requiredScope : requiredScopes) {
            boolean hasRequiredScope = false;
            for (String possessedScope : possessedScopes) {
                if (possessedScope.equals(requiredScope)) {
                    hasRequiredScope = true;
                    break;
                }
            }
            if (!hasRequiredScope) {
                throw new Exception(String.format("Required scope (%s) not in list (%s)", requiredScope, possessedScopes));
            }
        }

        // Set the appropriate fields from the claims
        membershipId = jwtClaims.getClaimValue("mem", Long.class);
        membershipRid = jwtClaims.getClaimValue("mrid", String.class);
        userId = jwtClaims.getClaimValue("usr", Long.class);
        userRid = jwtClaims.getClaimValue("urid", String.class);
        accountId = jwtClaims.getClaimValue("acc", Long.class);
        accountRid = jwtClaims.getClaimValue("arid", String.class);

        return this;
    }

    public Long getMembershipId() {
        return membershipId;
    }

    public String getMembershipRid() {
        return membershipRid;
    }

    public Long getUserId() {
        return userId;
    }

    public String getUserRid() {
        return userRid;
    }

    public Long getAccountId() {
        return accountId;
    }

    public String getAccountRid() {
        return accountRid;
    }
}
