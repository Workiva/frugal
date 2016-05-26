package com.workiva.frugal.protocol;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotEquals;
import static org.junit.Assert.assertNull;

import org.junit.Test;

public class FContextTest {

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
    public void testRequestHeader() {
        FContext ctx = new FContext();
        ctx.addRequestHeader("foo", "bar");
        assertEquals("bar", ctx.getRequestHeader("foo"));
        assertNull(ctx.getRequestHeader("blah"));
    }

    @Test
    public void testResponseHeader() {
        FContext ctx = new FContext();
        ctx.addResponseHeader("foo", "bar");
        assertEquals("bar", ctx.getResponseHeader("foo"));
        assertNull(ctx.getResponseHeader("blah"));
    }

}
