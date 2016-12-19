package examples;

import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.server.FNatsServer;
import com.workiva.frugal.server.FServer;
import com.workiva.frugal.transport.FNatsTransport;
import com.workiva.frugal.transport.FTransport;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import v1.music.FStore;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;

import java.io.IOException;
import java.util.concurrent.TimeoutException;

/**
 * Creates a NATS server listening for incoming requests.
 */
public class NatsServer {
    public static final String SERVICE_SUBJECT = "music-service";

    public static void main(String[] args) throws IOException, TimeoutException, TException {
        // Specify the protocol used for serializing requests.
        // Clients must use the same protocol stack
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());

        // Create a NATS client (using default options for local dev)
        ConnectionFactory cf = new ConnectionFactory(ConnectionFactory.DEFAULT_URL);
        Connection conn = cf.createConnection();

        // Create a new server processor.
        // Incoming requests to the server are passed to the processor.
        // Results from the processor are returned back to the client.
        FStore.Processor processor = new FStore.Processor(new FStoreHandler(), new LoggingMiddleware());

        // Create a new music store server using the processor
        // The server can be configured using the Builder interface.
        FServer server =
                new FNatsServer.Builder(conn, processor, protocolFactory, new String[]{SERVICE_SUBJECT})
                        .withQueueGroup(SERVICE_SUBJECT) // if set, all servers listen to the same queue group
                        .build();

        System.out.println("Starting nats server on " + SERVICE_SUBJECT);
        server.serve();
    }

}