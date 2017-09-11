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
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import com.workiva.frugal.util.BlockingRejectedExecutionHandler;
import io.nats.client.Connection;
import io.nats.client.MessageHandler;
import io.nats.client.Subscription;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
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

    private final Connection conn;
    private final FProcessor processor;
    private final FProtocolFactory inputProtoFactory;
    private final FProtocolFactory outputProtoFactory;
    private final String[] subjects;
    private final String queue;
    private final long highWatermark;

    private final CountDownLatch shutdownSignal = new CountDownLatch(1);
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
     * @param highWatermark   Milliseconds when high watermark logic is triggered
     * @param executorService Custom executor service for processing messages
     */
    private FNatsServer(Connection conn, FProcessor processor, FProtocolFactory protoFactory,
                        String[] subjects, String queue, long highWatermark, ExecutorService executorService) {
        this.conn = conn;
        this.processor = processor;
        this.inputProtoFactory = protoFactory;
        this.outputProtoFactory = protoFactory;
        this.subjects = subjects;
        this.queue = queue;
        this.highWatermark = highWatermark;
        this.executorService = executorService;
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
         * new ThreadPoolExecutor(1,
         *                        workerCount,
         *                        30,
         *                        TimeUnit.SECONDS,
         *                        new ArrayBlockingQueue<>(queueLength),
         *                        new BlockingRejectedExecutionHandler());
         * }
         * </pre>
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
         *
         * @param highWatermark duration in milliseconds
         * @return Builder
         */
        public Builder withHighWatermark(long highWatermark) {
            this.highWatermark = highWatermark;
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
                        1, workerCount, 30, TimeUnit.SECONDS,
                        new ArrayBlockingQueue<>(queueLength),
                        new BlockingRejectedExecutionHandler());
            }
            return new FNatsServer(conn, processor, protoFactory, subjects, queue, highWatermark, executorService);
        }

    }

    /**
     * Starts the server by subscribing to messages on the configured NATS subject.
     *
     * @throws TException if the server fails to start
     */
    @Override
    public void serve() throws TException {
        ArrayList<Subscription> subscriptionArrayList = new ArrayList<>();
        for (String subject : subjects) {
            subscriptionArrayList.add(conn.subscribe(subject, queue, newRequestHandler()));
        }

        LOGGER.info("Frugal server running...");
        try {
            shutdownSignal.await();
        } catch (InterruptedException ignored) {
        }
        LOGGER.info("Frugal server stopping...");

        for (Subscription subscription : subscriptionArrayList) {
            try {
                subscription.unsubscribe();
            } catch (IOException e) {
                LOGGER.warn("Frugal server failed to unsubscribe from " + subscription.getSubject() + ": " +
                        e.getMessage());
            }
        }
    }

    /**
     * Stops the server by shutting down the executor service processing tasks.
     *
     * @throws TException if the server fails to stop
     */
    @Override
    public void stop() throws TException {
        // Attempt to perform an orderly shutdown of the worker pool by trying to complete any in-flight requests.
        executorService.shutdown();
        try {
            if (!executorService.awaitTermination(30, TimeUnit.SECONDS)) {
                executorService.shutdownNow();
            }
        } catch (InterruptedException e) {
            executorService.shutdownNow();
            Thread.currentThread().interrupt();
        }

        // Unblock serving thread.
        shutdownSignal.countDown();
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

            executorService.submit(
                    new Request(message.getData(), System.currentTimeMillis(), message.getReplyTo(),
                            highWatermark, inputProtoFactory, outputProtoFactory, processor, conn));
        };
    }

    /**
     * Runnable which encapsulates a request received by the server.
     */
    static class Request implements Runnable {

        final byte[] frameBytes;
        final long timestamp;
        final String reply;
        final long highWatermark;
        final FProtocolFactory inputProtoFactory;
        final FProtocolFactory outputProtoFactory;
        final FProcessor processor;
        final Connection conn;

        Request(byte[] frameBytes, long timestamp, String reply, long highWatermark,
                FProtocolFactory inputProtoFactory, FProtocolFactory outputProtoFactory,
                FProcessor processor, Connection conn) {
            this.frameBytes = frameBytes;
            this.timestamp = timestamp;
            this.reply = reply;
            this.highWatermark = highWatermark;
            this.inputProtoFactory = inputProtoFactory;
            this.outputProtoFactory = outputProtoFactory;
            this.processor = processor;
            this.conn = conn;
        }

        @Override
        public void run() {
            long duration = System.currentTimeMillis() - timestamp;
            if (duration > highWatermark) {
                LOGGER.warn(String.format(
                        "request spent %d ms in the transport buffer, your consumer might be backed up", duration));
            }
            process();
        }

        private void process() {
            // Read and process frame (exclude first 4 bytes which represent frame size).
            byte[] frame = Arrays.copyOfRange(frameBytes, 4, frameBytes.length);
            TTransport input = new TMemoryInputTransport(frame);

            TMemoryOutputBuffer output = new TMemoryOutputBuffer(NATS_MAX_MESSAGE_SIZE);
            try {
                processor.process(inputProtoFactory.getProtocol(input), outputProtoFactory.getProtocol(output));
            } catch (TException e) {
                LOGGER.error("error processing request", e);
                return;
            } catch (RuntimeException e) {
                try {
                    conn.publish(reply, output.getWriteBytes());
                    conn.flush();
                } catch (Exception ignored) {
                }
                throw e;
            }

            if (!output.hasWriteData()) {
                return;
            }

            // Send response.
            try {
                conn.publish(reply, output.getWriteBytes());
            } catch (IOException e) {
                LOGGER.warn("failed to request response: " + e.getMessage());
            }
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
