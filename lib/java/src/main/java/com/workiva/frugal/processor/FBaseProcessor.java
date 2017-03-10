package com.workiva.frugal.processor;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TApplicationExceptionType;
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
            } catch (RuntimeException e) {
                LOGGER.error("User handler code threw unhandled exception on request with correlation id"
                        + ctx.getCorrelationId(), e);
                synchronized (WRITE_LOCK) {
                    writeApplicationException(ctx, oprot, TApplicationExceptionType.INTERNAL_ERROR, message.name,
                            "Internal error processing " + message.name);
                }
                throw e;
            }
            return;
        }

        LOGGER.warn(String.format("Client invoked unknown method %s on request with correlation id %s",
                message.name, ctx.getCorrelationId()));
        TProtocolUtil.skip(iprot, TType.STRUCT);
        iprot.readMessageEnd();
        synchronized (WRITE_LOCK) {
            writeApplicationException(ctx, oprot, TApplicationExceptionType.UNKNOWN_METHOD, message.name,
                    "Unknown function " + message.name);
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

    /**
     * Write a TApplicationException out to the given protocol with the given type, method, and message.
     *
     * @param ctx FContext associated with the request
     * @param oprot FProtocol tied to the client
     * @param type TApplicationErrorType for the exception
     * @param method the method name of the reqeust
     * @param message the error message to put on the exception
     * @return the written TAppicationException
     * @throws TException If there is a problem writiing to the protocol
     */
    protected static TApplicationException writeApplicationException(FContext ctx, FProtocol oprot, int type,
                                                                     String method, String message) throws TException {
        TApplicationException x = new TApplicationException(type, message);

        // TODO: upgrade to thrift 0.10 and include
        // https://github.com/apache/thrift/commit/1d4a4393c9a9396ec76c3ba674e0d6a65fe39cc1
        oprot.reset();
        oprot.writeResponseHeader(ctx);
        oprot.writeMessageBegin(new TMessage(method, TMessageType.EXCEPTION, 0));
        x.write(oprot);
        oprot.writeMessageEnd();
        oprot.getTransport().flush();
        return x;
    }
}
