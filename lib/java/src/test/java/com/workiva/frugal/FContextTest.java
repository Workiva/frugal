package com.workiva.frugal;

import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import java.util.HashMap;
import java.util.Map;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotEquals;
import static org.junit.Assert.assertNull;

/**
 * Tests for {@link FContext}.
 */
@RunWith(JUnit4.class)
public class FContextTest {

    @Test
    public void testGenerateOpId() {
        assertNotEquals(
                new FContext().getRequestHeader(FContext.OPID_HEADER),
                new FContext().getRequestHeader(FContext.OPID_HEADER)
        );
    }

    @Test
    public void testGenerateCorrelationId() {
        FContext ctx = new FContext();
        assertNotEquals("", ctx.getCorrelationId());
    }

    @Test
    public void testCorrelationId() {
        String correlationId = "abc";
        FContext ctx = new FContext(correlationId);
        assertEquals(correlationId, ctx.getCorrelationId());
    }

    @Test
    public void testSetCorrelationId() {
        String correlationId = "abc";
        FContext ctx = new FContext("blah");
        ctx.addRequestHeader(FContext.CID_HEADER, correlationId);
        assertEquals(correlationId, ctx.getCorrelationId());
    }

    @Test
    public void testRequestHeader() {
        FContext ctx = new FContext();
        assertEquals(ctx, ctx.addRequestHeader("foo", "bar"));
        assertEquals(ctx, ctx.addRequestHeader(FContext.CID_HEADER, "123"));
        assertEquals("bar", ctx.getRequestHeader("foo"));
        assertNull(ctx.getRequestHeader("blah"));
    }

    @Test
    public void testAddRequestHeaders() {
        FContext ctx = new FContext();
        Map<String, String> headers = new HashMap<>();
        headers.put("foo", "bar");
        headers.put("baz", "qux");
        headers.put(FContext.CID_HEADER, "123");
        assertEquals(ctx, ctx.addRequestHeaders(headers));
        assertEquals("bar", ctx.getRequestHeader("foo"));
        assertEquals("qux", ctx.getRequestHeader("baz"));
    }

    @Test
    public void testResponseHeader() {
        FContext ctx = new FContext();
        assertEquals(ctx, ctx.addResponseHeader("foo", "bar"));
        assertEquals(ctx, ctx.addResponseHeader(FContext.OPID_HEADER, "1"));
        assertEquals("bar", ctx.getResponseHeader("foo"));
        assertNull(ctx.getResponseHeader("blah"));
    }

    @Test
    public void testAddResponseHeaders() {
        FContext ctx = new FContext();
        Map<String, String> headers = new HashMap<>();
        headers.put("foo", "bar");
        headers.put("baz", "qux");
        headers.put(FContext.OPID_HEADER, "1");
        assertEquals(ctx, ctx.addResponseHeaders(headers));
        assertEquals("bar", ctx.getResponseHeader("foo"));
        assertEquals("qux", ctx.getResponseHeader("baz"));
    }

    @Test
    public void testTimeout() {
        // Check default timeout (5 seconds).
        FContext ctx = new FContext();
        assertEquals(5000, ctx.getTimeout());
        assertEquals("5000", ctx.getRequestHeader(FContext.TIMEOUT_HEADER));

        // Set timeout and check expected values.
        ctx.setTimeout(10000);
        assertEquals(10000, ctx.getTimeout());
        assertEquals("10000", ctx.getRequestHeader(FContext.TIMEOUT_HEADER));
    }

}
