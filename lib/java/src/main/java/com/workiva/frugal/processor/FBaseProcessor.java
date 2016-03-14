package com.workiva.frugal.processor;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TMessage;
import org.apache.thrift.protocol.TMessageType;
import org.apache.thrift.protocol.TProtocolUtil;
import org.apache.thrift.protocol.TType;

import java.util.Map;
import java.util.logging.Level;
import java.util.logging.Logger;

public class FBaseProcessor implements FProcessor {

    private static final Logger logger =
            Logger.getLogger(FBaseProcessor.class.getName());
    protected static final Object WRITE_LOCK = new Object();

    private final Map<String, FProcessorFunction> processMap;

    protected FBaseProcessor(Map<String, FProcessorFunction> processorFunctionMap) {
        this.processMap = processorFunctionMap;
    }

    @Override
    public void process(FProtocol iprot, FProtocol oprot) throws TException {
        FContext ctx = iprot.readRequestHeader();
        TMessage message = iprot.readMessageBegin();
        FProcessorFunction processor = this.processMap.get(message.name);
        if (processor != null) {
            try {
                processor.process(ctx, iprot, oprot);
            } catch (Exception e) {
                logger.log(Level.WARNING, "Error processing request with correlationID " + ctx.getCorrelationId() + ": " + e.getMessage());
                throw e;
            }
            return;
        }
        TProtocolUtil.skip(iprot, TType.STRUCT);
        iprot.readMessageEnd();
        TApplicationException e = new TApplicationException(TApplicationException.UNKNOWN_METHOD, "Unknown function " + message.name);
        synchronized (WRITE_LOCK) {
            oprot.writeResponseHeader(ctx);
            oprot.writeMessageBegin(new TMessage(message.name, TMessageType.EXCEPTION, 0));
            e.write(oprot);
            oprot.writeMessageEnd();
            oprot.getTransport().flush();
        }
        throw e;
    }
}
