package com.workiva.frugal.protocol;

import com.workiva.frugal.processor.FProcessor;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;

/**
 * FServerRegistry is intended for use only by Frugal servers.
 * This is only to be used by generated code.
 */
public class FServerRegistry implements FRegistry {
    FProcessor fProcessor;
    FProtocolFactory inputProtocolFactory;
    FProtocol outputProtocol;

    public FServerRegistry(FProcessor fProcessor, FProtocolFactory inputProtocolFactory,
                           FProtocol outputProtocol) {
        this.fProcessor = fProcessor;
        this.inputProtocolFactory = inputProtocolFactory;
        this.outputProtocol = outputProtocol;
    }

    /**
     * Register a callback for the given FContext.
     * THIS IS A NO-OP FOR FServerRegistry.
     *
     * @param context the FContext to register.
     * @param callback the callback to register.
     */
    public void register(FContext context, FAsyncCallback callback) {}

    /**
     * Unregister the callback for the given FContext.
     * THIS IS A NO-OP FOR FServerRegistry.
     *
     * @param context the FContext to unregister.
     */
    public void unregister(FContext context) {}

    /**
     * Dispatch a single Frugal message frame.
     *
     * @param frame an entire Frugal message frame.
     */
    public void execute(byte[] frame) throws TException {
        fProcessor.process(
                inputProtocolFactory.getProtocol(new TMemoryInputTransport(frame)), outputProtocol);
    }

    public void close() {}
}
