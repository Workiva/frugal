package com.workiva.frugal.processor;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TMessage;
import org.junit.Test;

import java.util.HashMap;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

public class FBaseProcessorTest {

    private final String oneWay = "oneWay";

    @Test
    public void testProcess() throws Exception {

        HashMap<String, FProcessorFunction> map = new HashMap<>();
        FProcessorFunction oneWayFunction = new FProcessorFunction() {
            @Override
            public void process(FContext ctx, FProtocol in, FProtocol out) throws TException {
                return;
            }
        };
        map.put(oneWay, oneWayFunction);
        
        FBaseProcessor processor = new FBaseProcessor(map);
        
        FProtocol iprot = mock(FProtocol.class);
        FProtocol oprot = mock(FProtocol.class);

        FContext ctx = new FContext();
        TMessage thriftMessage = new TMessage(oneWay, (byte) 0x00, 1);

        when(iprot.readRequestHeader()).thenReturn(ctx);
        when(iprot.readMessageBegin()).thenReturn(thriftMessage);
        
        processor.process(iprot, oprot);
    }
}
