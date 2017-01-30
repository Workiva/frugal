package com.workiva.frugal.processor;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.FrugalTApplicationExceptionType;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TMessage;
import org.apache.thrift.protocol.TMessageType;
import org.apache.thrift.protocol.TProtocolUtil;
import org.apache.thrift.protocol.TType;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.HashMap;
import java.util.Map;

/**
 * Abstract base FProcessor implementation. This should only be used by generated code.
 */
public abstract class FBaseProcessor implements FProcessor {

    private static final Logger LOGGER = LoggerFactory.getLogger(FBaseProcessor.class);
    protected static final Object WRITE_LOCK = new Object();

    private Map<String, FProcessorFunction> processMap;
    private Map<String, Map<String, String>> annotationsMap;

    @Override
    public void process(FProtocol iprot, FProtocol oprot) throws TException {
        if (processMap == null) {
            processMap = getProcessMap();
        }
        FContext ctx = iprot.readRequestHeader();
        TMessage message = iprot.readMessageBegin();
        FProcessorFunction processor = processMap.get(message.name);
        if (processor != null) {
            try {
                processor.process(ctx, iprot, oprot);
            } catch (TException e) {
                // Don't raise an exception because the server should still send a response to the client.
                LOGGER.error("Exception occurred while processing request with correlation id "
                        + ctx.getCorrelationId(), e);
            }
            return;
        }

        LOGGER.warn(String.format("Client invoked unknown method %s on request with correlation id %s",
                message.name, ctx.getCorrelationId()));
        TProtocolUtil.skip(iprot, TType.STRUCT);
        iprot.readMessageEnd();
        TApplicationException e = new TApplicationException(
                FrugalTApplicationExceptionType.UNKNOWN_METHOD, "Unknown function " + message.name);
        synchronized (WRITE_LOCK) {
            oprot.writeResponseHeader(ctx);
            oprot.writeMessageBegin(new TMessage(message.name, TMessageType.EXCEPTION, 0));
            e.write(oprot);
            oprot.writeMessageEnd();
            oprot.getTransport().flush();
        }
    }

    /**
     * Returns the map of method names to FProcessorFunctions.
     *
     * @return FProcessorFunction map
     */
    protected abstract Map<String, FProcessorFunction> getProcessMap();

    /**
     * Returns the map of method names to annotations.
     *
     * @return annotations map
     */
    protected abstract Map<String, Map<String, String>> getAnnotationsMap();

    @Override
    public Map<String, Map<String, String>> getAnnotations() {
        if (annotationsMap == null) {
            annotationsMap = getAnnotationsMap();
        }
        return new HashMap<>(annotationsMap);
    }
}
