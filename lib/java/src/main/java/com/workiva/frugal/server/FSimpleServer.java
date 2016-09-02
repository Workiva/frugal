package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.protocol.FServerRegistry;
import com.workiva.frugal.transport.FTransport;
import com.workiva.frugal.transport.FTransportFactory;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TServerTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.concurrent.atomic.AtomicBoolean;

/**
 * Simple multi-threaded server.
 */
public class FSimpleServer implements FServer {

    private static final Logger LOGGER = LoggerFactory.getLogger(FSimpleServer.class);

    private final FProcessor fProcessor;
    private final TServerTransport tServerTransport;
    private final FTransportFactory fTransportFactory;
    private final FProtocolFactory fProtocolFactory;
    private final AtomicBoolean isStopped = new AtomicBoolean(false);

    private FSimpleServer(FProcessor fProcessor, TServerTransport fServerTransport,
                          FTransportFactory fTransportFactory, FProtocolFactory fProtocolFactory) {
        this.fProcessor = fProcessor;
        this.tServerTransport = fServerTransport;
        this.fTransportFactory = fTransportFactory;
        this.fProtocolFactory = fProtocolFactory;
    }

    /**
     * Create a new FSimpleServer.
     *
     * @param fProcessor Processor for incoming requests.
     * @param fServerTransport Server transport listening to requests.
     * @param fTransportFactory Factory for creating transports
     * @param fProtocolFactory Protocol factory
     *
     * @return FSimpleServer
     */
    public static FSimpleServer of(FProcessor fProcessor, TServerTransport fServerTransport,
                                   FTransportFactory fTransportFactory, FProtocolFactory fProtocolFactory) {
        return new FSimpleServer(fProcessor, fServerTransport, fTransportFactory, fProtocolFactory);
    }

    /**
     * Starts the server by listening on the server transport and starting
     * an accept loop.
     *
     * @throws TException
     */
    @Override
    public void serve() throws TException {
        tServerTransport.listen();
        acceptLoop();
    }

    /**
     * Stops the server by interrupting the server transport.
     *
     * @throws TException
     */
    @Override
    public void stop() throws TException {
        isStopped.set(true);
        tServerTransport.interrupt();
    }

    /**
     * Loop while accepting incoming data on the configured transport.
     */
    void acceptLoop() throws TException {
        while (!isStopped.get()) {
            TTransport client;
            try {
                client = tServerTransport.accept();
            } catch (TException e) {
                if (isStopped.get()) {
                    return;
                }
                throw e;
            }
            if (client != null) {
                try {
                    accept(client);
                } catch (TException e) {
                    LOGGER.warn("frugal: error accepting client connection: " + e.getMessage());
                }
            }
        }
    }

    /**
     * Open the transport and set the server callback registry.
     */
    void accept(TTransport client) throws TException {
        FTransport transport = fTransportFactory.getTransport(client);
        FProtocol protocol = fProtocolFactory.getProtocol(transport);
        transport.setRegistry(new FServerRegistry(fProcessor, fProtocolFactory, protocol));
        transport.open();
    }
}
