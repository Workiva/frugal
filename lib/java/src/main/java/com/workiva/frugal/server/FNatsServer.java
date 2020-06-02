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

package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import com.workiva.frugal.util.BlockingRejectedExecutionHandler;
import io.nats.client.Connection;
import io.nats.client.Dispatcher;
import io.nats.client.MessageHandler;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Duration;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Phaser;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

import static com.workiva.frugal.transport.FNatsTransport.NATS_MAX_MESSAGE_SIZE;

/**
 * An implementation of FServer which uses NATS as the underlying transport.
 * Clients must connect with the FNatsTransport.
 */
public class FNatsServer implements FServer {

    private static final Logger LOGGER = LoggerFactory.getLogger(FNatsServer.class);
    public static final int DEFAULT_WORK_QUEUE_LEN = 64;
    public static final int DEFAULT_WATERMARK = 5000;
    public static final long DEFAULT_STOP_TIMEOUT_NS = TimeUnit.SECONDS.toNanos(30);

    private final Connection conn;
    private final FProcessor processor;
    private final FProtocolFactory inputProtoFactory;
    private final FProtocolFactory outputProtoFactory;
    private final String[] subjects;
    private final String queue;
    private final long stopTimeoutNS;
    private final FServerEventHandler eventHandler;

    private final CountDownLatch shutdownSignal = new CountDownLatch(1);
    /**
     * Synchronizes shutdown between {@link #serve} and {@link #stop} threads.
     * The implied states are:
     * <ol>
     * <li>Serve - initial state; serve thread waits, and stop thread calls
     * {@link CountDownLatch#countDown()} on {@link FNatsServer#shutdownSignal} before
     * advancing to the next state
     * <li>Unsubscribe - stop thread waits, and serve thread unsubscribes before
     * advancing to the next state
     * <li>Shutdown executor service - serve thread waits, and stop thread awaits
     * executor service shutdown before advancing to the next state
     * <li>Shutdown - serve and stop threads both complete
     * </ol>
     */
    private final Phaser partiesAwaitingFullShutdown = new Phaser();
    private final ExecutorService executorService;

    /**
     * Creates a new FNatsServer which receives requests on the given subjects and queue.
     * <p>
     * The worker count controls the size of the thread pool used to process requests. This uses a provided queue
     * length. If the queue fills up, newly received requests will block to be placed on the queue. If requests wait for
     * too long based on the high watermark, the server will log that it is backed up. Clients must connect with the
     * FNatsTransport.
     *
     * @param conn            NATS connection
     * @param processor       FProcessor used to process requests
     * @param protoFactory    FProtocolFactory used for input and output protocols
     * @param subjects        NATS subjects to receive requests on
     * @param queue           NATS queue group to receive requests on
     * @param stopTimeoutNS Nanoseconds to await current requests to finish when stopping server
     * @param executorService Custom executor service for processing messages
     */
    private FNatsServer(Connection conn, FProcessor processor, FProtocolFactory protoFactory,
                        String[] subjects, String queue, ExecutorService executorService,
                        long stopTimeoutNS, FServerEventHandler eventHandler) {
        this.conn = conn;
        this.processor = processor;
        this.inputProtoFactory = protoFactory;
        this.outputProtoFactory = protoFactory;
        this.subjects = subjects;
        this.queue = queue;
        this.executorService = executorService;
        this.stopTimeoutNS = stopTimeoutNS;
        this.eventHandler = eventHandler;
    }

    /**
     * Builder for configuring and constructing FNatsServer instances.
     */
    public static class Builder {

        private final Connection conn;
        private final FProcessor processor;
        private final FProtocolFactory protoFactory;
        private final String[] subjects;

        private String queue = "";
        private int workerCount = 1;
        private int queueLength = DEFAULT_WORK_QUEUE_LEN;
        private long highWatermark = DEFAULT_WATERMARK;
        private ExecutorService executorService;
        private long stopTimeoutNS = DEFAULT_STOP_TIMEOUT_NS;
        private FServerEventHandler eventHandler;

        /**
         * Creates a new Builder which creates FStatelessNatsServers that subscribe to the given NATS subjects.
         *
         * @param conn         NATS connection
         * @param processor    FProcessor used to process requests
         * @param protoFactory FProtocolFactory used for input and output protocols
         * @param subjects     NATS subjects to receive requests on
         */
        public Builder(Connection conn, FProcessor processor, FProtocolFactory protoFactory, String[] subjects) {
            this.conn = conn;
            this.processor = processor;
            this.protoFactory = protoFactory;
            this.subjects = subjects;
        }

        /**
         * Adds a NATS queue group to receive requests on to the Builder.
         *
         * @param queue NATS queue group
         * @return Builder
         */
        public Builder withQueueGroup(String queue) {
            this.queue = queue;
            return this;
        }

        /**
         * Adds a worker count which controls the size of the thread pool used to process requests (defaults to 1).
         *
         * @param workerCount thread pool size
         * @return Builder
         */
        public Builder withWorkerCount(int workerCount) {
            this.workerCount = workerCount;
            return this;
        }

        /**
         * Adds a queue length which controls the size of the work queue buffering requests (defaults to 64).
         *
         * @param queueLength work queue length
         * @return Builder
         */
        public Builder withQueueLength(int queueLength) {
            this.queueLength = queueLength;
            return this;
        }

        /**
         * Set the executor service used to execute incoming processor tasks.
         * If set, overrides withQueueLength and withWorkerCount options.
         * <p>
         * Defaults to:
         * <pre>
         * {@code
         * new ThreadPoolExecutor(workerCount,
         *                        workerCount,
         *                        0,
         *                        TimeUnit.MILLISECONDS,
         *                        new ArrayBlockingQueue<>(queueLength),
         *                        new BlockingRejectedExecutionHandler());
         * }
         * </pre>
         *
         * This is similar to {@link java.util.concurrent.Executors#newFixedThreadPool},
         * but with a different {@link java.util.concurrent.RejectedExecutionHandler},
         * which will block until the task can be processed.
         *
         * @param executorService ExecutorService to run tasks
         * @return Builder
         */
        public Builder withExecutorService(ExecutorService executorService) {
            this.executorService = executorService;
            return this;
        }

        /**
         * Controls the high watermark which determines the time spent waiting in the queue before triggering slow
         * consumer logic.
         * If a server event handler is passed in, this value will not be used.
         *
         * @param highWatermark duration in milliseconds
         * @return Builder
         */
        public Builder withHighWatermark(long highWatermark) {
            this.highWatermark = highWatermark;
            return this;
        }

        public Builder withServerEventHandler(FServerEventHandler eventHandler) {
            this.eventHandler = eventHandler;
            return this;
        }

        /**
         * Controls how long to attempt wait for existing tasks to complete when stopping
         * the server via {@link FNatsServer#stop()}.
         *
         * @param timeout max duration to wait when stopping
         * @param unit unit of time for timeout
         * @return Builder
         */
        public Builder withStopTimeout(long timeout, TimeUnit unit) {
            this.stopTimeoutNS = unit.toNanos(timeout);
            return this;
        }

        /**
         * Creates a new configured FNatsServer.
         *
         * @return FNatsServer
         */
        public FNatsServer build() {
            if (executorService == null) {
                this.executorService = new ThreadPoolExecutor(
                        workerCount, workerCount, 0, TimeUnit.MILLISECONDS,
                        new ArrayBlockingQueue<>(queueLength),
                        new BlockingRejectedExecutionHandler());
            }

            if (eventHandler == null) {
                eventHandler = new FDefaultNatsServerEventHandler(highWatermark);
            }

            return new FNatsServer(conn, processor, protoFactory, subjects, queue,
                executorService, stopTimeoutNS, eventHandler);
        }

    }

    /**
     * Starts the server by subscribing to messages on the configured NATS subject.
     *
     * @throws TException if the server fails to start
     */
    @Override
    public void serve() throws TException {
        Dispatcher dispatcher = conn.createDispatcher(newRequestHandler());
        for (String subject : subjects) {
            if (queue != null && !queue.isEmpty()) {
                dispatcher.subscribe(subject, queue);
            } else {
                dispatcher.subscribe(subject);
            }
        }

        LOGGER.info("Frugal server running...");
        partiesAwaitingFullShutdown.register();
        try {
            boolean shutdownSignalReceived;
            try {
                shutdownSignal.await();
                // This shutdown signal may have been received from a prior call to stop if
                // serve was inadvertantly called on an already stopped server.
                shutdownSignalReceived = true;
            } catch (InterruptedException ignored) {
                shutdownSignalReceived = false;
            }
            LOGGER.info("Frugal server stopping...");

            // Closing the dispatcher will unsubscribe from all current subscriptions for that
            // dispatcher
            conn.closeDispatcher(dispatcher);
            // If serving stopped due to stop being called, we want to give the signal the stop method
            // that this serve has completed its actions and the stop method can move on to the next
            // phase of the shutdown. If shutdown signal was not received (i.e., serve is exiting
            // because the thread was interrupted, then we do not want to wait as other calls to serve
            // may not have been interrupted and we don't want to be blocked by them.
            if (shutdownSignalReceived) {
                // Advance to next state now that all unsubscribes are finished
                // A phaser is used here because while we expect callers to have only issued a
                // single call to serve we do not currently enforce that. If someone is running
                // serve on the same instance in different threads, then we want to ensure all
                // unsubscribes are finished before moving on to shutting down. The phaser allows
                // us to notify when each call to serve is ready to move on to the next phase.
                // While we could add in some type of enforcement for not allowing more than one
                // call to serve or a call to serve after stop, that would likely be a breaking
                // change to callers doing this now.
                partiesAwaitingFullShutdown.arriveAndAwaitAdvance();
                // Wait for full shutdown to finish (if this was a call to serve after stop was
                // already called, then the shutdown would already be finished. But since stop
                // never registered with the phaser, then this below method should immediately
                // return as no other parties would be waiting.)
                partiesAwaitingFullShutdown.arriveAndAwaitAdvance();
            }
        } finally {
            partiesAwaitingFullShutdown.arriveAndDeregister();
        }
    }

    /**
     * Stops the server by shutting down the executor service processing tasks.
     *
     * @throws TException if the server fails to stop
     */
    @Override
    public void stop() throws TException {
        partiesAwaitingFullShutdown.register();
        try {
            // Unblock serving threads to allow them to unsubscribe
            shutdownSignal.countDown();

            // Wait for all unsubscriptions to finish if serve has been called and is currently
            // serving. If serve has not been called or all calls to serve have already exited,
            // then this call should be non-blocking.
            partiesAwaitingFullShutdown.arriveAndAwaitAdvance();

            // Attempt to perform an orderly shutdown of the worker pool by trying to complete any in-flight requests.
            executorService.shutdown();
            try {
                if (!executorService.awaitTermination(this.stopTimeoutNS, TimeUnit.NANOSECONDS)) {
                    executorService.shutdownNow();
                }
            } catch (InterruptedException e) {
                executorService.shutdownNow();
                Thread.currentThread().interrupt();
            }
        } finally {
            // Signal serving threads that shutdown is complete
            partiesAwaitingFullShutdown.arriveAndDeregister();
        }
    }

    /**
     * Creates a new NATS MessageHandler which is invoked when a request is received.
     *
     * @return MessageHandler used for handling requests
     */
    protected MessageHandler newRequestHandler() {
        return message -> {
            String reply = message.getReplyTo();
            if (reply == null || reply.isEmpty()) {
                LOGGER.warn("Discarding invalid NATS request (no reply)");
                return;
            }

            Map<Object, Object> ephemeralProperties = new HashMap<>();
            this.eventHandler.onRequestReceived(ephemeralProperties);
            executorService.execute(
                    new Request(message.getData(), message.getReplyTo(), ephemeralProperties));
        };
    }

    /**
     * Runnable which encapsulates a request received by the server.
     */
    public class Request implements Runnable {

        final byte[] frameBytes;
        final String reply;
        final Map<Object, Object> ephemeralProperties;

        public Request(byte[] frameBytes, String reply, Map<Object, Object> ephemeralProperties) {
            this.frameBytes = frameBytes;
            this.reply = reply;
            this.ephemeralProperties = ephemeralProperties;
        }

        /**
         * The frame bytes of the request.
         *
         * @return the frame bytes
         */
        public byte[] getFrameBytes() {
            return frameBytes;
        }

        /**
         * The NATS subject for the response.
         *
         * @return the reply subject
         */
        public String getReply() {
            return reply;
        }

        /**
         * The metadata of the request.
         *
         * @return the ephemeralProperties
         */
        public Map<Object, Object> getEphemeralProperties() {
            return ephemeralProperties;
        }

        @Override
        public void run() {
            processRequest(this);
        }
    }

    /**
     * Processes a request and publishes a NATS message for the response.
     *
     * @param request the request
     */
    private void processRequest(Request request) {
        eventHandler.onRequestStarted(request.ephemeralProperties);

        try {
            // Read and process frame (exclude first 4 bytes which represent frame size).
            TTransport input = new TMemoryInputTransport(request.frameBytes, 4, request.frameBytes.length);
            TMemoryOutputBuffer output = new TMemoryOutputBuffer(NATS_MAX_MESSAGE_SIZE);

            try {
                FProtocol inputProto = inputProtoFactory.getProtocol(input);
                inputProto.setEphemeralProperties(request.ephemeralProperties);
                FProtocol outputProto = outputProtoFactory.getProtocol(output);
                processor.process(inputProto, outputProto);
            } catch (TException e) {
                LOGGER.error("error processing request", e);
                return;
            } catch (RuntimeException e) {
                try {
                    conn.publish(request.reply, output.getWriteBytes());
                    conn.flush(Duration.ofSeconds(60));
                } catch (Exception ignored) {
                }
                return;
            }

            if (!output.hasWriteData()) {
                return;
            }

            // Send response.
            conn.publish(request.reply, output.getWriteBytes());
        } finally {
            eventHandler.onRequestEnded(request.ephemeralProperties);
        }
    }

    /**
     * The NATS subject this server is listening on.
     *
     * @return the subject
     */
    public String[] getSubjects() {
        return subjects;
    }

    /**
     * The NATS queue group this server is listening on.
     *
     * @return the queue
     */
    public String getQueue() {
        return queue;
    }

    ExecutorService getExecutorService() {
        return executorService;
    }
}
