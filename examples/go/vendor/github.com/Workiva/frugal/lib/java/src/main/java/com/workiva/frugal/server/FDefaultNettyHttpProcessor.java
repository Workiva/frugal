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

package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.handler.codec.http.DefaultFullHttpResponse;
import io.netty.handler.codec.http.DefaultHttpHeaders;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpHeaders;
import io.netty.handler.codec.http.HttpUtil;
import org.apache.commons.codec.binary.Base64;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.BufferUnderflowException;
import java.nio.ByteBuffer;
import java.time.ZonedDateTime;
import java.time.format.DateTimeFormatter;
import java.util.AbstractMap;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Map;

import static io.netty.handler.codec.http.HttpHeaderNames.CONTENT_LENGTH;
import static io.netty.handler.codec.http.HttpHeaderNames.CONTENT_TRANSFER_ENCODING;
import static io.netty.handler.codec.http.HttpHeaderNames.CONTENT_TYPE;
import static io.netty.handler.codec.http.HttpHeaderNames.DATE;
import static io.netty.handler.codec.http.HttpResponseStatus.BAD_REQUEST;
import static io.netty.handler.codec.http.HttpResponseStatus.CONTINUE;
import static io.netty.handler.codec.http.HttpResponseStatus.INTERNAL_SERVER_ERROR;
import static io.netty.handler.codec.http.HttpResponseStatus.OK;
import static io.netty.handler.codec.http.HttpResponseStatus.REQUEST_ENTITY_TOO_LARGE;
import static io.netty.handler.codec.http.HttpVersion.HTTP_1_1;

/**
 * Default processor implementation for {@link FNettyHttpProcessor}.
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
public class FDefaultNettyHttpProcessor implements FNettyHttpProcessor {

    private static final Logger LOGGER = LoggerFactory.getLogger(FDefaultNettyHttpProcessor.class);

    private final FProcessor processor;
    private final FProtocolFactory inProtocolFactory;
    private final FProtocolFactory outProtocolFactory;
    private final Collection<Map.Entry<String, String>> customHeaders;

    private FDefaultNettyHttpProcessor(
            FProcessor processor,
            FProtocolFactory inProtocolFactory,
            FProtocolFactory outProtocolFactory) {
        this.processor = processor;
        this.inProtocolFactory = inProtocolFactory;
        this.outProtocolFactory = outProtocolFactory;
        this.customHeaders = new ArrayList<>();
    }

    /**
     * Create a new HTTP processor, setting the input and output protocol.
     *
     * @param processor       Frugal request processor
     * @param protocolFactory input and output protocol
     * @return a new processor
     */
    public static FDefaultNettyHttpProcessor of(FProcessor processor, FProtocolFactory protocolFactory) {
        return new FDefaultNettyHttpProcessor(processor, protocolFactory, protocolFactory);
    }

    /**
     * Create a new HTTP processor, setting the input and output protocol.
     *
     * @param processor          Frugal request processor
     * @param inProtocolFactory  input protocol
     * @param outProtocolFactory output protocol
     * @return a new processor
     */
    public static FDefaultNettyHttpProcessor of(
            FProcessor processor,
            FProtocolFactory inProtocolFactory,
            FProtocolFactory outProtocolFactory) {
        return new FDefaultNettyHttpProcessor(processor, inProtocolFactory, outProtocolFactory);
    }

    /**
     * Returns the size limit of the response payload.
     * Set in the x-frugal-payload-limit HTTP header.
     *
     * @param headers HTTP headers from incoming request
     * @return The size limit of the response, 0 if no limit header set
     */
    protected Integer getResponseLimit(HttpHeaders headers) {
        String payloadHeader = headers.get("x-frugal-payload-limit");
        Integer responseLimit;
        try {
            responseLimit = Integer.parseInt(payloadHeader);
        } catch (NumberFormatException ignored) {
            responseLimit = 0;
        }
        return responseLimit;
    }

    /**
     * Process one frame of data.
     *
     * @param inputBuffer an input frame
     * @return The processes frame as an output buffer
     * @throws TException  if an application error occurred when processing a validly formed frame
     * @throws IOException if the frame is invalid, not conforming to the Frugal protocol
     */
    public ByteBuf processFrame(ByteBuf inputBuffer) throws TException, IOException {
        // Read base64 encoded input
        byte[] encodedBytes = new byte[inputBuffer.readableBytes()];
        inputBuffer.readBytes(encodedBytes);
        byte[] inputBytes = Base64.decodeBase64(encodedBytes);

        ByteBuffer buff = ByteBuffer.wrap(inputBytes);

        int sz;
        try {
            sz = buff.getInt();
        } catch (BufferUnderflowException e) {
            // Need 4 bytes for the frame size, at a minimum.
            throw new IOException("Invalid request size " + inputBytes.length);
        }

        // Ensure expected frame size equals actual size.
        if (sz != buff.remaining()) {
            throw new IOException(
                    String.format("Mismatch between expected frame size (%d) and actual size (%d)",
                            sz, inputBytes.length - 4)
            );
        }

        // Process a frame, exclude frame length (first 4 bytes)
        // TODO: use TByteBuffer that wraps buff once Thrift 0.10.0 is released to avoid this copy.
        byte[] inputFrame = Arrays.copyOfRange(inputBytes, 4, inputBytes.length);
        TTransport inTransport = new TMemoryInputTransport(inputFrame);
        TMemoryOutputBuffer outTransport = new TMemoryOutputBuffer();
        processor.process(inProtocolFactory.getProtocol(inTransport), outProtocolFactory.getProtocol(outTransport));

        // Write base64 encoded output
        byte[] outputBytes = Base64.encodeBase64(outTransport.getWriteBytes());
        return Unpooled.copiedBuffer(outputBytes);
    }

    /**
     * Process an HTTP request and return an HTTP response.
     *
     * @param request an HTTP request
     * @return an HTTP response, processed by an FProcessor
     */
    @Override
    public FullHttpResponse process(FullHttpRequest request) {
        if (HttpUtil.is100ContinueExpected(request)) {
            return new DefaultFullHttpResponse(HTTP_1_1, CONTINUE);
        }

        ByteBuf body = request.content();
        ByteBuf outputBuffer = Unpooled.EMPTY_BUFFER;
        try {
            outputBuffer = processFrame(body);
        } catch (TException e) {
            LOGGER.error("Frugal processor returned unhandled error:", e);
            String errorMessage = "";
            if (e.getMessage() != null) {
                errorMessage = e.getMessage();
            }
            return new DefaultFullHttpResponse(
                    HTTP_1_1,
                    INTERNAL_SERVER_ERROR,
                    Unpooled.copiedBuffer(errorMessage.getBytes()));

        } catch (IOException e) {
            LOGGER.error("Frugal processor invalid frame:", e);
            String errorMessage = "";
            if (e.getMessage() != null) {
                errorMessage = e.getMessage();
            }
            return new DefaultFullHttpResponse(
                    HTTP_1_1,
                    BAD_REQUEST,
                    Unpooled.copiedBuffer(errorMessage.getBytes()));
        } finally {
            body.release();
        }

        Integer responseLimit = getResponseLimit(request.headers());
        if (responseLimit > 0 && outputBuffer.readableBytes() > responseLimit) {
            LOGGER.error("Response size too large for client." +
                    " Received: " + outputBuffer.readableBytes() + ", Limit: " + responseLimit);

            FullHttpResponse errorResponse = new DefaultFullHttpResponse(
                    HTTP_1_1,
                    REQUEST_ENTITY_TOO_LARGE);
            errorResponse.headers().set(CONTENT_LENGTH, "0");
            return errorResponse;
        }

        FullHttpResponse response = new DefaultFullHttpResponse(
                HTTP_1_1,
                OK,
                outputBuffer);

        DefaultHttpHeaders headers = (DefaultHttpHeaders) response.headers();
        // Add custom headers
        for (Map.Entry<String, String> header : this.customHeaders) {
            headers.set(header.getKey(), header.getValue());
        }
        // Add required headers
        ZonedDateTime dateTime = ZonedDateTime.now();
        DateTimeFormatter formatter = DateTimeFormatter.RFC_1123_DATE_TIME;
        headers.set(DATE, dateTime.format(formatter));
        headers.set(CONTENT_TYPE, "application/x-frugal");
        headers.set(CONTENT_TRANSFER_ENCODING, "base64");
        headers.set(CONTENT_LENGTH, Integer.toString(outputBuffer.readableBytes()));

        return response;
    }

    /**
     * Add a custom header to the returned response.
     * NOTE: Once an HTTP handler is created with this processor,
     * this should not be called.
     *
     * @param key   Header name
     * @param value Header value
     */
    public void addCustomHeader(final String key, final String value) {
        this.customHeaders.add(new AbstractMap.SimpleEntry<>(key, value));
    }

    /**
     * Add a map of custom headers to the returned response.
     * NOTE: Once an HTTP handler is created with this processor,
     * this should not be called.
     *
     * @param headers Map of header name, header value pairs.
     */
    public void setCustomHeaders(Collection<Map.Entry<String, String>> headers) {
        this.customHeaders.clear();
        this.customHeaders.addAll(headers);
    }
}
