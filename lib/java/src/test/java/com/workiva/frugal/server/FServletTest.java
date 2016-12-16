package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import org.junit.Before;
import org.junit.Ignore;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;
import org.mockito.Matchers;

import javax.servlet.RequestDispatcher;
import javax.servlet.ServletContext;
import javax.servlet.ServletException;
import javax.servlet.ServletInputStream;
import javax.servlet.ServletOutputStream;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import javax.servlet.http.HttpSession;
import java.io.ByteArrayInputStream;
import java.io.IOException;

import static org.mockito.Mockito.anyInt;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FServlet}.
 */
@RunWith(JUnit4.class)
public class FServletTest extends HttpServlet {

    private static HttpServletRequest request;
    private static HttpServletResponse response;
    private static FServlet servlet;
    private ServletInputStream mockServletInputStream;
    private ServletOutputStream mockServletOutputStream;

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

        // Mock input
        byte[] myBinaryData = "TEST".getBytes();
        ByteArrayInputStream byteArrayInputStream = new ByteArrayInputStream(myBinaryData);
        mockServletInputStream = mock(ServletInputStream.class);
        when(mockServletInputStream.read(Matchers.any(), anyInt(), anyInt())).thenAnswer(invocationOnMock -> {
            Object[] args = invocationOnMock.getArguments();
            byte[] output = (byte[]) args[0];
            int offset = (int) args[1];
            int length = (int) args[2];
            return byteArrayInputStream.read(output, offset, length);
        });

        // Mock output
        mockServletOutputStream = mock(ServletOutputStream.class);
    }

    /**
     * Verifies that the doPost method throws an exception when passed null arguments.
     */
    @Ignore
    @Test(expected = NullPointerException.class)
    public final void testDoPostPositive() throws ServletException, IOException {
        servlet.doPost(null, null);
    }

    /**
     * Verifies that the doPost method runs without exception when data is available.
     */
    @Ignore
    @Test
    public final void testDoPostNegative() throws ServletException, IOException {
        when(request.getInputStream()).thenReturn(mockServletInputStream);
        when(response.getOutputStream()).thenReturn(mockServletOutputStream);
        servlet.doPost(request, response);
    }
}
