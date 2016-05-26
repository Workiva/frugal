package com.workiva.frugal.protocol;

import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransport;

/**
 * FProtocolFactory creates new FProtocol instances. It takes a TProtocolFactory
 * and a TTransport and returns an FProtocol which wraps a TProtocol produced by
 * the TProtocolFactory. The TProtocol itself wraps the provided TTransport. This
 * makes it easy to produce an FProtocol which uses any existing Thrift transports
 * and protocols in a composable manner.
 */
public class FProtocolFactory {

    private TProtocolFactory tProtocolFactory;

    public FProtocolFactory(TProtocolFactory tProtocolFactory) {
        this.tProtocolFactory = tProtocolFactory;
    }

    public FProtocol getProtocol(TTransport transport) {
        return new FProtocol(tProtocolFactory.getProtocol(transport));
    }

}
