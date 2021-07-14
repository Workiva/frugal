package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.thrift.TConfiguration;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServlet;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

import java.io.DataInputStream;
import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;

/**
 * Processes POST requests as Frugal requests for a processor.
 * <p>
 * By default, the HTTP request is limited to a 64MB Frugal payload size to
 * prevent client requests from causing the server to allocate too much memory.
 * <p>
 * The HTTP request may include an X-Frugal-Payload-Limit header setting the size
 * limit of responses from the server.
 * <p>
 * The HTTP processor returns a 500 response for any runtime errors when executing
 * a frame, a 400 response for an invalid frame, and a 413 response if the response
 * exceeds the payload limit specified by the client.
 * <p>
 * Both the request and response are base64 encoded.
 */
@SuppressWarnings("serial")
public class FJakartaServlet extends HttpServlet {
    private static final Logger LOGGER = LoggerFactory.getLogger(FJakartaServlet.class);

    private static final int DEFAULT_MAX_REQUEST_SIZE = 64 * 1024 * 1024;

    private final FProcessor processor;
    private final FProtocolFactory inProtocolFactory;
    private final FProtocolFactory outProtocolFactory;
    private final int maxRequestSize;
    private final ExecutorService exec;
    private final FServerEventHandler eventHandler;
    private final TConfiguration tconfig;

    /**
     * Creates a servlet for the specified processor and protocol factory, which
     * is used for both input and output.
     */
    public FJakartaServlet(FProcessor processor, FProtocolFactory protocolFactory) {
        this(processor, protocolFactory, DEFAULT_MAX_REQUEST_SIZE);
    }

    /**
     * Creates a servlet for the specified processor and protocol factory, which
     * is used for both input and output.
     *
     * @param maxRequestSize the maximum Frugal request size in bytes
     */
    public FJakartaServlet(FProcessor processor, FProtocolFactory protocolFactory, int maxRequestSize) {
        this(processor, protocolFactory, protocolFactory, maxRequestSize);
    }

    /**
     * Creates a servlet for the specified processor and input/output protocol
     * factories.
     */
    public FJakartaServlet(FProcessor processor, FProtocolFactory inProtocolFactory, FProtocolFactory outProtocolFactory) {
        this(processor, inProtocolFactory, outProtocolFactory, DEFAULT_MAX_REQUEST_SIZE);
    }

    /**
     * Creates a servlet for the specified processor and input/output protocol
     * factories.
     *
     * @param maxRequestSize the maximum Frugal request size in bytes
     */
    public FJakartaServlet(
            FProcessor processor,
            FProtocolFactory inProtocolFactory,
            FProtocolFactory outProtocolFactory,
            int maxRequestSize) {
        this(builder()
                .processor(processor)
                .inProtocolFactory(inProtocolFactory)
                .outProtocolFactory(outProtocolFactory)
                .maxRequestSize(maxRequestSize));
    }

    private FJakartaServlet(Builder b) {
        this.processor = b.processor;
        this.inProtocolFactory = b.inProtocolFactory;
        this.outProtocolFactory = b.outProtocolFactory;
        this.maxRequestSize = b.maxRequestSize;
        this.exec = b.exec;
        this.eventHandler = b.eventHandler != null ? b.eventHandler : new FDefaultServerEventHandler(5000);
        this.tconfig = new TConfiguration(
                Integer.MAX_VALUE,
                TConfiguration.DEFAULT_MAX_FRAME_SIZE,
                TConfiguration.DEFAULT_RECURSION_DEPTH
        );
    }

    @Override
    public void doPost(HttpServletRequest req, HttpServletResponse resp) throws ServletException, IOException {
        Map<Object, Object> ephemeralProperties = new HashMap<>();
        eventHandler.onRequestReceived(ephemeralProperties);
        try {
            process(req, resp, ephemeralProperties);
        } finally {
            eventHandler.onRequestEnded(ephemeralProperties);
        }
    }

    private void process(HttpServletRequest req, HttpServletResponse resp, Map<Object, Object> ephemeralProperties) throws ServletException, IOException {
        byte[] frame;
        try (InputStream decoderIn = Base64.getDecoder().wrap(req.getInputStream());
                DataInputStream dataIn = new DataInputStream(decoderIn)) {
            try {
                long size = dataIn.readInt() & 0xffff_ffffL;
                if (size > maxRequestSize) {
                    LOGGER.debug("Request size too large. Received: {}, Limit: {}", size, maxRequestSize);
                    resp.setStatus(HttpServletResponse.SC_REQUEST_ENTITY_TOO_LARGE);
                    return;
                }

                frame = new byte[(int) size];
                dataIn.readFully(frame);
            } catch (EOFException e) {
                LOGGER.debug("Request body too short");
                resp.setStatus(HttpServletResponse.SC_BAD_REQUEST);
                return;
            }

            if (dataIn.read() != -1) {
                LOGGER.debug("Request body too long");
                resp.setStatus(HttpServletResponse.SC_BAD_REQUEST);
                return;
            }
        }

        byte[] data;
        try {
            if (exec == null) {
                data = process(frame, ephemeralProperties);
            } else {
                try {
                    data = exec.submit(() -> process(frame, ephemeralProperties)).get();
                } catch (ExecutionException e) {
                    Throwable cause = e.getCause();
                    if (cause instanceof Error) {
                        throw (Error) cause;
                    }
                    if (cause instanceof RuntimeException) {
                        throw (RuntimeException) cause;
                    }
                    if (cause instanceof TException) {
                        throw (TException) cause;
                    }
                    throw new RuntimeException(e);
                }
            }
        } catch (InterruptedException | TException e) {
            LOGGER.error("Frugal processor returned unhandled error", e);
            resp.setStatus(HttpServletResponse.SC_INTERNAL_SERVER_ERROR);
            return;
        }

        int responseLimit = getResponseLimit(req);
        if (responseLimit > 0 && data.length > responseLimit) {
            LOGGER.debug("Response size too large for client. Received: {}, Limit: {}",
                    data.length, responseLimit);
            resp.setStatus(HttpServletResponse.SC_REQUEST_ENTITY_TOO_LARGE);
            return;
        }

        resp.setContentType("application/x-frugal");
        resp.setHeader("Content-Transfer-Encoding", "base64");
        try (OutputStream out = Base64.getEncoder().wrap(resp.getOutputStream())) {
            out.write(data);
        }
    }

    private byte[] process(byte[] frame, Map<Object, Object> ephemeralProperties) throws TException {
        eventHandler.onRequestStarted(ephemeralProperties);

        TTransport inTransport = new TMemoryInputTransport(tconfig, frame);
        TMemoryOutputBuffer outTransport = new TMemoryOutputBuffer();
        try {
            FProtocol inProtocol = inProtocolFactory.getProtocol(inTransport);
            FProtocol outProtocol = outProtocolFactory.getProtocol(outTransport);
            processor.process(inProtocol, outProtocol);
        } catch (RuntimeException e) {
            // Already logged by FBaseProcessor and written to the output buffer
            // as an application exception, so write that response back to the
            // client just like FNatsServer.
        }

        return outTransport.getWriteBytes();
    }

    // Visible for testing.
    static int getResponseLimit(HttpServletRequest req) {
        String payloadHeader = req.getHeader("x-frugal-payload-limit");
        int responseLimit;
        try {
            responseLimit = Integer.parseInt(payloadHeader);
        } catch (NumberFormatException ignored) {
            responseLimit = 0;
        }
        return responseLimit;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private FProcessor processor;
        private FProtocolFactory inProtocolFactory;
        private FProtocolFactory outProtocolFactory;
        private int maxRequestSize = DEFAULT_MAX_REQUEST_SIZE;
        private ExecutorService exec;
        private FServerEventHandler eventHandler;

        public FJakartaServlet build() {
            return new FJakartaServlet(this);
        }

        public Builder processor(FProcessor processor) {
            this.processor = processor;
            return this;
        }

        public Builder protocolFactory(FProtocolFactory protocolFactory) {
            return inProtocolFactory(protocolFactory)
                    .outProtocolFactory(protocolFactory);
        }

        public Builder inProtocolFactory(FProtocolFactory inProtocolFactory) {
            this.inProtocolFactory = inProtocolFactory;
            return this;
        }

        public Builder outProtocolFactory(FProtocolFactory outProtocolFactory) {
            this.outProtocolFactory = outProtocolFactory;
            return this;
        }

        public Builder maxRequestSize(int maxRequestSize) {
            this.maxRequestSize = maxRequestSize;
            return this;
        }

        public Builder executorService(ExecutorService exec) {
            this.exec = exec;
            return this;
        }

        public Builder eventHandler(FServerEventHandler eventHandler) {
            this.eventHandler = eventHandler;
            return this;
        }
    }
}
