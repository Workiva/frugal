package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.commons.codec.binary.Base64;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Map;

/**
 * Servlet implementation class for Frugal.
 */
public class FServlet extends HttpServlet {

    private static final Logger LOGGER = LoggerFactory.getLogger(FServlet.class);

    private final FProcessor processor;

    private final FProtocolFactory inProtocolFactory;

    private final FProtocolFactory outProtocolFactory;

    private final Collection<Map.Entry<String, String>> customHeaders;

    /**
     * @see HttpServlet#HttpServlet()
     */
    public FServlet(FProcessor processor,
                    FProtocolFactory inProtocolFactory,
                    FProtocolFactory outProtocolFactory) {
        super();
        this.processor = processor;
        this.inProtocolFactory = inProtocolFactory;
        this.outProtocolFactory = outProtocolFactory;
        this.customHeaders = new ArrayList<>();
    }

    /**
     * @see HttpServlet#HttpServlet()
     */
    public FServlet(FProcessor processor, FProtocolFactory protocolFactory) {
        this(processor, protocolFactory, protocolFactory);
    }

    /**
     * Add a custom header to the returned response.
     *
     * @param key Header name
     * @param value Header value
     */
    public void addCustomHeader(final String key, final String value) {
        this.customHeaders.add(new Map.Entry<String, String>() {
            public String getKey() {
                return key;
            }

            public String getValue() {
                return value;
            }

            public String setValue(String value) {
                return null;
            }
        });
    }

    /**
     * Add a map of custom header to the returned response.
     *
     * @param headers Map of header name, header value pairs.
     */
    public void setCustomHeaders(Collection<Map.Entry<String, String>> headers) {
        this.customHeaders.clear();
        this.customHeaders.addAll(headers);
    }

    /**
     * @see HttpServlet#doPost(HttpServletRequest request, HttpServletResponse response)
     */
    @Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {

        // Read input data as bytes
        byte[] inputBytes;
        try {
            inputBytes = getInputBytes(request);
        } catch (IOException e) {
            LOGGER.error("Error reading input", e);
            response.setStatus(HttpServletResponse.SC_BAD_REQUEST);
            return;
        }

        // Process a frame of data
        TMemoryOutputBuffer outputBuffer;
        try {
            outputBuffer = processFrame(inputBytes);
        } catch (TException e) {
            LOGGER.error("Error processing frame", e);
            response.setStatus(HttpServletResponse.SC_BAD_REQUEST);
            return;
        }
        byte[] outputBytes = Base64.encodeBase64(outputBuffer.getWriteBytes());

        // Ensure response is within limit
        Integer responseLimit = getResponseLimit(request);
        if (responseLimit > 0 && outputBytes.length > responseLimit) {
            LOGGER.info("Response limit exceeded", responseLimit);
            response.setStatus(HttpServletResponse.SC_REQUEST_ENTITY_TOO_LARGE);
            return;
        }

        // Set response headers
        response.setContentType("application/x-frugal");
        response.setContentLength(outputBytes.length);
        response.setHeader("Content-Transfer-Encoding", "base64");

        // Add custom headers
        if (null != this.customHeaders) {
            for (Map.Entry<String, String> header : this.customHeaders) {
                response.addHeader(header.getKey(), header.getValue());
            }
        }

        // Write output body
        OutputStream outputStream = response.getOutputStream();
        outputStream.write(outputBytes);

        response.setStatus(HttpServletResponse.SC_OK);
    }

    /**
     * @see HttpServlet#doGet(HttpServletRequest request, HttpServletResponse response)
     */
    protected void doGet(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {
        doPost(request, response);
    }

    /**
     * Process one frame of data.
     *
     * @param inputBytes an input frame
     * @return The processes frame as an output buffer
     * @throws TException if error processing frame
     */
    protected TMemoryOutputBuffer processFrame(byte[] inputBytes) throws TException {
        if (inputBytes == null) {
            throw new TException("inputBytes must not be null.");
        }
        // Exclude first 4 bytes which represent frame size
        byte[] inputFrame = Arrays.copyOfRange(inputBytes, 4, inputBytes.length);

        TTransport inTransport = new TMemoryInputTransport(inputFrame);
        TMemoryOutputBuffer outTransport = new TMemoryOutputBuffer();

        processor.process(inProtocolFactory.getProtocol(inTransport), outProtocolFactory.getProtocol(outTransport));

        return outTransport;
    }

    /**
     * Returns the size limit of the response payload.
     * Set in the x-frugal-payload-limit HTTP header.
     *
     * @param request an HTTP request
     * @return The size limit of the response, 0 if no limit header set
     */
    protected Integer getResponseLimit(HttpServletRequest request) {
        String payloadHeader = request.getHeader("x-frugal-payload-limit");
        Integer responseLimit;
        try {
            responseLimit = Integer.parseInt(payloadHeader);
        } catch (NumberFormatException ignored) {
            responseLimit = 0;
        }
        return responseLimit;
    }

    /**
     * Returns payload body from the request as a byte[].
     *
     * @param request an HTTP request
     * @return The payload body
     * @throws IOException when invalid request frame
     */
    protected byte[] getInputBytes(HttpServletRequest request) throws IOException {
        StringBuilder buffer = new StringBuilder();
        BufferedReader reader = request.getReader();
        String line;
        while ((line = reader.readLine()) != null) {
            buffer.append(line);
        }
        String data = buffer.toString();
        byte[] inputBytes = Base64.decodeBase64(data);

        if (inputBytes.length <= 4) {
            throw new IOException("Invalid request frame");
        }
        return inputBytes;
    }

}
