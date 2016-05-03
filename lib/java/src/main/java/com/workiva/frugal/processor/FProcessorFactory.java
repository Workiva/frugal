package com.workiva.frugal.processor;

import org.apache.thrift.transport.TTransport;

/**
 * FProcessorFactory produces FProcessors and is used by an FServer. It takes a
 * TTransport and returns an FProcessor wrapping it.
 */
public class FProcessorFactory {

    private FProcessor processor;

    public FProcessorFactory(FProcessor processor) {
        this.processor = processor;
    }

    public FProcessor getProcessor(TTransport trans) {
        return processor;
    }

}
