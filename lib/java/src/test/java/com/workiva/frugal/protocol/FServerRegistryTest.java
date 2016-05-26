package com.workiva.frugal.protocol;

import com.workiva.frugal.processor.FProcessor;
import org.apache.thrift.transport.TTransport;
import org.junit.Test;

import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

public class FServerRegistryTest {

    @Test
    public void testExecute() throws Exception {

        FProcessor fProcessor = mock(FProcessor.class);
        FProtocolFactory inputProtocolFactory = mock(FProtocolFactory.class);
        FProtocol outputProtocol = mock(FProtocol.class);

        FServerRegistry registry = new FServerRegistry(fProcessor, inputProtocolFactory, outputProtocol);

        FProtocol inputProtocol = mock(FProtocol.class);

        when(inputProtocolFactory.getProtocol(any(TTransport.class))).thenReturn(inputProtocol);

        registry.execute(new byte[] {});

        verify(fProcessor).process(inputProtocol, outputProtocol);
    }
}
