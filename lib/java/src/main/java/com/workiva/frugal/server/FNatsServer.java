package com.workiva.frugal.server;

import com.amazonaws.AmazonClientException;
import com.amazonaws.AmazonServiceException;
import com.amazonaws.services.s3.AmazonS3;
import com.amazonaws.services.s3.model.GetObjectRequest;
import com.amazonaws.services.s3.model.S3Object;
import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FBoundedMemoryBuffer;
import com.workiva.frugal.transport.JsonDataConverter;
import com.workiva.frugal.transport.MessageS3Pointer;
import com.workiva.frugal.util.BlockingRejectedExecutionHandler;
import com.workiva.frugal.util.ProtocolUtils;
import io.nats.client.Connection;
import io.nats.client.MessageHandler;
import io.nats.client.Subscription;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
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
    private final String subject;
    private final String queue;
    private final long highWatermark;

    private final CountDownLatch shutdownSignal = new CountDownLatch(1);
    private final ExecutorService executorService;
    private final AmazonS3 s3;
    private final String s3BucketName;
    private final boolean largePayloadSupport;

    /**
     * Creates a new FNatsServer which receives requests on the given subject and queue.
     * <p>
     * The worker count controls the size of the thread pool used to process requests. This uses a provided queue
     * length. If the queue fills up, newly received requests will block to be placed on the queue. If requests wait for
     * too long based on the high watermark, the server will log that it is backed up. Clients must connect with the
     * FNatsTransport.
     *
     * @param conn            NATS connection
     * @param processor       FProcessor used to process requests
     * @param protoFactory    FProtocolFactory used for input and output protocols
     * @param subject         NATS subject to receive requests on
     * @param queue           NATS queue group to receive requests on
     * @param highWatermark   Milliseconds when high watermark logic is triggered
     * @param executorService Custom executor service for processing messages
     */
    private FNatsServer(Connection conn, FProcessor processor, FProtocolFactory protoFactory,
                        String subject, String queue, long highWatermark, ExecutorService executorService,
                        AmazonS3 s3, String s3BucketName, boolean largePayloadSupport) {
        this.conn = conn;
        this.processor = processor;
        this.inputProtoFactory = protoFactory;
        this.outputProtoFactory = protoFactory;
        this.subject = subject;
        this.queue = queue;
        this.highWatermark = highWatermark;
        this.executorService = executorService;
        this.s3 = s3;
        this.s3BucketName = s3BucketName;
        this.largePayloadSupport = largePayloadSupport;
    }

    /**
     * Builder for configuring and constructing FNatsServer instances.
     */
    public static class Builder {

        private final Connection conn;
        private final FProcessor processor;
        private final FProtocolFactory protoFactory;
        private final String subject;

        private String queue = "";
        private int workerCount = 1;
        private int queueLength = DEFAULT_WORK_QUEUE_LEN;
        private long highWatermark = DEFAULT_WATERMARK;
        private ExecutorService executorService;
        private AmazonS3 s3;
        private String s3BucketName;
        private boolean largePayloadSupport = false;

        /**
         * Creates a new Builder which creates FStatelessNatsServers that subscribe to the given NATS subject.
         *
         * @param conn         NATS connection
         * @param processor    FProcessor used to process requests
         * @param protoFactory FProtocolFactory used for input and output protocols
         * @param subject      NATS subject to receive requests on
         */
        public Builder(Connection conn, FProcessor processor, FProtocolFactory protoFactory, String subject) {
            this.conn = conn;
            this.processor = processor;
            this.protoFactory = protoFactory;
            this.subject = subject;
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
         * Returns a new FNatsTransport with support for sending large messages.
         *
         * @param s3           Amazon S3 client used for storing large-payload messages.
         * @param s3BucketName Name of the bucket used for storing large-payload messages.
         *                     The bucket must be already created and configured in s3.
         */
        public Builder withLargePayloadEnabled(AmazonS3 s3, String s3BucketName) {
            this.s3 = s3;
            this.s3BucketName = s3BucketName;
            this.largePayloadSupport = true;
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
            FNatsServer server =
                    new FNatsServer(conn, processor, protoFactory, subject, queue, highWatermark, executorService,
                            s3, s3BucketName, largePayloadSupport);
            return server;
        }

    }

    /**
     * Starts the server by subscribing to messages on the configured NATS subject.
     *
     * @throws TException
     */
    @Override
    public void serve() throws TException {
        Subscription sub = conn.subscribe(subject, queue, newRequestHandler());
        LOGGER.info("Frugal server running...");
        try {
            shutdownSignal.await();
        } catch (InterruptedException ignored) {
        }
        LOGGER.info("Frugal server stopping...");

        try {
            sub.unsubscribe();
        } catch (IOException e) {
            LOGGER.warn("Frugal server failed to unsubscribe: " + e.getMessage());
        }
    }

    /**
     * Stops the server by shutting down the executor service processing tasks.
     *
     * @throws TException
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
                            highWatermark, inputProtoFactory, outputProtoFactory, processor, conn,
                            s3, s3BucketName, largePayloadSupport));
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
        final AmazonS3 s3;
        final String s3BucketName;
        final boolean largePayloadSupport;

        Request(byte[] frameBytes, long timestamp, String reply, long highWatermark,
                FProtocolFactory inputProtoFactory, FProtocolFactory outputProtoFactory,
                FProcessor processor, Connection conn, AmazonS3 s3, String s3BucketName, boolean largePayloadSupport) {
            this.frameBytes = frameBytes;
            this.timestamp = timestamp;
            this.reply = reply;
            this.highWatermark = highWatermark;
            this.inputProtoFactory = inputProtoFactory;
            this.outputProtoFactory = outputProtoFactory;
            this.processor = processor;
            this.conn = conn;
            this.s3 = s3;
            this.s3BucketName = s3BucketName;
            this.largePayloadSupport = largePayloadSupport;
        }

        @Override
        public void run() {
            long duration = System.currentTimeMillis() - timestamp;
            if (duration > highWatermark) {
                LOGGER.warn("frame spent " + duration + "ms in the transport buffer, your consumer might be backed up");
            }
            process();
        }

        private void process() {
            TTransport input;

            // Read the S3 pointer data
            if (largePayloadSupport) {
                String s3PointerJson = new String(frameBytes, StandardCharsets.UTF_8);
                MessageS3Pointer s3Pointer = readMessageS3PointerFromJSON(s3PointerJson);

                if (s3Pointer != null) {
                    // Data is an S3 pointer, retrieve message from S3
                    String s3MsgBucketName = s3Pointer.getS3BucketName();
                    String s3MsgKey = s3Pointer.getS3Key();

                    byte[] message = getMessageFromS3(s3, s3MsgBucketName, s3MsgKey);

                    input = new TMemoryInputTransport(message);
                } else {
                    // Read and process frame (exclude first 4 bytes which represent frame size).
                    byte[] frame = Arrays.copyOfRange(frameBytes, 4, frameBytes.length);

                    // Data is not an S3 pointer, read message directly
                    input = new TMemoryInputTransport(frame);
                }
            } else {
                // Read and process frame (exclude first 4 bytes which represent frame size).
                byte[] frame = Arrays.copyOfRange(frameBytes, 4, frameBytes.length);

                // Data is not an S3 pointer, read message directly
                input = new TMemoryInputTransport(frame);
            }

            // Buffer 1MB - 4 bytes since frame size is copied directly.
            FBoundedMemoryBuffer output = new FBoundedMemoryBuffer(NATS_MAX_MESSAGE_SIZE - 4);
            try {
                processor.process(inputProtoFactory.getProtocol(input), outputProtoFactory.getProtocol(output));
            } catch (TException e) {
                LOGGER.warn("error processing frame: " + e.getMessage());
                return;
            }

            if (output.length() == 0) {
                return;
            }

            // Add frame size (4-byte int32).
            byte[] response = new byte[output.length() + 4];
            ProtocolUtils.writeInt(output.length(), response, 0);
            System.arraycopy(output.getArray(), 0, response, 4, output.length());

            // Send response.
            try {
                conn.publish(reply, response);
            } catch (IOException e) {
                LOGGER.warn("failed to send response: " + e.getMessage());
            }
        }

        private MessageS3Pointer readMessageS3PointerFromJSON(String messageBody) {
            MessageS3Pointer s3Pointer = null;
            System.out.println("Receiving: " + messageBody);
            try {
                JsonDataConverter jsonDataConverter = new JsonDataConverter();
                s3Pointer = jsonDataConverter.deserializeFromJson(messageBody, MessageS3Pointer.class);
            } catch (Exception e) {
                // Failed to deserialize, most not be s3 pointer
                System.out.println(e.getMessage());
                return null;
            }
            System.out.println(s3Pointer.getS3BucketName());
            return s3Pointer;
        }

    }

    /**
     * The NATS subject this server is listening on.
     *
     * @return the subject
     */
    public String getSubject() {
        return subject;
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

    private static byte[] getMessageFromS3(AmazonS3 s3, String s3BucketName, String s3Key) {
        // Retrieve the object from S3
        GetObjectRequest getObjectRequest = new GetObjectRequest(s3BucketName, s3Key);
        S3Object obj = null;
        try {
            obj = s3.getObject(getObjectRequest);
        } catch (AmazonServiceException e) {
            String errorMessage = "Failed to get the S3 object which contains the message payload. Message was not received.";
            LOGGER.error(errorMessage, e);
            throw new AmazonServiceException(errorMessage, e);
        } catch (AmazonClientException e) {
            String errorMessage = "Failed to get the S3 object which contains the message payload. Message was not received.";
            LOGGER.error(errorMessage, e);
            throw new AmazonClientException(errorMessage, e);
        }

        // Parse the message
        byte[] message;
        try {
            InputStream objContent = obj.getObjectContent();
            message = getBytesFromInputStream(objContent);
        } catch (IOException e) {
            String errorMessage = "Failure when handling the message which was read from S3 object. Message was not received.";
            LOGGER.error(errorMessage, e);
            throw new AmazonClientException(errorMessage, e);
        }
        return message;
    }

    public static byte[] getBytesFromInputStream(InputStream is) throws IOException {
        try (ByteArrayOutputStream os = new ByteArrayOutputStream();) {
            byte[] buffer = new byte[0xFFFF];

            for (int len; (len = is.read(buffer)) != -1; )
                os.write(buffer, 0, len);

            os.flush();

            return os.toByteArray();
        }
    }
}
