package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.protocol.Headers;
import com.workiva.frugal.util.ProtocolUtils;
import org.apache.commons.codec.binary.Base64;
import org.apache.http.Header;
import org.apache.http.HttpEntity;
import org.apache.http.HttpEntityEnclosingRequest;
import org.apache.http.HttpException;
import org.apache.http.HttpRequest;
import org.apache.http.HttpResponse;
import org.apache.http.HttpStatus;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.protocol.HttpContext;
import org.apache.http.protocol.HttpRequestHandler;
import org.apache.http.util.EntityUtils;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryBuffer;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.Arrays;


/**
 * Provides a request handler for Apache HTTP components. The handler
 * is instantiated with an FProcessor for responding to requests.
 */
public class FHttpRequestHandler implements HttpRequestHandler {
    private static final Logger LOGGER = LoggerFactory.getLogger(FHttpRequestHandler.class);

    private final int requestSizeLimit;
    private final FProcessor processor;
    private final FProtocolFactory inputProtoFactory;
    private final FProtocolFactory outputProtoFactory;

    private FHttpRequestHandler(FProcessor processor,
                                FProtocolFactory inputProtoFactory,
                                FProtocolFactory outputProtoFactory,
                                int requestSizeLimit) {
        this.processor = processor;
        this.inputProtoFactory = inputProtoFactory;
        this.outputProtoFactory = outputProtoFactory;
        this.requestSizeLimit = requestSizeLimit;
    }

    /**
     * Create a new instance of an HttpRequestHandler with no limit on request size.
     *
     * @param processor the processor for incoming messages
     * @param inputProtoFactory input serialization protocol
     * @param outputProtoFactory output serialization protocol
     * @return FHttpRequestHandler
     */
    public static FHttpRequestHandler of(FProcessor processor,
                                         FProtocolFactory inputProtoFactory,
                                         FProtocolFactory outputProtoFactory) {
        return new FHttpRequestHandler(processor, inputProtoFactory, outputProtoFactory, 0);
    }

    /**
     * Set the request size limit for this handler.
     * <p>
     * A size limit <= 0 implies no size limit
     *
     * @param requestSizeLimit the size limit of incoming requests (bytes)
     * @return a new FHttpRequestHandler instance.
     */
    public FHttpRequestHandler withRequestSizeLimit(int requestSizeLimit) {
        return new FHttpRequestHandler(processor, inputProtoFactory, outputProtoFactory, requestSizeLimit);
    }

    /**
     * Handles the request and produces a response to be sent back to
     * the client.
     *
     * @param request  the HTTP request.
     * @param response the HTTP response.
     * @param context  the HTTP execution context.
     * @throws IOException   in case of an I/O error.
     * @throws HttpException in case of HTTP protocol violation or a processing
     *                       problem.
     */
    @Override
    public void handle(HttpRequest request,
                       HttpResponse response,
                       HttpContext context) throws HttpException, IOException {
        if (request instanceof HttpEntityEnclosingRequest) {
            // Read in bytes
            HttpEntity entity = ((HttpEntityEnclosingRequest) request).getEntity();
            byte[] inBytes = Base64.decodeBase64(EntityUtils.toByteArray(entity));
            EntityUtils.consume(entity);

            // Return error if size is greater than limit
            if ((requestSizeLimit > 0) && (inBytes.length > requestSizeLimit)) {
                // Exit with correct status
                response.setStatusCode(HttpStatus.SC_REQUEST_TOO_LONG);
                response.setReasonPhrase("PAYLOAD TOO LARGE");
                return;
            }

            // Read in frame (exclude first 4 bytes which represent frame size).
            TTransport input = new TMemoryInputTransport(Arrays.copyOfRange(inBytes, 4, inBytes.length));
            TMemoryBuffer output = new TMemoryBuffer(inBytes.length);
            try {
                processor.process(inputProtoFactory.getProtocol(input), outputProtoFactory.getProtocol(output));
            } catch (TException e) {
                // Exit with correct status
                response.setStatusCode(HttpStatus.SC_BAD_REQUEST);
                response.setReasonPhrase("BAD REQUEST");
                return;
            }

            // Respond with error if response too large
            Header payloadLimit = request.getFirstHeader(Headers.X_FRUGAL_PAYLOAD_LIMIT_HEADER);
            if (payloadLimit != null) {
                Integer limit = Integer.parseInt(payloadLimit.getValue());
                if (output.getArray().length > limit) {
                    // Exit with correct status
                    response.setStatusCode(HttpStatus.SC_FORBIDDEN);
                    response.setReasonPhrase("FORBIDDEN");
                    return;
                }
            }

            // Add frame size (4-byte int32).
            byte[] outBytes = new byte[output.length() + 4];
            ProtocolUtils.writeInt(output.length(), outBytes, 0);
            System.arraycopy(output.getArray(), 0, outBytes, 4, output.length());

            // Populate HTTP response.
            response.setStatusCode(HttpStatus.SC_OK);
            response.setReasonPhrase("OK");
            response.setHeader(Headers.ACCEPT_HEADER, Headers.APPLICATION_X_FRUGAL_HEADER);
            response.setHeader(Headers.CONTENT_TRANSFER_ENCODING_HEADER, Headers.CONTENT_TRANSFER_ENCODING);
            response.setEntity(new StringEntity(Base64.encodeBase64String(outBytes),
                                                ContentType.create(Headers.APPLICATION_X_FRUGAL_HEADER,
                                                                   Headers.CONTENT_TYPE)));
        }
    }

}
