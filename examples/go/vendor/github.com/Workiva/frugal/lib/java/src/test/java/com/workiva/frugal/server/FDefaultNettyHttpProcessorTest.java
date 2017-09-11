package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpHeaders;
import org.apache.commons.codec.binary.Base64;
import org.apache.thrift.TException;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;

import java.io.IOException;
import java.nio.ByteBuffer;

import static io.netty.handler.codec.http.HttpResponseStatus.BAD_REQUEST;
import static io.netty.handler.codec.http.HttpResponseStatus.INTERNAL_SERVER_ERROR;
import static io.netty.handler.codec.http.HttpResponseStatus.OK;
import static io.netty.handler.codec.http.HttpResponseStatus.REQUEST_ENTITY_TOO_LARGE;
import static io.netty.handler.codec.http.HttpVersion.HTTP_1_1;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.equalTo;
import static org.hamcrest.Matchers.notNullValue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.spy;

/**
 * Tests for {@link FDefaultNettyHttpProcessor}.
 */
public class FDefaultNettyHttpProcessorTest {

    @Rule
    public ExpectedException thrown = ExpectedException.none();

    private static FullHttpRequest mockRequest;
    private static FullHttpResponse mockResponse;

    private static HttpHeaders mockRequestHeaders;
    private static HttpHeaders mockResponseHeaders;

    private static FDefaultNettyHttpProcessor httpProcessor;

    @Before
    public final void setUp() throws IOException {
        mockRequest = mock(FullHttpRequest.class);
        mockRequestHeaders = mock(HttpHeaders.class);
        doReturn(mockRequestHeaders).when(mockRequest).headers();
        doReturn(HTTP_1_1).when(mockRequest).protocolVersion();
        mockResponse = mock(FullHttpResponse.class);
        mockResponseHeaders = mock(HttpHeaders.class);

        FProcessor mockProcessor = mock(FProcessor.class);
        FProtocolFactory mockProtocolFactory = mock(FProtocolFactory.class);
        httpProcessor = FDefaultNettyHttpProcessor.of(mockProcessor, mockProtocolFactory);
    }

    @Test
    public final void testValidResponseLimit() {
        doReturn("2096").when(mockRequestHeaders).get("x-frugal-payload-limit");

        Integer limit = httpProcessor.getResponseLimit(mockRequestHeaders);
        assertThat(limit, equalTo(2096));
    }

    @Test
    public final void testNullResponseLimit() {
        doReturn(null).when(mockRequestHeaders).get("x-frugal-payload-limit");

        Integer limit = httpProcessor.getResponseLimit(mockRequestHeaders);
        assertThat(limit, equalTo(0));
    }

    @Test
    public final void testStringResponseLimit() {
        doReturn("not-a-number").when(mockRequestHeaders).get("x-frugal-payload-limit");

        Integer limit = httpProcessor.getResponseLimit(mockRequestHeaders);
        assertThat(limit, equalTo(0));
    }

    @Test
    public final void testProcessValidFrame() throws TException, IOException {
        byte[] requestBody = "request_body".getBytes();
        byte[] bytes = ByteBuffer.allocate(4 + requestBody.length)
                .putInt(requestBody.length)
                .put(requestBody)
                .array();
        ByteBuf inputBytes = Unpooled.copiedBuffer(Base64.encodeBase64(bytes));

        ByteBuf outputBytes = httpProcessor.processFrame(inputBytes);

        assertThat(outputBytes, notNullValue());
    }

    @Test
    public final void testProcessFrameThrowsOnInvalidFrame() throws IOException, TException {
        ByteBuf inputBytes = Unpooled.copiedBuffer(Base64.encodeBase64("r".getBytes()));

        thrown.expect(IOException.class);
        thrown.expectMessage("Invalid request size 1");
        httpProcessor.processFrame(inputBytes);
    }

    @Test
    public final void testProcessFrameThrowsOnSizeMismatch() throws TException, IOException {
        ByteBuf inputBytes = Unpooled.copiedBuffer(Base64.encodeBase64("request_body".getBytes()));
        doReturn(inputBytes).when(mockRequest).content();

        thrown.expect(IOException.class);
        thrown.expectMessage("Mismatch between expected frame size (1919250805) and actual size (8)");
        httpProcessor.processFrame(inputBytes);
    }

    @Test
    public final void test400OnInvalidFrame() throws IOException {
        ByteBuf inputBytes = Unpooled.copiedBuffer(Base64.encodeBase64("r".getBytes()));
        doReturn(inputBytes).when(mockRequest).content();

        FullHttpResponse response = httpProcessor.process(mockRequest);
        assertThat(response.status(), equalTo(BAD_REQUEST));
    }

    @Test
    public final void test413OverResponseLimit() throws IOException, TException {
        byte[] requestBody = "request_body".getBytes();
        byte[] bytes = ByteBuffer.allocate(4 + requestBody.length)
                .putInt(requestBody.length)
                .put(requestBody)
                .array();
        ByteBuf inputBytes = Unpooled.copiedBuffer(Base64.encodeBase64(bytes));
        doReturn(inputBytes).when(mockRequest).content();
        doReturn("1").when(mockRequestHeaders).get("x-frugal-payload-limit");

        FDefaultNettyHttpProcessor spyProcessor = spy(httpProcessor);
        ByteBuf outputBytes = Unpooled.copiedBuffer(Base64.encodeBase64("response_body".getBytes()));
        doReturn(outputBytes).when(spyProcessor).processFrame(any(ByteBuf.class));

        FullHttpResponse response = spyProcessor.process(mockRequest);

        assertThat(response.status(), equalTo(REQUEST_ENTITY_TOO_LARGE));
    }

    @Test
    public final void test500OnProcessorException() throws IOException, TException {
        byte[] requestBody = "request_body".getBytes();
        byte[] bytes = ByteBuffer.allocate(4 + requestBody.length)
                .putInt(requestBody.length)
                .put(requestBody)
                .array();
        ByteBuf inputBytes = Unpooled.copiedBuffer(Base64.encodeBase64(bytes));
        doReturn(inputBytes).when(mockRequest).content();

        FDefaultNettyHttpProcessor spyProcessor = spy(httpProcessor);
        doThrow(new TException()).when(spyProcessor).processFrame(any(ByteBuf.class));

        FullHttpResponse response = spyProcessor.process(mockRequest);
        assertThat(response.status(), equalTo(INTERNAL_SERVER_ERROR));
    }

    @Test
    public final void test200Ok() throws IOException, TException {
        byte[] requestBody = "request_body".getBytes();
        byte[] bytes = ByteBuffer.allocate(4 + requestBody.length)
                .putInt(requestBody.length)
                .put(requestBody)
                .array();
        ByteBuf inputBytes = Unpooled.copiedBuffer(Base64.encodeBase64(bytes));
        doReturn(inputBytes).when(mockRequest).content();

        FDefaultNettyHttpProcessor spyProcessor = spy(httpProcessor);
        ByteBuf outputBytes = Unpooled.copiedBuffer(Base64.encodeBase64("response_body".getBytes()));
        doReturn(outputBytes).when(spyProcessor).processFrame(any(ByteBuf.class));

        FullHttpResponse response = spyProcessor.process(mockRequest);
        assertThat(response.status(), equalTo(OK));
    }
}
