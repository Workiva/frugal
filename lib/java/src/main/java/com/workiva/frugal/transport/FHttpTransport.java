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


import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TTransportExceptionType;
import org.apache.commons.codec.binary.Base64;
import org.apache.http.HttpEntity;
import org.apache.http.HttpStatus;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.conn.ConnectTimeoutException;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.util.EntityUtils;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.net.SocketTimeoutException;
import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.Map;


/**
 * FHttpTransport extends FTransport. This is a "stateless" transport in the
 * sense that this transport is not persistently connected to a single server.
 * A request is simply an http request and a response is an http response.
 * This assumes requests/responses fit within a single http request.
 */
public class FHttpTransport extends FTransport {
    // Logger
    private static final Logger LOGGER = LoggerFactory.getLogger(FHttpTransport.class);

    // Immutable
    private final CloseableHttpClient httpClient;
    private final String url;
    private final int responseSizeLimit;
    private final FHttpTransportHeaders requestHeaders;

    private FHttpTransport(CloseableHttpClient httpClient, String url, int requestSizeLimit, int responseSizeLimit,
            FHttpTransportHeaders requestHeaders) {
        super();
        this.httpClient = httpClient;
        this.url = url;
        this.requestSizeLimit = requestSizeLimit;
        this.responseSizeLimit = responseSizeLimit;
        this.requestHeaders = requestHeaders;
    }

    /**
     * Interface that returns a Map of HTTP request headers.
     */
    public interface FHttpTransportHeaders {

        /**
         * Returns a Map of HTTP request headers.
         *
         * @return Map of HTTP request headers.
         */
        public Map<String, String> getRequestHeaders();
    }

    /**
     * Builder for configuring and construction FHttpTransport instances.
     */
    public static class Builder {
        private final CloseableHttpClient httpClient;
        private final String url;
        private int requestSizeLimit;
        private int responseSizeLimit;
        private FHttpTransportHeaders requestHeaders;

        /**
         * Create a new Builder which create FHttpTransports that communicate with a server
         * at the given url.
         *
         * @param httpClient HTTP client
         * @param url        Server URL
         */
        public Builder(CloseableHttpClient httpClient, String url) {
            this.httpClient = httpClient;
            this.url = url;
        }

        /**
         * Adds a request size limit to the Builder. If non-positive, there will
         * be no request size limit (the default behavior).
         *
         * @param requestSizeLimit Size limit for outgoing requests.
         * @return Builder
         */
        public Builder withRequestSizeLimit(int requestSizeLimit) {
            this.requestSizeLimit = requestSizeLimit;
            return this;
        }

        /**
         * Adds a response size limit to the Builder. If non-positive, there will
         * be no response size limit (the default behavior).
         *
         * @param responseSizeLimit Size limit for incoming responses.
         * @return Builder
         */
        public Builder withResponseSizeLimit(int responseSizeLimit) {
            this.responseSizeLimit = responseSizeLimit;
            return this;
        }

        /**
         * Adds HTTP request headers to the builder.
         *
         * @param requestHeaders Map of HTTP request headers to add to request.
         * @return Builder
         */
        public Builder withRequestHeaders(FHttpTransportHeaders requestHeaders) {
            this.requestHeaders = requestHeaders;
            return this;
        }

        /**
         * Creates new configured FHttpTransport.
         *
         * @return FHttpTransport
         */
        public FHttpTransport build() {
            return new FHttpTransport(this.httpClient, this.url,
                    this.requestSizeLimit, this.responseSizeLimit,
                    this.requestHeaders);
        }
    }

    /**
     * Queries whether the transport is open.
     *
     * @return True
     */
    @Override
    public boolean isOpen() {
        return true;
    }

    /**
     * This is a no-op for FHttpTransport.
     */
    @Override
    public void open() throws TTransportException {
    }

    /**
     * This is a no-op for FHttpTransport.
     */
    @Override
    public void close() {
    }

    /**
     * Sends the framed frugal payload over HTTP.
     *
     * @throws TTransportException if there was an error writing out data.
     */
    @Override
    public void oneway(FContext context, byte[] payload) throws TTransportException {
        preflightRequestCheck(payload.length);

        makeRequest(context, payload);
    }

    /**
     * Sends the framed frugal payload over HTTP.
     *
     * @throws TTransportException if there was an error writing out data.
     */
    @Override
    public TTransport request(FContext context, byte[] payload) throws TTransportException {
        preflightRequestCheck(payload.length);

        byte[] response = makeRequest(context, payload);

        // All responses should be framed with 4 bytes
        if (response.length < 4) {
            throw new TTransportException("invalid frame size");
        }

        // If there are only 4 bytes, this needs to be a one-way
        // (i.e. frame size 0)
        if (response.length == 4) {
            if (ByteBuffer.wrap(response).getInt() != 0) {
                throw new TTransportException("missing data");
            }
            return null;
        }
        return new TMemoryInputTransport(Arrays.copyOfRange(response, 4, response.length));
    }

    private byte[] makeRequest(FContext context, byte[] requestPayload) throws TTransportException {
        // Encode request payload
        String encoded = Base64.encodeBase64String(requestPayload);
        StringEntity requestEntity = new StringEntity(encoded, ContentType.create("application/x-frugal", "utf-8"));

        // Set headers and payload
        HttpPost request = new HttpPost(url);

        // add user supplied headers first, to avoid monkeying
        // with the size limits headers below.
        if (requestHeaders != null) {
            for (Map.Entry<String, String> entry : requestHeaders.getRequestHeaders().entrySet()) {
                String key = entry.getKey();
                String value = entry.getValue();
                if (key != null && value != null) {
                    request.setHeader(key, value);
                }
            }
        }

        request.setHeader("accept", "application/x-frugal");
        request.setHeader("content-transfer-encoding", "base64");
        if (responseSizeLimit > 0) {
            request.setHeader("x-frugal-payload-limit", Integer.toString(responseSizeLimit));
        }
        request.setEntity(requestEntity);
        request.setConfig(RequestConfig.custom()
                .setConnectTimeout((int) context.getTimeout())
                .setSocketTimeout((int) context.getTimeout())
                .build());

        // Make request
        CloseableHttpResponse response;
        try {
            response = httpClient.execute(request);
        } catch (ConnectTimeoutException | SocketTimeoutException e) {
            throw new TTransportException(TTransportExceptionType.TIMED_OUT,
                    "http request timed out: " + e.getMessage());
        } catch (IOException e) {
            throw new TTransportException("http request failed: " + e.getMessage());
        }

        try {
            // Response too large
            int status = response.getStatusLine().getStatusCode();
            if (status == HttpStatus.SC_REQUEST_TOO_LONG) {
                throw new TTransportException(
                        TTransportExceptionType.RESPONSE_TOO_LARGE, "response was too large for the transport");
            }

            // Decode body
            String responseBody = "";
            HttpEntity responseEntity = response.getEntity();
            if (responseEntity != null) {
                responseBody = EntityUtils.toString(responseEntity, "utf-8");
            }
            // Check bad status code
            if (status >= 300) {
                throw new TTransportException("response errored with code " + status + " and message " + responseBody);
            }
            // Decode and return response body
            return Base64.decodeBase64(responseBody);

        } catch (IOException e) {
            throw new TTransportException("could not decodeFromFrame response body: " + e.getMessage());
        } finally {
            try {
                response.close();
            } catch (IOException e) {
                LOGGER.warn("could not close server response: " + e.getMessage());
            }
        }
    }
}
