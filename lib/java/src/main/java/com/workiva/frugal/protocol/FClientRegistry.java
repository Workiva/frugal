package com.workiva.frugal.protocol;

import com.workiva.frugal.exception.FException;
import com.workiva.frugal.util.Pair;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;


/**
 * FClientRegistry is intended for use only by Frugal clients.
 * This is only to be used by generated code.
 */
public class FClientRegistry implements FRegistry {

    private static final Logger LOGGER = LoggerFactory.getLogger(FClientRegistry.class);
    private static final AtomicLong NEXT_OP_ID = new AtomicLong(0);

    protected Map<Long, Pair<FAsyncCallback, Thread>> handlers;

    public FClientRegistry() {
        handlers = new ConcurrentHashMap<>();
    }

    /**
     * Register a callback for the given FContext.
     *
     * @param context  the FContext to register.
     * @param callback the callback to register.
     */
    public void register(FContext context, FAsyncCallback callback) throws TException {
        // An FContext can be reused for multiple requests. Because of this, every
        // time an FContext is registered, it must be assigned a new op id to
        // ensure we can properly correlate responses. We use a monotonically
        // increasing atomic uint64 for this purpose. If the FContext already has
        // an op id, it has been used for a request. We check the handlers map to
        // ensure that request is not still in-flight.
        if (handlers.containsKey(context.getOpId())) {
            throw new FException("context already registered");
        }
        long opId = NEXT_OP_ID.incrementAndGet();
        context.setOpId(opId);
        handlers.put(opId, Pair.of(callback, Thread.currentThread()));
    }

    /**
     * Unregister the callback for the given FContext.
     *
     * @param context the FContext to unregister.
     */
    public void unregister(FContext context) {
        if (context == null) {
            return;
        }
        handlers.remove(context.getOpId());
    }

    /**
     * Dispatch a single Frugal message frame.
     *
     * @param frame an entire Frugal message frame.
     */
    public void execute(byte[] frame) throws TException {
        Map<String, String> headers;
        headers = HeaderUtils.decodeFromFrame(frame);

        long opId;
        try {
            opId = Long.parseLong(headers.get(FContext.OP_ID));
        } catch (NumberFormatException e) {
            throw new FException("frame missing opId");
        }

        Pair<FAsyncCallback, Thread> callbackThreadPair = handlers.get(opId);
        if (callbackThreadPair == null) {
            LOGGER.info("Got a message for an unregistered context. Dropping.");
            return;
        }
        callbackThreadPair.getLeft().onMessage(new TMemoryInputTransport(frame));
    }

    /**
     * Interrupt any registered contexts.
     */
    public void close() {
        handlers.values().parallelStream().forEach(pair -> pair.getRight().interrupt());
        handlers.clear();
    }
}
