/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.workiva.frugal.transport;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TTransportExceptionType;
import com.workiva.frugal.protocol.HeaderUtils;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TProtocolException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

import java.util.HashMap;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.TimeUnit;

/**
 * FAsyncTransport is an extension of FTransport that asynchronous frameworks can implement.
 * Implementations need only implement <code>flush</code> to send request data and call
 * <code>handleResponse</code> when asynchronous responses are received.
 */
public abstract class FAsyncTransport extends FTransport {

    protected static final byte[] POISON_PILL = new byte[0];
    // Visible for testing
    static final byte[] SERVICE_NOT_AVAILABLE = new byte[0];

    protected Map<Long, BlockingQueue<byte[]>> queueMap = new HashMap<>();

    /**
     * Interrupt all pending requests and signal close.
     */
    public void close() {
        close(null);
    }

    /**
     * Interrupt all pending requests and signal close.
     *
     * @param cause Exception if not a clean close (null otherwise)
     */
    protected void close(final Exception cause) {
        synchronized (this) {
            queueMap.values().parallelStream()
                    .filter(Objects::nonNull)
                    .forEach(queue -> queue.offer(POISON_PILL));
            queueMap.clear();
        }
        super.close(cause);
    }

    /**
     * Send the given framed frugal payload over the transport.
     *
     * @param context FContext associated with the request (used for timeout and logging)
     * @param payload framed frugal bytes
     * @throws TTransportException if the request times out or encounters other problems
     */
    public void oneway(FContext context, byte[] payload) throws TTransportException {
        preflightRequestCheck(payload.length);

        flush(payload);
    }

    /**
     * Send the given framed frugal payload over the transport and returns the response.
     *
     * @param context FContext associated with the request (used for timeout and logging)
     * @param payload framed frugal bytes
     * @return the response in TTransport form
     * @throws TTransportException if the request times out or encounters other problems
     */
    public TTransport request(FContext context, byte[] payload) throws TTransportException {
        preflightRequestCheck(payload.length);

        BlockingQueue<byte[]> queue = new ArrayBlockingQueue<>(1);
        long opId = getOpId(context);
        synchronized (this) {
            if (queueMap.containsKey(opId)) {
                throw new TTransportException("request already in flight for context");
            }
            queueMap.put(opId, queue);
        }

        try {
            flushOp(opId, payload);

            byte[] response;
            try {
                response = queue.poll(context.getTimeout(), TimeUnit.MILLISECONDS);
            } catch (InterruptedException e) {
                throw new TTransportException("request: interrupted");
            }

            if (response == null) {
                throw new TTransportException(TTransportExceptionType.TIMED_OUT, "request: timed out");
            }

            if (response == SERVICE_NOT_AVAILABLE) {
                throw new TTransportException(TTransportExceptionType.SERVICE_NOT_AVAILABLE,
                        "request: service not available");
            }

            if (response == POISON_PILL) {
                throw new TTransportException(TTransportExceptionType.NOT_OPEN,
                        "request: transport closed, request canceled");
            }

            return new TMemoryInputTransport(response);
        } finally {
            synchronized (this) {
                queueMap.remove(opId);
            }
        }
    }

    /**
     * Flush a oneway or request payload to the server.
     * Implementations must not block and must be thread-safe.
     * This method is not called for requests if {@link #flushOp} is overridden.
     *
     * @param payload framed frugal bytes
     * @throws TTransportException if flushing the transport fails.
     * @see #flushOp
     */
    protected abstract void flush(byte[] payload) throws TTransportException;

    /**
     * Flush a request payload to the server.
     * By default, this method calls {@link #flush}.
     */
    protected void flushOp(long opId, byte[] payload) throws TTransportException {
        flush(payload);
    }

    /**
     * Handles a frugal frame response (NOTE: this frame must NOT include the frame size).
     * Implementations should call this when asynchronous responses are received from the server.
     *
     * @param frame frugal frame
     * @throws TException if we cannot handle a response.
     */
    protected void handleResponse(byte[] frame) throws TException {
        Map<String, String> headers;
        headers = HeaderUtils.decodeFromFrame(frame);

        long opId;
        try {
            opId = Long.parseLong(headers.get(FContext.OPID_HEADER));
        } catch (NumberFormatException e) {
            throw new TProtocolException("invalid protocol frame: op id not a uint64", e);
        }

        handleOpResponse(opId, frame);
    }

    /**
     * Handles a frugal response for
     * {@link TTransportExceptionType#SERVICE_NOT_AVAILABLE}.
     */
    protected void handleServiceNotAvailable(long opId) throws TException {
        handleOpResponse(opId, SERVICE_NOT_AVAILABLE);
    }

    /**
     * Handles a frugal frame response for a specific op.
     */
    private void handleOpResponse(long opId, byte[] frame) throws TException {
        BlockingQueue<byte[]> queue;
        synchronized (this) {
            queue = queueMap.get(opId);
        }

        // Ignore unregistered frames
        if (queue == null) {
            return;
        }

        try {
            queue.put(frame);
        } catch (InterruptedException e) {
            throw new TException(e);
        }
    }

    /**
     * Returns the operation id for the FContext. This is a unique long per context. This is protected as operation
     * ids are an internal implementation detail.
     *
     * @param context FContext associated with the request (used for timeout and logging)
     *
     * @return operation id
     */
    protected static long getOpId(FContext context) {
        String opIdStr = context.getRequestHeader(FContext.OPID_HEADER);
        if (opIdStr == null) {
            return 0;
        }
        return Long.valueOf(opIdStr);
    }
}
