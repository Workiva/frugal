package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.commons.codec.binary.Base64;
import org.apache.thrift.TException;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import javax.servlet.RequestDispatcher;
import javax.servlet.ServletContext;
import javax.servlet.ServletException;
import javax.servlet.ServletOutputStream;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import javax.servlet.http.HttpSession;
import java.io.BufferedReader;
import java.io.IOException;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.equalTo;
import static org.hamcrest.Matchers.notNullValue;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.spy;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FServlet}.
 */
@RunWith(JUnit4.class)
public class FServletTest extends HttpServlet {

    @Rule
    public ExpectedException thrown = ExpectedException.none();

    private static HttpServletRequest request;
    private static HttpServletResponse response;
    private static FServlet servlet;

    @Before
    public final void setUp() throws IOException {
        request = mock(HttpServletRequest.class);
        response = mock(HttpServletResponse.class);
        HttpSession session = mock(HttpSession.class);
        when(request.getSession()).thenReturn(session);
        final ServletContext servletContext = mock(ServletContext.class);
        RequestDispatcher dispatcher = mock(RequestDispatcher.class);
        when(servletContext.getRequestDispatcher("/")).thenReturn(dispatcher);

        // Mock servlet
        FProcessor mockProcessor = mock(FProcessor.class);
        FProtocolFactory mockProtocolFactory = mock(FProtocolFactory.class);
        servlet = new FServlet(mockProcessor, mockProtocolFactory) {
            public ServletContext getServletContext() {
                return servletContext; // return the mock
            }
        };
    }

    @Test
    public final void testValidResponseLimit() {
        doReturn("2096").when(request).getHeader("x-frugal-payload-limit");

        Integer limit = servlet.getResponseLimit(request);
        assertThat(limit, equalTo(2096));
    }

    @Test
    public final void testNullResponseLimit() {
        doReturn(null).when(request).getHeader("x-frugal-payload-limit");

        Integer limit = servlet.getResponseLimit(request);
        assertThat(limit, equalTo(0));
    }

    @Test
    public final void testStringResponseLimit() {
        doReturn("not-a-number").when(request).getHeader("x-frugal-payload-limit");

        Integer limit = servlet.getResponseLimit(request);
        assertThat(limit, equalTo(0));
    }

    @Test
    public final void testGetInputBytesFromRequest() throws IOException {
        BufferedReader bufferedReader = mock(BufferedReader.class);
        when(bufferedReader.readLine()).thenReturn(
                Base64.encodeBase64String("request_body".getBytes()), null);
        doReturn(bufferedReader).when(request).getReader();

        byte[] bytes = servlet.getInputBytes(request);
        assertThat(bytes, notNullValue());
        assertThat(new String(bytes), equalTo("request_body"));
    }

    @Test
    public final void testGetInputBytesThrowsOnInvalidFrameSize() throws IOException {
        BufferedReader bufferedReader = mock(BufferedReader.class);
        when(bufferedReader.readLine()).thenReturn(
                Base64.encodeBase64String("r".getBytes()), null);
        doReturn(bufferedReader).when(request).getReader();

        thrown.expect(IOException.class);
        thrown.expectMessage("Invalid request frame");
        servlet.getInputBytes(request);
    }

    @Test
    public final void testGetInputBytesThrowsOnReaderError() throws IOException {
        BufferedReader bufferedReader = mock(BufferedReader.class);
        when(bufferedReader.readLine()).thenThrow(new IOException("Reader error"));
        doReturn(bufferedReader).when(request).getReader();

        thrown.expect(IOException.class);
        thrown.expectMessage("Reader error");
        servlet.getInputBytes(request);
    }

    @Test
    public final void test400OnInvalidFrame() throws ServletException, IOException {
        BufferedReader bufferedReader = mock(BufferedReader.class);
        when(bufferedReader.readLine()).thenThrow(new IOException("Reader error"));
        doReturn(bufferedReader).when(request).getReader();

        servlet.doPost(request, response);
        verify(response).setStatus(400);
    }

    @Test
    public final void test400OnSmallFrame() throws IOException, ServletException {
        BufferedReader bufferedReader = mock(BufferedReader.class);
        when(bufferedReader.readLine()).thenReturn(
                Base64.encodeBase64String("r".getBytes()), null);
        doReturn(bufferedReader).when(request).getReader();

        when(bufferedReader.readLine()).thenThrow(new IOException("Reader error"));

        servlet.doPost(request, response);
        verify(response).setStatus(400);
    }

    @Test
    public final void test400OnProcessorException() throws ServletException, IOException, TException {
        // given
        BufferedReader bufferedReader = mock(BufferedReader.class);
        when(bufferedReader.readLine()).thenReturn(
                Base64.encodeBase64String("request_body".getBytes()), null);
        doReturn(bufferedReader).when(request).getReader();

        // and
        FServlet spyServlet = spy(servlet);
        doThrow(new TException()).when(spyServlet).processFrame(any(byte[].class));

        // then
        spyServlet.doPost(request, response);
        verify(response).setStatus(400);
    }

    @Test
    public final void test413OverResponseLimit() throws ServletException, IOException, TException {
        // given
        doReturn("1").when(request).getHeader("x-frugal-payload-limit");

        // and
        BufferedReader bufferedReader = mock(BufferedReader.class);
        when(bufferedReader.readLine()).thenReturn(
                Base64.encodeBase64String("request_body".getBytes()), null);
        doReturn(bufferedReader).when(request).getReader();

        // and
        TMemoryOutputBuffer outputBuffer = mock(TMemoryOutputBuffer.class);
        when(outputBuffer.getWriteBytes()).thenReturn("request_body".getBytes());

        // and
        FServlet spyServlet = spy(servlet);
        doReturn(outputBuffer).when(spyServlet).processFrame("request_body".getBytes());

        // when
        spyServlet.doPost(request, response);

        // then
        verify(response).setStatus(413);
    }

    @Test
    public final void test200Ok() throws ServletException, IOException, TException {
        // given
        doReturn("100").when(request).getHeader("x-frugal-payload-limit");

        // and
        BufferedReader bufferedReader = mock(BufferedReader.class);
        when(bufferedReader.readLine()).thenReturn(
                Base64.encodeBase64String("request_body".getBytes()), null);
        doReturn(bufferedReader).when(request).getReader();

        // and
        TMemoryOutputBuffer outputBuffer = mock(TMemoryOutputBuffer.class);
        when(outputBuffer.getWriteBytes()).thenReturn("request_body".getBytes());

        // and
        FServlet spyServlet = spy(servlet);
        doReturn(outputBuffer).when(spyServlet).processFrame("request_body".getBytes());

        // and
        ServletOutputStream outputStream = mock(ServletOutputStream.class);
        doReturn(outputStream).when(response).getOutputStream();

        // when
        spyServlet.doPost(request, response);

        // then
        verify(response).setStatus(200);
    }
}
