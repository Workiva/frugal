package com.workiva.frugal.processor;

import com.workiva.frugal.FContext;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TField;
import org.apache.thrift.protocol.TMessage;
import org.apache.thrift.transport.TTransport;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import java.util.HashMap;
import java.util.Map;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FBaseProcessor}.
 */
@RunWith(JUnit4.class)
public class FBaseProcessorTest {

    private final String oneWay = "oneWay";
    private Map<String, FProcessorFunction> procMap;
    private FProcessorFunction oneWayFunction;
    private Map<String, Map<String, String>> annoMap;
    private FBaseProcessor processor;
    private FProtocol iprot;
    private FProtocol oprot;

    @Before
    public void setUp() throws Exception {
        procMap = new HashMap<>();
        annoMap = new HashMap<>();
        oneWayFunction = mock(FProcessorFunction.class);

        processor = new TestFProcessor(procMap, annoMap);

        iprot = mock(FProtocol.class);
        oprot = mock(FProtocol.class);
    }

    @Test
    public void testProcess() throws Exception {

        procMap.put(oneWay, oneWayFunction);

        FContext ctx = new FContext();
        TMessage thriftMessage = new TMessage(oneWay, (byte) 0x00, 1);

        when(iprot.readRequestHeader()).thenReturn(ctx);
        when(iprot.readMessageBegin()).thenReturn(thriftMessage);

        processor.process(iprot, oprot);

        verify(oneWayFunction).process(ctx, iprot, oprot);
    }

    @Test
    public void testProcessCatchTExceptionOnProcessorError() throws Exception {

        procMap.put(oneWay, oneWayFunction);

        FContext ctx = new FContext();
        doThrow(new TException("error")).when(oneWayFunction).process(ctx, iprot, oprot);
        TMessage thriftMessage = new TMessage(oneWay, (byte) 0x00, 1);

        when(iprot.readRequestHeader()).thenReturn(ctx);
        when(iprot.readMessageBegin()).thenReturn(thriftMessage);

        processor.process(iprot, oprot);

        verify(oneWayFunction).process(ctx, iprot, oprot);
    }

    @Test(expected = RuntimeException.class)
    public void testProcessCatchExceptionOnProcessorError() throws Exception {

        procMap.put(oneWay, oneWayFunction);

        FContext ctx = new FContext();
        doThrow(new Exception()).when(oneWayFunction).process(ctx, iprot, oprot);
        TMessage thriftMessage = new TMessage(oneWay, (byte) 0x00, 1);

        when(iprot.readRequestHeader()).thenReturn(ctx);
        when(iprot.readMessageBegin()).thenReturn(thriftMessage);

        processor.process(iprot, oprot);

        verify(oneWayFunction).process(ctx, iprot, oprot);
    }

    @Test
    public void testProcessCatchTApplicationExceptionOnUnknownMethod() throws Exception {
        TField tField = mock(TField.class);
        when(iprot.readFieldBegin()).thenReturn(tField);

        FContext ctx = new FContext();
        when(iprot.readRequestHeader()).thenReturn(ctx);

        TMessage thriftMessage = new TMessage("unknown", (byte) 0x00, 1);
        when(iprot.readMessageBegin()).thenReturn(thriftMessage);

        TTransport tTransport = mock(TTransport.class);
        when(oprot.getTransport()).thenReturn(tTransport);

        processor.process(iprot, oprot);
    }

    @Test
    public void testGetAnnotationsMap() {
        Map<String, String> fooMap = new HashMap<>();
        fooMap.put("foo", "bar");
        annoMap.put("baz", fooMap);

        Map<String, Map<String, String>> actualMap = processor.getAnnotationsMap();
        assertEquals(fooMap, actualMap.get("baz"));
    }

    private class TestFProcessor extends FBaseProcessor {

        private Map<String, FProcessorFunction> procMap;
        private Map<String, Map<String, String>> annoMap;

        public TestFProcessor(Map<String, FProcessorFunction> procMap,
                              Map<String, Map<String, String>> annoMap) {
            this.procMap = procMap;
            this.annoMap = annoMap;
        }

        @Override
        public Map<String, FProcessorFunction> getProcessMap() {
            return procMap;
        }

        @Override
        protected Map<String, Map<String, String>> getAnnotationsMap() {
            return annoMap;
        }

        @Override
        public void addMiddleware(ServiceMiddleware middleware) {
        }
    }
}
