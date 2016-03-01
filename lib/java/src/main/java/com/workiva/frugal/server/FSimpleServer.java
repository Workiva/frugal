package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.processor.FProcessorFactory;
import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FServerTransport;
import com.workiva.frugal.transport.FTransport;
import com.workiva.frugal.transport.FTransportFactory;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;

import java.util.logging.Logger;

/**
 * Simple single-threaded server.
 */
public class FSimpleServer implements FServer {

    private FProcessorFactory fProcessorFactory;
    private FServerTransport fServerTransport;
    private FTransportFactory fTransportFactory;
    private FProtocolFactory fProtocolFactory;
    private boolean stopped;

    private static Logger LOGGER = Logger.getLogger(FSimpleServer.class.getName());

    public FSimpleServer(FProcessorFactory fProcessorFactory, FServerTransport fServerTransport,
                         FTransportFactory fTransportFactory, FProtocolFactory fProtocolFactory) {
        this.fProcessorFactory = fProcessorFactory;
        this.fServerTransport = fServerTransport;
        this.fTransportFactory = fTransportFactory;
        this.fProtocolFactory = fProtocolFactory;
    }

    public void acceptLoop() throws TException {
        while (!stopped) {
            FTransport client;
            try {
                client = fServerTransport.accept();
            } catch (TException e) {
                if (stopped) {
                    return;
                }
                throw e;
            }
            if (client != null) {
                ProcessorThread processorThread = new ProcessorThread(client);
                processorThread.run();
            }
        }
    }

    private class ProcessorThread extends Thread {
        FTransport client;

        ProcessorThread(FTransport client) {
            this.client = client;
            setName("processor");
        }

        public void run()  {
            try {
                processRequests(client);
            } catch (TTransportException ttx) {
                // Client died, just move on
            } catch (TException tx) {
                if (!stopped) {
                    LOGGER.warning("frugal: Thrift error occurred during processing of message. " + tx.getMessage());
                }
            } catch (Exception x) {
                if (!stopped) {
                    LOGGER.warning("frugal: Error occurred during processing of message. " + x.getMessage());
                }
            }
        }
    }

    protected void processRequests(FTransport client) throws TException {
        FProcessor processor = fProcessorFactory.getProcessor(client);
        FTransport transport = fTransportFactory.getTransport(client);
        FProtocol protocol = fProtocolFactory.getProtocol(transport);
        // TODO: Set server registry here
    }

    public void serve() throws TException {
        acceptLoop();
    }

    public void stop() throws TException{
        stopped = true;
        fServerTransport.interrupt();
    }
}
