package com.workiva.frugal.transport;

import com.google.gson.Gson;
import com.workiva.frugal.exception.FMessageSizeException;
import com.workiva.frugal.internal.NatsConnectionProtocol;
import io.nats.client.*;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

import java.io.IOException;
import java.io.PipedInputStream;
import java.io.PipedOutputStream;
import java.nio.ByteBuffer;
import java.util.Timer;
import java.util.TimerTask;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.logging.Logger;

/**
 * TNatsServiceTransport is an extension of thrift.TTransport exclusively used for services which uses NATS as the
 * underlying transport. Message frames are limited to 1MB in size.
 */
public class TNatsServiceTransport extends TTransport {

    // NATS limits messages to 1MB.
    public static final int NATS_MAX_MESSAGE_SIZE = 1024 * 1024;
    public static final String FRUGAL_PREFIX = "frugal.";

    private static final String DISCONNECT = "DISCONNECT";
    private static final long HEARTBEAT_GRACE_PERIOD = 5 * 1000;

    private Connection conn;
    private PipedOutputStream writer;
    private PipedInputStream reader;
    private ByteBuffer writeBuffer;
    private Subscription sub;
    private String listenTo;
    private String writeTo;
    private AsyncSubscription heartbeatSub;
    private String heartbeatListen;
    private String heartbeatReply;
    private long heartbeatInterval;
    private Timer heartbeatTimer;
    private AtomicInteger missedHeartbeats;
    private String connectionSubject;
    private final long connectionTimeout;
    private final int maxMissedHeartbeats;
    private boolean isOpen;

    private static Logger LOGGER = Logger.getLogger(TNatsServiceTransport.class.getName());

    /**
     * Used for constructing server side of TNatsServiceTransport
     */
    private TNatsServiceTransport(Connection conn, String listenTo, String writeTo) {
        this.conn = conn;
        this.listenTo = listenTo;
        this.writeTo = writeTo;
        this.missedHeartbeats = new AtomicInteger(0);
        // Neither of these are needed for the server side of the transport.
        this.connectionTimeout = 0;
        this.maxMissedHeartbeats = 0;
    }

    /**
     * Used for constructing client side of TNatsServiceTransport
     */
    private TNatsServiceTransport(Connection conn, String connectionSubject, long connectionTimeout, int maxMissedHeartbeats) {
        this.conn = conn;
        this.connectionSubject = connectionSubject;
        this.connectionTimeout = connectionTimeout;
        this.maxMissedHeartbeats = maxMissedHeartbeats;
        this.missedHeartbeats = new AtomicInteger(0);
    }

    /**
     * Returns a new thrift TTransport which uses the NATS messaging system as the
     * underlying transport. It performs a handshake with a server listening on the
     * given NATS subject upon open. This TTransport can only be used with
     * FNatsServer.
     */
    public static TNatsServiceTransport client(Connection conn, String subject, long timeout, int maxMissedHeartbeats) {
        return new TNatsServiceTransport(conn, subject, timeout, maxMissedHeartbeats);
    }

    /**
     * Returns a new thrift TTransport which uses the NATS messaging system as the
     * underlying transport. This TTransport can only be used with FNatsServer.
     */
    public static TNatsServiceTransport server(Connection conn, String listenTo, String writeTo) {
        return new TNatsServiceTransport(conn, listenTo, writeTo);
    }

    @Override
    public synchronized boolean isOpen() {
        return conn.getState() == Constants.ConnState.CONNECTED && isOpen;
    }

    /**
     * Opens the transport for reading/writing.
     * Performs a handshake with the server if this is a client transport.
     *
     * @throws TTransportException if the transport could not be opened
     */
    @Override
    public synchronized void open() throws TTransportException {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            throw new TTransportException(TTransportException.NOT_OPEN,
                    "NATS not connected, has status " + conn.getState());
        }
        if (isOpen) {
            throw new TTransportException(TTransportException.ALREADY_OPEN, "NATS transport already open");
        }

        if (connectionSubject != null) {
            handshake();
        }

        if (listenTo == null || "".equals(listenTo) || writeTo == null || "".equals(writeTo)) {
            throw new TTransportException("listenTo and writeTo cannot be empty.");
        }

        writeBuffer = ByteBuffer.allocate(NATS_MAX_MESSAGE_SIZE);

        try {
            writer = new PipedOutputStream();
            reader = new PipedInputStream(writer);
        } catch (IOException e) {
            throw new TTransportException(e);
        }

        sub = conn.subscribe(listenTo, new MessageHandler() {
            @Override
            public void onMessage(Message msg) {
                if (DISCONNECT.equals(msg.getReplyTo())) {
                    close();
                    return;
                }
                try {
                    writer.write(msg.getData());
                    writer.flush();
                } catch (IOException e) {
                    LOGGER.warning("could not write incoming data to buffer" + e.getMessage());
                }
            }
        });

        if (heartbeatInterval > 0) {
            startTimer();
            heartbeatSub = conn.subscribe(heartbeatListen, new MessageHandler() {

                @Override
                public void onMessage(Message message) {
                    receiveHeartbeat();
                    conn.publish(heartbeatReply, null);
                }
            });
        }
        isOpen = true;
    }

    private void handshake() throws TTransportException {
        NatsConnectionProtocol connectionProtocol = new NatsConnectionProtocol(NatsConnectionProtocol.NATS_V0);
        Gson gson = new Gson();
        String serializedVersion = gson.toJson(connectionProtocol);
        Message message;
        try {
            message = handshakeRequest(serializedVersion.getBytes("UTF-8"));
        } catch (IOException e) {
            throw new TTransportException(e);
        } catch (TimeoutException e) {
            throw new TTransportException(TTransportException.TIMED_OUT, "Handshake timed out", e);
        }
        String reply = message.getReplyTo();
        if (reply == null || reply.isEmpty()) {
            throw new TTransportException("No reply subject on connect.");
        }

        String[] subjects = new String(message.getData()).split(" ");
        if (subjects.length != 3) {
            throw new TTransportException("Invalid connect message.");
        }

        String heartbeatListen = subjects[0];
        String heartbeatReply = subjects[1];
        int deadline;
        try {
            deadline = Integer.parseInt(subjects[2]);
        } catch (NumberFormatException e) {
            throw new TTransportException("Connection deadline not an integer.", e);
        }

        long heartbeatInterval = 0;
        if (deadline > 0) {
            heartbeatInterval = deadline;
        }

        this.heartbeatListen = heartbeatListen;
        this.heartbeatReply = heartbeatReply;
        this.heartbeatInterval = heartbeatInterval;
        this.listenTo = message.getSubject();
        this.writeTo = reply;
    }

    private Message handshakeRequest(byte[] handshakeBytes) throws TimeoutException, IOException {
        String inbox = newFrugalInbox();
        try (SyncSubscription s = conn.subscribeSync(inbox, null)) {
            s.autoUnsubscribe(1);
            conn.publish(this.connectionSubject, inbox, handshakeBytes);
            return s.nextMessage(this.connectionTimeout, TimeUnit.MILLISECONDS);
        }
    }


    private String newFrugalInbox() {
        return TNatsServiceTransport.FRUGAL_PREFIX + conn.newInbox();
    }

    private void startTimer() {
        heartbeatTimer = new Timer();
        heartbeatTimer.schedule(new TimerTask() {
            @Override
            public void run() {
                missedHeartbeat();
            }
        }, heartbeatTimeoutPeriod());
    }

    private void missedHeartbeat() {
        int missed = missedHeartbeats.getAndIncrement();
        if (missed >= maxMissedHeartbeats) {
            LOGGER.warning("missed " + missed + " heartbeats from peer, closing transport");
            close();
            return;
        }
        startTimer();
    }

    private void receiveHeartbeat() {
        heartbeatTimer.cancel();
        missedHeartbeats.set(0);
        startTimer();
    }

    @Override
    public synchronized void close() {
        if (!isOpen) {
            return;
        }
        // Signal remote peer for a graceful disconnect.
        conn.publish(writeTo, DISCONNECT, null);

        if (heartbeatSub != null) {
            try {
                heartbeatSub.unsubscribe();
            } catch (IOException e) {
                LOGGER.warning("close: could not unsubscribe from heartbeat subscription. " + e.getMessage());
            }
            heartbeatSub = null;
        }
        if (heartbeatTimer != null) {
            heartbeatTimer.cancel();
            heartbeatTimer = null;
        }

        try {
            sub.unsubscribe();
        } catch (IOException e) {
            LOGGER.warning("close: could not unsubscribe from inbox subscription. " + e.getMessage());
        }
        sub = null;

        // Flush the NATS connection to avoid an edge case where the program exits after closing the transport. This is
        // because NATS asynchronously flushes in the background, so explicitly flushing prevents us from losing
        // anything buffered when we exit.
        try {
            conn.flush();
        } catch (Exception e) {
            LOGGER.warning("close: could not flush NATS connection. " + e.getMessage());
        }

        try {
            writer.close();
        } catch (IOException e) {
            LOGGER.warning("close: could not close write buffer. " + e.getMessage());
        }
        isOpen = false;
    }

    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw getClosedConditionException(conn, "read:");
        }
        try {
            int bytesRead = this.reader.read(bytes, off, len);
            if (bytesRead < 0) {
                throw new TTransportException(TTransportException.END_OF_FILE);
            }
            return bytesRead;
        } catch (IOException e) {
            throw new TTransportException(TTransportException.END_OF_FILE, e);
        }
    }

    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw getClosedConditionException(conn, "write:");
        }
        if (writeBuffer.remaining() < len) {
            writeBuffer.clear();
            throw new FMessageSizeException(
                    String.format("Message exceeds %d bytes, was %d bytes",
                            TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE,
                            len + TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE - writeBuffer.remaining()));
        }
        writeBuffer.put(bytes, off, len);
    }

    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw getClosedConditionException(conn, "flush:");
        }
        byte[] data = new byte[writeBuffer.position()];
        writeBuffer.flip();
        writeBuffer.get(data);
        if (data.length == 0) {
            return;
        }
        if (data.length > TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE) {
            throw new FMessageSizeException(String.format(
                    "Message exceeds %d bytes, was %d bytes",
                    TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE, data.length));
        }
        conn.publish(writeTo, data);
        writeBuffer.clear();
    }

    private long heartbeatTimeoutPeriod() {
        // The server is expected to heartbeat at every heartbeatInterval. Add an additional grace period.
        return heartbeatInterval + HEARTBEAT_GRACE_PERIOD;
    }

    static TTransportException getClosedConditionException(Connection conn, String prefix) {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            return new TTransportException(TTransportException.NOT_OPEN,
                    String.format("%s NATS client not connected (has status %s)", prefix, conn.getState().name()));
        }
        return new TTransportException(TTransportException.NOT_OPEN,
                String.format("%s NATS FScopeTransport not open", prefix));
    }
}
