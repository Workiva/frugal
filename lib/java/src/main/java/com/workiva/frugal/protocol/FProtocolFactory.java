package com.workiva.frugal.protocol;

import org.apache.thrift.protocol.TProtocol;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransport;

/**
 * FProtocolFactory creates FProtocols.
 */
public class FProtocolFactory {
    private TProtocolFactory tProtocolFactory;

    public FProtocolFactory(TProtocolFactory tProtocolFactory) {
        this.tProtocolFactory = tProtocolFactory;
    }

    public FProtocol getProtocol(TTransport transport) {
        TProtocol protocol = tProtocolFactory.getProtocol(transport);
        return new FProtocol(protocol);
    }

}
