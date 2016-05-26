package com.workiva.frugal.protocol;

import com.workiva.frugal.exception.FException;
import com.workiva.frugal.internal.Headers;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;
import java.util.logging.Logger;


/**
 * FClientRegistry is intended for use only by Frugal clients.
 * This is only to be used by generated code.
 */
public class FClientRegistry implements FRegistry {

    private static final Logger LOGGER = Logger.getLogger(FClientRegistry.class.getName());
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
        if (handlers.containsKey(context.getOpId())) {
            throw new FException("context already registered");
        }
        long opId = NEXT_OP_ID.incrementAndGet();
        context.setOpId(opId);
        handlers.put(opId, new Pair<>(callback, Thread.currentThread()));
    }

    /**
     * Unregister the callback for the given FContext.
     *
     * @param context the FContext to unregister.
     */
    public void unregister(FContext context) {
        handlers.remove(context.getOpId());
    }

    /**
     * Dispatch a single Frugal message frame.
     *
     * @param frame an entire Frugal message frame.
     */
    public void execute(byte[] frame) throws TException {
        Map<String, String> headers;
        headers = Headers.decodeFromFrame(frame);

        long opId;
        try {
            opId = Long.parseLong(headers.get(FContext.OP_ID));
        } catch (NumberFormatException e) {
            throw new FException("frame missing opId");
        }
        if (!handlers.containsKey(opId)) {
            LOGGER.info("Got a message for an unregistered context. Dropping.");
            return;
        }

        Pair<FAsyncCallback, Thread> pair = handlers.get(opId);
        pair.first.onMessage(new TMemoryInputTransport(frame));
    }

    /**
     * Interrupt any registered contexts.
     */
    public void close() {
        for (Pair<FAsyncCallback, Thread> pair : handlers.values()) {
            pair.second.interrupt();
        }
    }

    private static class Pair<K, V> {
        K first;
        V second;

        Pair(K first, V second) {
            this.first = first;
            this.second = second;
        }
    }
}
