package com.workiva.frugal.processor;

import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TException;

/**
 * FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic object
 * which operates upon an input stream and writes to an output stream.
 * Specifically, an FProcessor is provided to an FServer in order to wire up a
 * service handler to process requests.
 */
public interface FProcessor {

    void process(FProtocol in, FProtocol out) throws TException;

}
