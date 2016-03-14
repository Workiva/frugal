package com.workiva.frugal.server;

import com.google.gson.Gson;
import com.workiva.frugal.exception.FException;
import com.workiva.frugal.internal.NatsConnectionProtocol;
import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.processor.FProcessorFactory;
import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.protocol.FServerRegistry;
import com.workiva.frugal.transport.*;
import io.nats.client.Connection;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import io.nats.client.Subscription;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;

import java.io.IOException;
import java.io.UnsupportedEncodingException;
import java.util.Collection;
import java.util.concurrent.*;
import java.util.logging.Logger;

public class FNatsServer implements FServer {

    private static final int DEFAULT_MAX_MISSED_HEARTBEATS = 3;
    private static final String QUEUE = "rpc";

    private Connection conn;
    private String subject;
    private String heartbeatSubject;
    private final long heartbeatInterval;
    private final int maxMissedHeartbeats;
    private ConcurrentHashMap<String, Client> clients;
    private FProcessorFactory processorFactory;
    private FTransportFactory transportFactory;
    private FProtocolFactory protocolFactory;
    private final BlockingQueue<Object> shutdown = new ArrayBlockingQueue<>(1);

    private Subscription inboxSub;
    private final ScheduledExecutorService heartbeatExecutor = Executors.newScheduledThreadPool(1);

    private static Logger LOGGER = Logger.getLogger(FNatsServer.class.getName());

    public FNatsServer(Connection conn, String subject, long heartbeatInterval,
                       FProcessor processor, FTransportFactory transportFactory,
                       FProtocolFactory protocolFactory) {
        this(conn, subject, heartbeatInterval, DEFAULT_MAX_MISSED_HEARTBEATS,
                new FProcessorFactory(processor), transportFactory, protocolFactory);
    }

    public FNatsServer(Connection conn, String subject, long heartbeatInterval, int maxMissedHeartbeats,
                       FProcessorFactory processorFactory, FTransportFactory transportFactory,
                       FProtocolFactory protocolFactory) {
        this.conn = conn;
        this.subject = subject;
        this.heartbeatSubject = conn.newInbox();
        this.heartbeatInterval = heartbeatInterval;
        this.maxMissedHeartbeats = maxMissedHeartbeats;
        this.clients = new ConcurrentHashMap<>();
        this.processorFactory = processorFactory;
        this.transportFactory = transportFactory;
        this.protocolFactory = protocolFactory;
    }

    private class Client {

        TTransport transport;
        String heartbeat;
        AcceptHeartbeatThread heartbeatThread;

        Client(TTransport transport, String heartbeat) {
            this.transport = transport;
            this.heartbeat = heartbeat;
        }

        void start() {
            this.heartbeatThread = new AcceptHeartbeatThread(this);
            this.heartbeatThread.start();
        }

        void kill() {
            transport.close();
            this.heartbeatThread.kill();
            this.heartbeatThread = null;
        }

    }

    public void serve() throws TException {
        inboxSub = conn.subscribe(subject, QUEUE, new MessageHandler() {
            @Override
            public void onMessage(Message message) {
                String reply = message.getReplyTo();
                if (reply == null || reply.isEmpty()) {
                    LOGGER.warning("Received a bad connection handshake. Discarding.");
                    return;
                }

                NatsConnectionProtocol connProtocol;
                Gson gson = new Gson();
                try {
                    connProtocol = gson.fromJson(new String(message.getData(), "UTF-8"), NatsConnectionProtocol.class);
                    if (connProtocol.getVersion() != NatsConnectionProtocol.NATS_V0) {
                        LOGGER.severe(String.format("%d not a supported connect version", connProtocol.getVersion()));
                        return;
                    }
                } catch (UnsupportedEncodingException e) {
                    LOGGER.severe("could not deserialize connect message");
                    return;
                }

                String heartbeat = conn.newInbox();
                String listenTo = newFrugalInbox(message.getReplyTo());
                TTransport transport;
                try {
                    transport = accept(listenTo, reply, heartbeat);
                } catch (TException e) {
                    LOGGER.severe("error accepting client transport " + e.getMessage());
                    return;
                }

                Client client = new Client(transport, heartbeat);
                if (isHeartbeating()) {
                    client.start();
                    clients.put(heartbeat, client);
                }

                // Connect message consists of "[heartbeat subject] [heartbeat reply subject] [expected interval ms]"
                String connectMsg = heartbeatSubject + " " + heartbeat + " " + heartbeatInterval;
                try {
                    conn.publish(reply, listenTo, connectMsg.getBytes());
                } catch (Exception e) {
                    LOGGER.warning("error publishing transport inbox " + e.getMessage());
                    transport.close();
                }
            }
        });

        if (isHeartbeating()) {
            heartbeatExecutor.scheduleAtFixedRate(new MakeHeartbeatRunnable(), heartbeatInterval,
                    heartbeatInterval, TimeUnit.MILLISECONDS);
        }
        try {
            shutdown.take();
        } catch (InterruptedException ignored) {
        }
    }

    public void stop() throws TException {
        // Unsubscribing ensures no more clients will be added
        try {
            inboxSub.unsubscribe();
        } catch (IOException e) {
            throw new FException("could not unsubscribe from inbox", e);
        }

        Collection<Client> collection = clients.values();
        for (Client client : collection) {
            client.kill();
        }
        clients.clear();
        heartbeatExecutor.shutdown();

        try {
            shutdown.put(new Object());
        } catch (InterruptedException ignored) {
        }
    }

    private String newFrugalInbox(String prefix) {
        String[] tokens = prefix.split("\\.");
        tokens[tokens.length-1] = conn.newInbox(); // Always at least 1 token
        String inbox = "";
        String pre = "";
        for (String token : tokens) {
            inbox += pre + token;
            pre = ".";
        }
        return inbox;
    }

    private TTransport accept(String listenTo, String replyTo, String heartbeatSubject) throws TException {
        TTransport client = TNatsServiceTransport.server(conn, listenTo, replyTo);
        FTransport transport = transportFactory.getTransport(client);
        transport.setClosedCallback(new ClientRemover(heartbeatSubject));
        FProcessor processor = processorFactory.getProcessor(transport);
        FProtocol protocol = protocolFactory.getProtocol(transport);
        transport.setRegistry(new FServerRegistry(processor, protocolFactory, protocol));
        transport.open();
        return client;
    }

    private void remove(String heartbeat) {
        Client client = clients.get(heartbeat);
        if (client == null) {
            return;
        }
        client.kill();
        clients.remove(heartbeat);
    }

    private class MakeHeartbeatRunnable implements Runnable {
        public void run() {
            if (clients.size() == 0) {
                return;
            }
            try {
                conn.publish(heartbeatSubject, null);
            } catch (Exception e) {
                LOGGER.severe("error publishing heartbeat " + e.getMessage());
            }
        }
    }

    private class AcceptHeartbeatThread extends Thread {
        private volatile boolean running;
        private int missed;
        private final Client client;
        private SynchronousQueue<Object> heartbeatQueue = new SynchronousQueue<>();

        AcceptHeartbeatThread(Client client) {
            this.client = client;
            setName("heartbeat-accept");
        }

        public void kill() {
            if (this != Thread.currentThread()) {
                interrupt();
            }
            running = false;
        }

        public void run() {
            Subscription sub = conn.subscribe(client.heartbeat, new MessageHandler() {
                @Override
                public void onMessage(Message message) {
                    missed = 0;
                    heartbeatQueue.offer(new Object());
                }
            });

            running = true;
            while (running) {
                long wait = maxMissedHeartbeats > 1 ? heartbeatInterval : heartbeatInterval + heartbeatInterval / 4;
                try {
                    Object ret = heartbeatQueue.poll(wait, TimeUnit.MILLISECONDS);
                    if (ret == null) {
                        missed++;
                    } else {
                        missed = 0;
                    }
                } catch (InterruptedException e) {
                    continue;
                }
                if (missed >= maxMissedHeartbeats) {
                    LOGGER.info("client heartbeat expired");
                    remove(client.heartbeat);
                }
            }
            try {
                sub.unsubscribe();
            } catch (IOException e) {
                LOGGER.warning("error unsubscribing from heartbeat " + e.getMessage());
            }
        }
    }

    private boolean isHeartbeating() {
        return (heartbeatInterval > 0);
    }

    private class ClientRemover implements FTransportClosedCallback {
        private String heartbeat;

        ClientRemover(String heartbeat) {
            this.heartbeat = heartbeat;
        }

        public void onClose(Exception cause) {
            remove(this.heartbeat);
        }
    }
}
