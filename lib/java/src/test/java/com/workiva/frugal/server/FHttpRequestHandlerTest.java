package com.workiva.frugal.server;

import com.workiva.frugal.processor.FBaseProcessor;
import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.protocol.HttpHeaders;
import org.apache.http.HttpEntityEnclosingRequest;
import org.apache.http.HttpException;
import org.apache.http.HttpResponse;
import org.apache.http.entity.ByteArrayEntity;
import org.apache.http.message.BasicHeader;
import org.apache.http.protocol.HttpContext;
import org.junit.Rule;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;
import org.mockito.Mock;
import org.mockito.junit.MockitoJUnit;
import org.mockito.junit.MockitoRule;

import java.io.IOException;

import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

/**
 * Tests for {@link FHttpRequestHandler}.
 */
@RunWith(JUnit4.class)
public class FHttpRequestHandlerTest {

    @Rule
    public MockitoRule rule = MockitoJUnit.rule();

    @Mock
    HttpEntityEnclosingRequest mockRequest;

    @Mock
    HttpResponse mockResponse;

    @Mock
    HttpContext mockContext;

    @Test
    public void testRequestTooLongIfExceedsSizeLimit() throws IOException, HttpException {
        // given
        FProcessor processor = mock(FBaseProcessor.class);
        FProtocolFactory protocolFactory = mock(FProtocolFactory.class);
        FHttpRequestHandler requestHandler = FHttpRequestHandler
                .of(processor, protocolFactory, protocolFactory)
                .withRequestSizeLimit(1);

        doReturn(new ByteArrayEntity("Hello World".getBytes())).when(mockRequest).getEntity();

        // when
        requestHandler.handle(mockRequest, mockResponse, mockContext);

        // then
        verify(mockResponse).setStatusCode(413);
        verify(mockResponse).setReasonPhrase("PAYLOAD TOO LARGE");
    }

    @Test
    public void testForbiddenIfResponseExceedsSizeLimit() throws IOException, HttpException {
        // given
        FProcessor processor = mock(FBaseProcessor.class);
        FProtocolFactory protocolFactory = mock(FProtocolFactory.class);
        FHttpRequestHandler requestHandler = FHttpRequestHandler
                .of(processor, protocolFactory, protocolFactory);

        doReturn(new ByteArrayEntity("Hello World".getBytes()))
                .when(mockRequest).getEntity();
        doReturn(new BasicHeader(HttpHeaders.X_FRUGAL_PAYLOAD_LIMIT_HEADER, "1"))
                .when(mockRequest).getFirstHeader(HttpHeaders.X_FRUGAL_PAYLOAD_LIMIT_HEADER);

        // when
        requestHandler.handle(mockRequest, mockResponse, mockContext);

        // then
        verify(mockResponse).setStatusCode(403);
        verify(mockResponse).setReasonPhrase("FORBIDDEN");
    }

    @Test
    public void testSuccessResponse() throws IOException, HttpException {
        // given
        FProcessor processor = mock(FBaseProcessor.class);
        FProtocolFactory protocolFactory = mock(FProtocolFactory.class);
        FHttpRequestHandler requestHandler = FHttpRequestHandler
                .of(processor, protocolFactory, protocolFactory);

        doReturn(new ByteArrayEntity("Hello World".getBytes()))
                .when(mockRequest).getEntity();

        // when
        requestHandler.handle(mockRequest, mockResponse, mockContext);

        // then
        verify(mockResponse).setStatusCode(200);
        verify(mockResponse).setReasonPhrase("OK");
        verify(mockResponse).setHeader(HttpHeaders.ACCEPT_HEADER, HttpHeaders.APPLICATION_X_FRUGAL_HEADER);
        verify(mockResponse).setHeader(HttpHeaders.CONTENT_TRANSFER_ENCODING_HEADER,
                                       HttpHeaders.CONTENT_TRANSFER_ENCODING);
    }
}

