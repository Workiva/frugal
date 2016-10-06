package com.workiva.frugal.transport;

import com.amazonaws.AmazonClientException;
import com.amazonaws.AmazonServiceException;
import com.amazonaws.services.s3.AmazonS3;
import com.amazonaws.services.s3.model.ObjectMetadata;
import com.amazonaws.services.s3.model.PutObjectRequest;
import com.workiva.frugal.exception.FMessageSizeException;
import io.nats.client.Connection;
import io.nats.client.Constants;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import io.nats.client.Subscription;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.UUID;

/**
 * FNatsTransport is an extension of FTransport. This is a "stateless" transport
 * in the sense that there is no connection with a server. A request is simply
 * published to a subject and responses are received on another subject. This
 * assumes requests/responses fit within a single NATS message.
 *
 * For delivering payloads larger than the 1MB NATS limit, you may configure an
 * S3 bucket for storing a payload.
 */
public class FNatsTransport extends FTransport {

    private static final Logger LOGGER = LoggerFactory.getLogger(FNatsTransport.class);

    public static final int NATS_MAX_MESSAGE_SIZE = 10; // for testing
    public static final String FRUGAL_PREFIX = "frugal.";

    private final Connection conn;
    private final String subject;
    private final String inbox;
    private final AmazonS3 s3;
	private final String s3BucketName;
	private final boolean largePayloadSupport;

    private Subscription sub;

    private FNatsTransport(Connection conn, String subject, String inbox, AmazonS3 s3, String s3BucketName, boolean largePayloadSupport) {
        // Leave room for the frame size
        super(NATS_MAX_MESSAGE_SIZE - 4);
        this.conn = conn;
        this.subject = subject;
        this.inbox = inbox;
        this.s3 = s3;
        this.s3BucketName = s3BucketName;
        this.largePayloadSupport = largePayloadSupport;
    }

    /**
     * Creates a new FTransport which uses the NATS messaging system as the underlying transport.
     * A request is simply published to a subject and responses are received on a randomly generated
     * subject. This requires requests to fit within a single NATS message.
     * <p>
     * This transport uses a randomly generated inbox for receiving NATS replies.
     *
     * @param conn    NATS connection
     * @param subject subject to publish requests on
     */
    public static FNatsTransport of(Connection conn, String subject) {
        return new FNatsTransport(conn, subject, conn.newInbox(), null, null, false);
    }

    /**
     * Returns a new FNatsTransport configured with the specified inbox.
     *
     * @param inbox NATS subject to receive responses on
     */
    public FNatsTransport withInbox(String inbox) {
        return new FNatsTransport(conn, subject, inbox, s3, s3BucketName, largePayloadSupport);
    }

    /**
     * Returns a new FNatsTransport with support for sending large messages.
	 *
	 * @param s3 Amazon S3 client used for storing large-payload messages.
	 * @param s3BucketName Name of the bucket used for storing large-payload messages.
     *                     The bucket must be already created and configured in s3.
     */
    public FNatsTransport withLargePayloadEnabled(AmazonS3 s3, String s3BucketName) {
        return new FNatsTransport(conn, subject, inbox, s3, s3BucketName, true);
    }

    /**
     * Query transport open state.
     *
     * @return true if transport and NATS connection are open.
     */
    @Override
    public boolean isOpen() {
        return sub != null && conn.getState() == Constants.ConnState.CONNECTED;
    }

    /**
     * Subscribes to the configured inbox subject.
     *
     * @throws TTransportException
     */
    @Override
    public void open() throws TTransportException {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            throw getClosedConditionException(conn.getState(), "open:");
        }
        if (sub != null) {
            throw new TTransportException(TTransportException.ALREADY_OPEN, "NATS transport already open");
        }
        sub = conn.subscribe(inbox, new Handler());
    }

    /**
     * Unsubscribes from the inbox subject and closes the response buffer.
     */
    @Override
    public void close() {
        if (sub == null) {
            return;
        }
        try {
            sub.unsubscribe();
        } catch (IOException e) {
            LOGGER.warn("NATS transport could not unsubscribe from subscription: " + e.getMessage());
        }
        sub = null;
        super.close();
    }

    /**
     * Sends any buffered bytes over NATS.
     *
     * @throws TTransportException
     */
    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw getClosedConditionException(conn.getState(), "flush:");
        }

        if (!hasWriteData()) {
            return;
        }

        byte[] data = getFramedWriteBytes();
        resetWriteBuffer();

        if (data.length > NATS_MAX_MESSAGE_SIZE && largePayloadSupport) {
            byte[] s3Pointer = storeMessageInS3(data);
            try {
                conn.publish(subject, inbox, s3Pointer);
            } catch (IOException e) {
                throw new TTransportException("flush: unable to publish data: " + e.getMessage());
            }
        } else {
            try {
                conn.publish(subject, inbox, data);
            } catch (IOException e) {
                throw new TTransportException("flush: unable to publish data: " + e.getMessage());
            }
        }
    }

    /**
     * NATS message handler that executes Frugal frames.
     */
    protected class Handler implements MessageHandler {
        public void onMessage(Message message) {
            try {
                executeFrame(message.getData());
            } catch (TException e) {
                LOGGER.warn("Could not execute frame", e);
            }
        }

    }

    /**
     * Convert NATS connection state to a suitable exception type.
     * @param connState nats connection state
     * @param prefix prefix to add to exception message
     *
     * @return a TTransportException type
     */
    protected static TTransportException getClosedConditionException(Constants.ConnState connState, String prefix) {
        if (connState != Constants.ConnState.CONNECTED) {
            return new TTransportException(TTransportException.NOT_OPEN,
                    String.format("%s NATS client not connected (has status %s)", prefix, connState.name()));
        }
        return new TTransportException(TTransportException.NOT_OPEN,
                String.format("%s NATS Transport not open", prefix));

    }

    private byte[] storeMessageInS3(byte[] data) {
        // Generate a random key for this message
		String s3Key = UUID.randomUUID().toString();

		// Store the message content in S3.
        System.out.println("Data: " + Arrays.toString(data));
        InputStream messageContentStream = new ByteArrayInputStream(data);
		ObjectMetadata messageContentStreamMetadata = new ObjectMetadata();
		messageContentStreamMetadata.setContentLength(data.length);
		PutObjectRequest putObjectRequest =
                new PutObjectRequest(s3BucketName, s3Key, messageContentStream, messageContentStreamMetadata);
		try {
			s3.putObject(putObjectRequest);
		} catch (AmazonServiceException e) {
			String errorMessage = "Failed to store the message content in an S3 object. NATS message was not sent.";
			LOGGER.error(errorMessage, e);
			throw new AmazonServiceException(errorMessage, e);
		} catch (AmazonClientException e) {
			String errorMessage = "Failed to store the message content in an S3 object. NATS message was not sent.";
			LOGGER.error(errorMessage, e);
			throw new AmazonClientException(errorMessage, e);
		}

		LOGGER.info("S3 object created, Bucket name: " + s3BucketName + ", Object key: " + s3Key + ".");

		// Convert S3 pointer (bucket name, key, etc) to JSON representation
		MessageS3Pointer s3Pointer = new MessageS3Pointer(s3BucketName, s3Key);
		String s3PointerStr = getJSONFromS3Pointer(s3Pointer);

		// Return S3 pointer (as bytes)
		return s3PointerStr.getBytes(StandardCharsets.UTF_8);
	}

    private String getJSONFromS3Pointer(MessageS3Pointer s3Pointer) {
		String s3PointerStr = null;
		try {
			JsonDataConverter jsonDataConverter = new JsonDataConverter();
			s3PointerStr = jsonDataConverter.serializeToJson(s3Pointer);
		} catch (Exception e) {
			String errorMessage = "Failed to convert S3 object pointer to text. Message was not sent.";
			LOGGER.error(errorMessage, e);
			throw new AmazonClientException(errorMessage, e);
		}
		return s3PointerStr;
	}
}
