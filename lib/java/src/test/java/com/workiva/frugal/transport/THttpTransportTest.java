package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;
import org.apache.commons.codec.binary.Base64;
import org.apache.http.*;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.message.BasicHttpResponse;
import org.apache.http.util.EntityUtils;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.mockito.ArgumentCaptor;

import java.io.IOException;
import java.util.concurrent.TimeUnit;

import static org.mockito.Matchers.any;
import static org.junit.Assert.*;
import static org.mockito.Mockito.*;

public class THttpTransportTest {

    private CloseableHttpClient client;
    private String url = "http://foo.com";
    private THttpTransport transport;

    @Before
    public void setUp() {
        client = mock(CloseableHttpClient.class);
        transport = new THttpTransport.Builder(client, url).build();
    }

    @Test
    public void testOpenClose() throws TTransportException, IOException, InterruptedException {
        assertFalse(transport.isOpen());
        transport.open();
        assertTrue(transport.isOpen());
        transport.close();
        assertEquals(THttpTransport.FRAME_BUFFER_CLOSED, transport.frameBuffer.poll(1, TimeUnit.SECONDS));
        assertFalse(transport.isOpen());
    }

    @Test(expected = TTransportException.class)
    public void testRead_notOpen() throws TTransportException {
        transport.read(new byte[5], 0, 5);
    }

    @Test
    public void testRead() throws TTransportException, InterruptedException {
        transport.open();

        transport.frameBuffer.put("hello".getBytes());
        transport.frameBuffer.put("world".getBytes());

        byte[] buff = new byte[5];
        assertEquals(5, transport.read(buff, 0, 5));
        assertArrayEquals("hello".getBytes(), buff);
        assertEquals(5, transport.read(buff, 0, 5));
        assertArrayEquals("world".getBytes(), buff);

        transport.frameBuffer.put(THttpTransport.FRAME_BUFFER_CLOSED);

        try {
            transport.read(buff, 0, 5);
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(TTransportException.END_OF_FILE, e.getType());
        }
    }

    @Test(expected = TTransportException.class)
    public void testWrite_notOpen() throws TTransportException {
        transport.write(new byte[5]);
    }

    @Test(expected = FMessageSizeException.class)
    public void testWrite_sizeException() throws TTransportException {
        int requestSizeLimit = 1024 * 4;
        transport = new THttpTransport.Builder(client, url).withRequestSizeLimit(requestSizeLimit).build();
        transport.open();
        transport.write(new byte[requestSizeLimit + 1]);
    }

    @Test
    public void testWriteFlush() throws TTransportException, IOException {
        int responseSizeLimit = 1024 * 4;
        transport = new THttpTransport.Builder(client, url).withResponseSizeLimit(responseSizeLimit).build();
        transport.open();

        byte[] buff = "helloserver".getBytes();
        transport.write(buff);

        StatusLine statusLine = new StatusLineImpl(HttpVersion.HTTP_1_1, HttpStatus.SC_OK, null);
        byte[] responsePayload = "helloclient".getBytes();
        String encoded = Base64.encodeBase64String(responsePayload);
        StringEntity responseEntity = new StringEntity(encoded, ContentType.create("application/x-frugal", "utf-8"));

        CloseableHttpResponse response = new BasicClosableHttpResponse(statusLine);
        response.setEntity(responseEntity);

        ArgumentCaptor<HttpPost> topicCaptor = ArgumentCaptor.forClass(HttpPost.class);
        when(client.execute(topicCaptor.capture())).thenReturn(response);

        transport.flush();

        HttpPost actual = topicCaptor.getValue();
        HttpPost expected = validRequest(buff, responseSizeLimit);
        assertEquals(EntityUtils.toString(expected.getEntity()), EntityUtils.toString(actual.getEntity()));
        assertEquals(expected.getFirstHeader("content-type"), actual.getFirstHeader("content-type"));
        assertEquals(expected.getURI(), actual.getURI());

        byte[] responseActual = new byte[responsePayload.length];
        transport.read(responseActual, 0, responseActual.length);

        assertArrayEquals(responsePayload, responseActual);
    }

    @Test(expected = TTransportException.class)
    public void testFlush_notOpen() throws TTransportException {
        transport.flush();
    }

    @Test
    public void testFlush_noData() throws TTransportException, IOException {
        transport.open();
        transport.flush();
        verify(client, times(0)).execute(any(HttpPost.class));
    }

    @Test(expected = TTransportException.class)
    public void testFlush_requestIOException() throws TTransportException, IOException {
        transport.open();
        byte[] buff = "helloserver".getBytes();
        transport.write(buff);

        when(client.execute(any(HttpPost.class))).thenThrow(new IOException());
        transport.flush();
    }

    @Test(expected = FMessageSizeException.class)
    public void testFlush_responseTooLarge() throws TTransportException, IOException {
        int responseSizeLimit = 1024 * 4;
        transport = new THttpTransport.Builder(client, url).withResponseSizeLimit(responseSizeLimit).build();
        transport.open();

        byte[] buff = "helloserver".getBytes();
        transport.write(buff);

        StatusLine statusLine = new StatusLineImpl(HttpVersion.HTTP_1_1, HttpStatus.SC_REQUEST_TOO_LONG, null);
        CloseableHttpResponse response = new BasicClosableHttpResponse(statusLine);

        ArgumentCaptor<HttpPost> topicCaptor = ArgumentCaptor.forClass(HttpPost.class);
        when(client.execute(topicCaptor.capture())).thenReturn(response);

        transport.flush();
    }

    @Test(expected = TTransportException.class)
    public void testFlush_responseBadStatus() throws TTransportException, IOException {
        transport = new THttpTransport.Builder(client, url).build();
        transport.open();

        byte[] buff = "helloserver".getBytes();
        transport.write(buff);

        StatusLine statusLine = new StatusLineImpl(HttpVersion.HTTP_1_1, HttpStatus.SC_BAD_REQUEST, null);
        CloseableHttpResponse response = new BasicClosableHttpResponse(statusLine);

        ArgumentCaptor<HttpPost> topicCaptor = ArgumentCaptor.forClass(HttpPost.class);
        when(client.execute(topicCaptor.capture())).thenReturn(response);

        transport.flush();
    }

    @Test(expected = TTransportException.class)
    public void testFlush_badResponseLength() throws TTransportException, IOException {
        transport = new THttpTransport.Builder(client, url).build();
        transport.open();

        byte[] buff = "helloserver".getBytes();
        transport.write(buff);

        StatusLine statusLine = new StatusLineImpl(HttpVersion.HTTP_1_1, HttpStatus.SC_OK, null);
        byte[] responsePayload = new byte[1];
        String encoded = Base64.encodeBase64String(responsePayload);
        StringEntity responseEntity = new StringEntity(encoded, ContentType.create("application/x-frugal", "utf-8"));

        CloseableHttpResponse response = new BasicClosableHttpResponse(statusLine);
        response.setEntity(responseEntity);

        ArgumentCaptor<HttpPost> topicCaptor = ArgumentCaptor.forClass(HttpPost.class);
        when(client.execute(topicCaptor.capture())).thenReturn(response);

        transport.flush();
    }

    @Test(expected = TTransportException.class)
    public void testFlush_missingData() throws TTransportException, IOException {
        transport = new THttpTransport.Builder(client, url).build();
        transport.open();

        byte[] buff = "helloserver".getBytes();
        transport.write(buff);

        StatusLine statusLine = new StatusLineImpl(HttpVersion.HTTP_1_1, HttpStatus.SC_OK, null);
        byte[] responsePayload = new byte[] {(byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x01};
        String encoded = Base64.encodeBase64String(responsePayload);
        StringEntity responseEntity = new StringEntity(encoded, ContentType.create("application/x-frugal", "utf-8"));

        CloseableHttpResponse response = new BasicClosableHttpResponse(statusLine);
        response.setEntity(responseEntity);

        ArgumentCaptor<HttpPost> topicCaptor = ArgumentCaptor.forClass(HttpPost.class);
        when(client.execute(topicCaptor.capture())).thenReturn(response);

        transport.flush();
    }

    private HttpPost validRequest(byte[] payload, int responseSizeLimit) {
        // Encode request payload
        String encoded = Base64.encodeBase64String(payload);
        StringEntity requestEntity = new StringEntity(encoded, ContentType.create("application/x-frugal", "utf-8"));

        // Set headers and payload
        HttpPost request = new HttpPost(url);
        request.setHeader("accept", "application/x-frugal");
        request.setHeader("content-transfer-encoding", "base64");
        if (responseSizeLimit > 0) {
            request.setHeader("x-frugal-payload-limit", Integer.toString(responseSizeLimit));
        }
        request.setEntity(requestEntity);
        return request;
    }

    private class StatusLineImpl implements StatusLine {
        private ProtocolVersion protocolVersion;
        private int statusCode;
        private String reasonPhrase;

        StatusLineImpl(ProtocolVersion protocolVersion, int statusCode, String reasonPhrase) {
            this.protocolVersion = protocolVersion;
            this.statusCode = statusCode;
            this.reasonPhrase = reasonPhrase;
        }

        @Override
        public ProtocolVersion getProtocolVersion() {
            return protocolVersion;
        }

        @Override
        public int getStatusCode() {
            return statusCode;
        }

        @Override
        public String getReasonPhrase() {
            return reasonPhrase;
        }
    }

    private class BasicClosableHttpResponse extends BasicHttpResponse implements CloseableHttpResponse {

        BasicClosableHttpResponse(StatusLine statusline) {
            super(statusline);
        }

        public void close() throws IOException { }
    }
}
