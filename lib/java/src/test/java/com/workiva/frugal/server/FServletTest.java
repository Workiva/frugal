package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import org.apache.thrift.TException;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.mockito.ArgumentCaptor;
import org.mockito.InOrder;
import javax.servlet.ReadListener;
import javax.servlet.ServletInputStream;
import javax.servlet.ServletOutputStream;
import javax.servlet.WriteListener;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.ByteBuffer;
import java.util.AbstractMap;
import java.util.Arrays;
import java.util.Base64;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.emptyArray;
import static org.hamcrest.Matchers.equalTo;
import static org.hamcrest.Matchers.not;
import static org.hamcrest.Matchers.nullValue;
import static org.hamcrest.Matchers.sameInstance;
import static org.junit.Assert.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.doAnswer;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.inOrder;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.verifyNoMoreInteractions;
import static org.mockito.Mockito.when;
import static org.mockito.hamcrest.MockitoHamcrest.argThat;

/**
 * Tests for {@link FServlet}.
 */
@SuppressWarnings("deprecation")
public class FServletTest {

    private static class ProxyServletInputStream extends ServletInputStream {
        private final InputStream in;

        ProxyServletInputStream(InputStream in) {
            this.in = in;
        }

        @Override
        public int read() throws IOException {
            return in.read();
        }

        @Override
        public boolean isFinished() {
            throw new UnsupportedOperationException();
        }

        @Override
        public boolean isReady() {
            throw new UnsupportedOperationException();
        }

        @Override
        public void setReadListener(ReadListener readListener) {
            throw new UnsupportedOperationException();
        }
    }

    private static class ProxyServletOutputStream extends ServletOutputStream {
        private final OutputStream out;

        ProxyServletOutputStream(OutputStream out) {
            this.out = out;
        }

        @Override
        public void write(int b) throws IOException {
            out.write(b);
        }

        @Override
        public boolean isReady() {
            throw new UnsupportedOperationException();
        }

        @Override
        public void setWriteListener(WriteListener writeListener) {
            throw new UnsupportedOperationException();
        }
    }

    private final FProcessor mockProcessor = mock(FProcessor.class);
    private final FProtocolFactory mockProtocolFactory = mock(FProtocolFactory.class);
    private final FProtocol mockProtocol = mock(FProtocol.class);
    private final ExecutorService executorService = Executors.newSingleThreadExecutor();
    private final FServerEventHandler mockEventHandler = mock(FServerEventHandler.class);
    private FServlet servlet = new FServlet(mockProcessor, mockProtocolFactory);

    private final HttpServletRequest mockRequest = mock(HttpServletRequest.class);
    private final HttpServletResponse mockResponse = mock(HttpServletResponse.class);

    private final ByteArrayOutputStream out = new ByteArrayOutputStream();
    private final ServletOutputStream servletOut = new ProxyServletOutputStream(out);

    @Before
    public void before() throws Exception {
        doReturn("POST").when(mockRequest).getMethod();
        doReturn("HTTP/1.1").when(mockRequest).getProtocol();
        doReturn(Collections.enumeration(Collections.emptyList())).when(mockRequest).getHeaderNames();

        doReturn(mockProtocol).when(mockProtocolFactory).getProtocol(any());

        doReturn(servletOut).when(mockResponse).getOutputStream();
    }

    @After
    public void after() throws Exception {
        verifyNoMoreInteractions(mockResponse);
        executorService.shutdown();
        assertThat(executorService.awaitTermination(5, TimeUnit.SECONDS), equalTo(true));
    }

    private void setupExecutor() {
        servlet = FServlet.builder()
                .processor(mockProcessor)
                .protocolFactory(mockProtocolFactory)
                .executorService(executorService)
                .build();
    }

    private void setupEventHandler() {
        servlet = FServlet.builder()
                .processor(mockProcessor)
                .protocolFactory(mockProtocolFactory)
                .eventHandler(mockEventHandler)
                .build();
    }

    @Test
    public final void testValidResponseLimit() {
        doReturn("2096").when(mockRequest).getHeader("x-frugal-payload-limit");

        Integer limit = FServlet.getResponseLimit(mockRequest);
        assertThat(limit, equalTo(2096));
    }

    @Test
    public final void testNullResponseLimit() {
        doReturn(null).when(mockRequest).getHeader("x-frugal-payload-limit");

        Integer limit = FServlet.getResponseLimit(mockRequest);
        assertThat(limit, equalTo(0));
    }

    @Test
    public final void testStringResponseLimit() {
        doReturn("not-a-number").when(mockRequest).getHeader("x-frugal-payload-limit");

        Integer limit = FServlet.getResponseLimit(mockRequest);
        assertThat(limit, equalTo(0));
    }

    @Test
    public void testGet() throws Exception {
        doReturn("GET").when(mockRequest).getMethod();
        doReturn("HTTP/1.1").when(mockRequest).getProtocol();

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).sendError(eq(HttpServletResponse.SC_METHOD_NOT_ALLOWED), any());
    }

    @Test
    public void testInputEmpty() throws Exception {
        ByteArrayInputStream in = new ByteArrayInputStream(new byte[0]);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setStatus(eq(HttpServletResponse.SC_BAD_REQUEST));
    }

    @Test
    public void testInputTooShort() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(new byte[] { 0, 0, 0, 1 });
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setStatus(eq(HttpServletResponse.SC_BAD_REQUEST));
    }

    @Test
    public void testInputTooLong() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(new byte[5]);
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setStatus(eq(HttpServletResponse.SC_BAD_REQUEST));
    }

    @Test
    public void testDefaultRequestSize() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(new byte[] { (byte) 0xff, (byte) 0xff, (byte) 0xff, (byte) 0xff });
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setStatus(eq(HttpServletResponse.SC_REQUEST_ENTITY_TOO_LARGE));
    }

    @Test
    public void testRequestSize() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(new byte[] { 0, 0, 0, 2 });
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        servlet = new FServlet(mockProcessor, mockProtocolFactory, 1);
        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setStatus(eq(HttpServletResponse.SC_REQUEST_ENTITY_TOO_LARGE));
    }

    private static byte[] withLength(byte[] b) {
        return ByteBuffer.allocate(4 + b.length)
                .putInt(b.length)
                .put(b)
                .array();
    }

    @Test
    public void testProcessorRuntimeException() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(withLength(new byte[0]));
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        doThrow(new RuntimeException("test")).when(mockProcessor).process(any(), any());

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setContentType("application/x-frugal");
        verify(mockResponse).setHeader("Content-Transfer-Encoding", "base64");
        verify(mockResponse).getOutputStream();
    }

    @Test
    public void testProcessorRuntimeExceptionWithExecutor() throws Exception {
        setupExecutor();
        testProcessorRuntimeException();
    }

    @Test
    public void testProcessorUnhandledException() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(withLength(new byte[0]));
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        doThrow(new TException("test")).when(mockProcessor).process(any(), any());

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setStatus(eq(HttpServletResponse.SC_INTERNAL_SERVER_ERROR));
    }

    @Test
    public void testProcessorUnhandledExceptionWithExecutor() throws Exception {
        setupExecutor();
        testProcessorUnhandledException();
    }

    @Test
    public void testResponseTooLong() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(withLength(new byte[2]));
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        doReturn("1").when(mockRequest).getHeader("x-frugal-payload-limit");

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setStatus(eq(HttpServletResponse.SC_REQUEST_ENTITY_TOO_LARGE));
    }

    @Test
    public void testOk() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(withLength(new byte[0]));
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        Thread testThread = Thread.currentThread();
        doAnswer(inv -> {
            assertThat(Thread.currentThread(), sameInstance(testThread));
            return null;
        }).when(mockProcessor).process(any(), any());

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setContentType("application/x-frugal");
        verify(mockResponse).setHeader("Content-Transfer-Encoding", "base64");
        verify(mockResponse).getOutputStream();
    }

    @Test
    public void testOkWithExecutor() throws Exception {
        setupExecutor();

        byte[] bytes = Base64.getEncoder().encode(withLength(new byte[0]));
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        Thread testThread = Thread.currentThread();
        doAnswer(inv -> {
            assertThat(Thread.currentThread(), not(sameInstance(testThread)));
            return null;
        }).when(mockProcessor).process(any(), any());

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setContentType("application/x-frugal");
        verify(mockResponse).setHeader("Content-Transfer-Encoding", "base64");
        verify(mockResponse).getOutputStream();
    }

    @Test
    public void testOkWithEventHandler() throws Exception {
        setupEventHandler();
        when(mockRequest.getHeaderNames()).thenAnswer(inv -> Collections.enumeration(Arrays.asList("a", "b")));
        when(mockRequest.getHeaders(any())).thenAnswer(inv -> Collections.enumeration(Collections.emptyList()));
        when(mockRequest.getHeaders(eq("a"))).thenAnswer(inv -> Collections.enumeration(Arrays.asList("a1")));
        when(mockRequest.getHeaders(eq("b"))).thenAnswer(inv -> Collections.enumeration(Arrays.asList("b1", "b2")));

        testOk();

        InOrder inOrder = inOrder(mockEventHandler, mockProcessor);
        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<Object, Object>> propsCaptor = ArgumentCaptor.forClass(Map.class);
        inOrder.verify(mockEventHandler).onRequestReceived(propsCaptor.capture());
        Map<Object, Object> props = propsCaptor.getValue();
        inOrder.verify(mockEventHandler).onRequestStarted(argThat(sameInstance(props)));
        inOrder.verify(mockProcessor).process(any(), any());
        inOrder.verify(mockEventHandler).onRequestEnded(argThat(sameInstance(props)));

        @SuppressWarnings("unchecked")
        Map<String, List<String>> headers = (Map<String, List<String>>) props.get("http_request_headers");
        Map<String, List<String>> expectedHeaders = new HashMap<>();
        expectedHeaders.put("a", Arrays.asList("a1"));
        expectedHeaders.put("b", Arrays.asList("b1", "b2"));
        assertThat(headers, equalTo(expectedHeaders));
        assertThat(headers.get("doesnotexist"), nullValue());
        assertThat(headers.get("a"), equalTo(Arrays.asList("a1")));
        assertThat(headers.get("A"), equalTo(Arrays.asList("a1")));
        assertThat(headers.keySet(), equalTo(expectedHeaders.keySet()));
        assertThat(headers.keySet().contains("doesnotexist"), equalTo(false));
        assertThat(headers.keySet().contains("a"), equalTo(true));
        assertThat(headers.keySet().contains("A"), equalTo(true));
        assertThat(headers.entrySet(), equalTo(expectedHeaders.entrySet()));
        assertThat(headers.entrySet().contains(
                new AbstractMap.SimpleEntry<>("a", null)), equalTo(false));
        assertThat(headers.entrySet().contains(
                new AbstractMap.SimpleEntry<>("a", Arrays.asList("x"))), equalTo(false));
        assertThat(headers.entrySet().contains(
                new AbstractMap.SimpleEntry<>("doesnotexist", Arrays.asList("a1"))), equalTo(false));
        assertThat(headers.entrySet().contains(
                new AbstractMap.SimpleEntry<>("a", Arrays.asList("a1"))), equalTo(true));
        assertThat(headers.entrySet().contains(
                new AbstractMap.SimpleEntry<>("A", Arrays.asList("a1"))), equalTo(true));
    }

    @Test
    public void testOkPayloadLimit() throws Exception {
        byte[] bytes = Base64.getEncoder().encode(withLength(new byte[0]));
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        doReturn("4").when(mockRequest).getHeader("x-frugal-payload-limit");

        servlet.service(mockRequest, mockResponse);

        verify(mockResponse).setContentType("application/x-frugal");
        verify(mockResponse).setHeader("Content-Transfer-Encoding", "base64");
        verify(mockResponse).getOutputStream();
    }

    @Test
    public void testWriteException() throws Exception {
        ServletOutputStream servletOut = new ProxyServletOutputStream(new OutputStream() {
            @Override
            public void write(int b) throws IOException {
                throw new IOException("test");
            }
        });
        doReturn(servletOut).when(mockResponse).getOutputStream();

        byte[] bytes = Base64.getEncoder().encode(withLength(new byte[0]));
        ByteArrayInputStream in = new ByteArrayInputStream(bytes);
        doReturn(new ProxyServletInputStream(in)).when(mockRequest).getInputStream();

        IOException e = assertThrows(IOException.class, () -> servlet.service(mockRequest, mockResponse));
        assertThat(e.getSuppressed(), emptyArray());

        verify(mockResponse).setContentType("application/x-frugal");
        verify(mockResponse).setHeader("Content-Transfer-Encoding", "base64");
        verify(mockResponse).getOutputStream();
    }
}
