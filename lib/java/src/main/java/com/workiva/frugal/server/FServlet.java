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
     * @see HttpServlet#doPost(HttpServletRequest request, HttpServletResponse response)
     */
    @Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {
        Integer responseLimit;
        try {
            responseLimit = Integer.parseInt(request.getHeader("x-frugal-payload-limit"));
        } catch (NumberFormatException ignored) {
            responseLimit = 0;
        }

        // Read input bytes
        StringBuilder buffer = new StringBuilder();
        BufferedReader reader = request.getReader();
        String line;
        while ((line = reader.readLine()) != null) {
            buffer.append(line);
        }
        String data = buffer.toString();
        byte[] inputBytes = Base64.decodeBase64(data);

        if (inputBytes.length <= 4) {
            LOGGER.info("Invalid request frame length", inputBytes.length);
            response.setStatus(HttpServletResponse.SC_BAD_REQUEST);
            return;
        }

        // Process frame (exclude first 4 bytes which represent frame size).
        byte[] inputFrame = Arrays.copyOfRange(inputBytes, 4, inputBytes.length);
        TTransport inTransport = new TMemoryInputTransport(inputFrame);
        TMemoryOutputBuffer outTransport = new TMemoryOutputBuffer();

        try {
            processor.process(inProtocolFactory.getProtocol(inTransport), outProtocolFactory.getProtocol(outTransport));
        } catch (TException te) {
            throw new ServletException(te);
        }

        byte[] framedOutput = Base64.encodeBase64(outTransport.getWriteBytes());

        // Make sure response is within limit
        if (responseLimit > 0 && framedOutput.length > responseLimit) {
            LOGGER.info("Response limit exceeded", responseLimit);
            response.setStatus(HttpServletResponse.SC_REQUEST_ENTITY_TOO_LARGE);
            return;
        }

        // Base64 encode and return
        OutputStream out = response.getOutputStream();
        out.write(framedOutput);

        // Set response headers
        response.setContentType("application/x-frugal");
        response.setContentLength(framedOutput.length);
        response.setHeader("Content-Transfer-Encoding", "base64");

        // Add custom headers
        if (null != this.customHeaders) {
            for (Map.Entry<String, String> header : this.customHeaders) {
                response.addHeader(header.getKey(), header.getValue());
            }
        }
    }

    /**
     * @see HttpServlet#doGet(HttpServletRequest request, HttpServletResponse response)
     */
    protected void doGet(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {
        doPost(request, response);
    }

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

    public void setCustomHeaders(Collection<Map.Entry<String, String>> headers) {
        this.customHeaders.clear();
        this.customHeaders.addAll(headers);
    }
}
