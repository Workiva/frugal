package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.concurrent.ArrayBlockingQueue;
import java.io.ByteArrayOutputStream;

import org.apache.commons.codec.binary.Base64;
import org.apache.http.HttpEntity;
import org.apache.http.HttpStatus;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.util.EntityUtils;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;


/**
 * THttpTransport extends thrift.TTransport. This is a "stateless"
 * transport in the sense that this transport is not persistently connected to
 * a single server. A request is simply an http request and a response is an
 * http response. This assumes requests/responses fit within a single http
 * request.
 */
public class THttpTransport extends TTransport {

    // Controls how many responses to buffer
    private static final int FRAME_BUFFER_SIZE = 5;

    // Frame buffer poison pill
    protected static final byte[] FRAME_BUFFER_CLOSED = new byte[0];

    // Logger
    private static final Logger LOGGER = LoggerFactory.getLogger(THttpTransport.class);

    // Immutable
    private final CloseableHttpClient httpClient;
    private final String url;
    private final int requestSizeLimit;
    private final int responseSizeLimit;
    protected final ArrayBlockingQueue<byte[]> frameBuffer;
    private final ByteArrayOutputStream requestBuffer;

    // Mutable
    private byte[] currentFrame = new byte[0];
    private int readPos;
    private boolean isOpen;

    private THttpTransport(CloseableHttpClient httpClient, String url,
                          int requestSizeLimit, int responseSizeLimit) {
        this.httpClient = httpClient;
        this.url = url;
        this.requestSizeLimit = requestSizeLimit;
        this.responseSizeLimit = responseSizeLimit;
        this.frameBuffer = new ArrayBlockingQueue<>(FRAME_BUFFER_SIZE);
        this.requestBuffer = new ByteArrayOutputStream();
    }

    /**
     * Builder for configuring and construction THttpTransport instances.
     */
    public static class Builder {
        private final CloseableHttpClient httpClient;
        private final String url;
        private int requestSizeLimit;
        private int responseSizeLimit;

        /**
         * Create a new Builder which create THttpTransports that communicate with a server
         * at the given url.
         *
         * @param httpClient    HTTP client
         * @param url           Server URL
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
         * Creates new configured THttpTransport.
         *
         * @return THttpTransport
         */
        public THttpTransport build() {
             return new THttpTransport(this.httpClient, this.url,
                     this.requestSizeLimit, this.responseSizeLimit);
        }
    }

    /**
     * Queries whether the transport is open.
     *
     * @return True if the transport is open.
     */
    @Override
    public synchronized boolean isOpen() {
        return isOpen;
    }

    /**
     * Opens the transport for reading/writing.
     *
     * @throws TTransportException if the transport could not be opened
     */
    @Override
    public void open() throws TTransportException {
        // Make sure the frame buffer is completely empty (in case this
        // transport was re-opened).
        frameBuffer.clear();
        isOpen = true;
    }

    /**
     * Clears and puts the close frame into the response buffer
     */
    @Override
    public synchronized void close() {
        frameBuffer.clear();
        try {
            frameBuffer.put(FRAME_BUFFER_CLOSED);
        } catch (InterruptedException e) {
            LOGGER.warn("HTTP transport could not close frame buffer: " + e.getMessage());
        }
        isOpen = false;
    }

    /**
     * Reads up to len bytes into buffer buf, starting at offset off.
     *
     * @param buf Array to read into
     * @param off Index to start reading at
     * @param len Maximum number of bytes to read
     * @return The number of bytes actually read
     * @throws TTransportException if there was an error reading data
     */
    @Override
    public int read(byte[] buf, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.NOT_OPEN, "read: HTTP transport not open");
        }

        if (readPos == currentFrame.length) {
            try {
                currentFrame = frameBuffer.take();
            } catch (InterruptedException e) {
                throw new TTransportException(TTransportException.END_OF_FILE, e.getMessage());
            }
            readPos = 0;
        }

        if (currentFrame == FRAME_BUFFER_CLOSED) {
            throw new TTransportException(TTransportException.END_OF_FILE);
        }

        // Can only copy bytes that are available in the current buffer
        len = Math.min(currentFrame.length - readPos, len);

        System.arraycopy(currentFrame, readPos, buf, off, len);
        readPos += len;

        return len;
    }

    /**
     * Writes the bytes to a buffer. Throws FMessageSizeException if the buffer exceeds
     * {@code requestSizeLimit} bytes.
     *
     * @param buf The output data buffer
     * @param off The offset to start writing from
     * @param len The number of bytes to write
     * @throws TTransportException if there was an error writing data
     */
    @Override
    public void write(byte[] buf, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.NOT_OPEN, "write: HTTP transport not open");
        }

        if (requestSizeLimit > 0 && len + requestBuffer.size() > requestSizeLimit) {
            int size = len + requestBuffer.size();
            requestBuffer.reset();
            throw new FMessageSizeException(
                    String.format("Message exceeds %d bytes, was %d bytes", requestSizeLimit, size));
        }
        requestBuffer.write(buf, off, len);
    }

    /**
     * Sends the buffered bytes over HTTP.
     *
     * @throws TTransportException if there was an error writing out data.
     */
    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.NOT_OPEN, "flush: HTTP transport not open");
        }

        if (requestBuffer.size() == 0) {
            return;
        }

        byte[] data = requestBuffer.toByteArray();
        requestBuffer.reset();

        byte[] response = makeRequest(data);

        // All responses should be framed with 4 bytes
        if (response.length < 4) {
            throw new TTransportException(TTransportException.UNKNOWN, "invalid frame size");
        }

        // If there are only 4 bytes, this needs to be a one-way
        // (i.e. frame size 0)
        if (response.length == 4) {
            if (ByteBuffer.wrap(response).getInt() != 0) {
                throw new TTransportException(TTransportException.UNKNOWN, "missing data");
            }
        }

        // Put the frame in the buffer
        try {
            frameBuffer.put(response);
        } catch (InterruptedException e) {
            throw new TTransportException(TTransportException.UNKNOWN,
                    "could not put frame in buffer: " + e.getMessage());
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
            throw new TTransportException(TTransportException.UNKNOWN,
                    "http request failed: " + e.getMessage());
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
                throw new TTransportException(TTransportException.UNKNOWN,
                        "response errored with code " + status + " and message " + responseBody);
            }
            // Decode and return response body
            return Base64.decodeBase64(responseBody);

        } catch (IOException e) {
            throw new TTransportException(TTransportException.UNKNOWN,
                    "could not decode response body: " + e.getMessage());
        } finally {
            try {
                response.close();
            } catch (IOException e) {
                LOGGER.warn("could not close server response: " + e.getMessage());
            }
        }
    }
}
