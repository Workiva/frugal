package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.commons.codec.binary.Base64;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;

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
import java.util.Enumeration;
import java.util.Map;
import java.util.Scanner;

/**
 * Servlet implementation class for Frugal.
 */
public class FServlet extends HttpServlet {

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

        // Read input bytes
        StringBuilder buffer = new StringBuilder();
        BufferedReader reader = request.getReader();
        String line;
        while ((line = reader.readLine()) != null) {
            buffer.append(line);
        }
        String data = buffer.toString();
        byte[] inputBytes = Base64.decodeBase64(data);

        // Process frame (exclude first 4 bytes which represent frame size).
        byte[] inputFrame = Arrays.copyOfRange(inputBytes, 4, inputBytes.length);
        TTransport inTransport = new TMemoryInputTransport(inputFrame);
        TMemoryOutputBuffer outTransport = new TMemoryOutputBuffer();

        try {
            processor.process(inProtocolFactory.getProtocol(inTransport), outProtocolFactory.getProtocol(outTransport));
        } catch (TException te) {
            throw new ServletException(te);
        }

        // Frame output
        byte[] outputBytes = outTransport.getWriteBytes();
//        byte[] frameBytes = ByteBuffer.allocate(4).putInt(outputBytes.length).array();
//        byte[] output = new byte[frameBytes.length + outputBytes.length];
//        System.arraycopy(frameBytes, 0, output, 0, frameBytes.length);
//        System.arraycopy(outputBytes, 0, output, frameBytes.length, outputBytes.length);
//
//        System.out.println(Arrays.toString(outputBytes));
//        System.out.println(Arrays.toString(frameBytes));
//        System.out.println(Arrays.toString(output));

        // Base64 encode and return
        byte[] framedOutput = Base64.encodeBase64(outputBytes);
        OutputStream out = response.getOutputStream();
        out.write(framedOutput);

        // Set response headers
        response.setContentType("application/x-frugal");
        response.setContentLength(framedOutput.length);
        response.setHeader("Content-Transfer-Encoding", "base64");
    }

    /**
     * @see HttpServlet#doGet(HttpServletRequest request, HttpServletResponse response)
     */
    protected void doGet(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {
        doPost(request, response);
    }

    private void printRequest(HttpServletRequest httpRequest) {
        System.out.println(" \n\n Headers");

        Enumeration headerNames = httpRequest.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = (String) headerNames.nextElement();
            System.out.println(headerName + " = " + httpRequest.getHeader(headerName));
        }

        System.out.println("\n\nParameters");

        Enumeration params = httpRequest.getParameterNames();
        while (params.hasMoreElements()) {
            String paramName = (String) params.nextElement();
            System.out.println(paramName + " = " + httpRequest.getParameter(paramName));
        }

        System.out.println("\n\n Row data");
        System.out.println(extractPostRequestBody(httpRequest));
    }

    static String extractPostRequestBody(HttpServletRequest request) {
        if ("POST".equalsIgnoreCase(request.getMethod())) {
            Scanner s = null;
            try {
                s = new Scanner(request.getInputStream(), "UTF-8").useDelimiter("\\A");
            } catch (IOException e) {
                e.printStackTrace();
            }
            return s.hasNext() ? s.next() : "";
        }
        return "";
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
