package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;

import org.apache.commons.codec.binary.Base64;
import org.apache.http.HttpEntity;
import org.apache.http.HttpStatus;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.util.EntityUtils;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.ByteBuffer;


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

    private FHttpTransport(CloseableHttpClient httpClient, String url,
                           int requestSizeLimit, int responseSizeLimit) {
        super(requestSizeLimit - 4);
        this.httpClient = httpClient;
        this.url = url;
        this.responseSizeLimit = responseSizeLimit;
    }

    /**
     * Builder for configuring and construction FHttpTransport instances.
     */
    public static class Builder {
        private final CloseableHttpClient httpClient;
        private final String url;
        private int requestSizeLimit;
        private int responseSizeLimit;

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
         * Creates new configured FHttpTransport.
         *
         * @return FHttpTransport
         */
        public FHttpTransport build() {
            return new FHttpTransport(this.httpClient, this.url,
                    this.requestSizeLimit, this.responseSizeLimit);
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
     * Sends the buffered bytes over HTTP.
     *
     * @throws TTransportException if there was an error writing out data.
     */
    @Override
    public void flush() throws TTransportException {
        if (!hasWriteData()) {
            return;
        }
        byte[] data = getFramedWriteBytes();
        resetWriteBuffer();
        byte[] response = makeRequest(data);

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
            return;
        }

        // Put the frame in the buffer
        try {
            executeFrame(response);
        } catch (TException e) {
            throw new TTransportException("could not execute response callback: " + e.getMessage());
        }
    }

    private byte[] makeRequest(byte[] requestPayload) throws TTransportException {
        // Encode request payload
        String encoded = Base64.encodeBase64String(requestPayload);
        StringEntity requestEntity = new StringEntity(encoded, ContentType.create("application/x-frugal", "utf-8"));

        // Set headers and payload
        HttpPost request = new HttpPost(url);
        request.setHeader("accept", "application/x-frugal");
        request.setHeader("content-transfer-encoding", "base64");
        if (responseSizeLimit > 0) {
            request.setHeader("x-frugal-payload-limit", Integer.toString(responseSizeLimit));
        }
        request.setEntity(requestEntity);

        // Make request
        CloseableHttpResponse response;
        try {
            response = httpClient.execute(request);
        } catch (IOException e) {
            throw new TTransportException("http request failed: " + e.getMessage());
        }

        try {
            // Response too large
            int status = response.getStatusLine().getStatusCode();
            if (status == HttpStatus.SC_REQUEST_TOO_LONG) {
                throw new FMessageSizeException(FTransport.RESPONSE_TOO_LARGE,
                        "response was too large for the transport");
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
