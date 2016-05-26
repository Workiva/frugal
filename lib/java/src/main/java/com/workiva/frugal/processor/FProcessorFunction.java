package com.workiva.frugal.processor;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TException;

/**
 * FProcessorFunction is used internally by generated code. An FProcessor
 * registers an FProcessorFunction for each service method. Like FProcessor, an
 * FProcessorFunction exposes a single process call, which is used to handle a
 * method invocation.
 */
public interface FProcessorFunction {

    void process(FContext ctx, FProtocol in, FProtocol out) throws TException;

}
