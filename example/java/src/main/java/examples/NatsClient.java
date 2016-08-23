package examples;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FNatsTransport;
import com.workiva.frugal.transport.FTransport;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import io.nats.client.Constants;
import music.Album;
import music.FStore;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;

import java.io.IOException;
import java.util.concurrent.TimeoutException;

/**
 * Creates a stateless NATS client.
 */
public class NatsClient {
    public static void main(String[] args) throws IOException, TimeoutException, TException {
        // Specify the protocol used for serializing requests.
        // The protocol stack must match the protocol stack of the server.
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());

        // Create a NATS client (using default options for local dev)
        ConnectionFactory cf = new ConnectionFactory(ConnectionFactory.DEFAULT_URL);
        Connection conn = cf.createConnection();

        // Create and open a new transport that uses NATS for sending data.
        // The NATS transport will communicate over the music-service topic.
        FTransport transport = new FNatsTransport(conn, "music-service");
        transport.open();

        // Create a new client for the music store
        FStore.Client storeClient = new FStore.Client(transport, protocolFactory);

        // Configure the context used for sending requests
        FContext context = new FContext();
        context.setTimeout(1000);

        // Request to buy an album
        Album album = storeClient.buyAlbum(context, "ASIN-1290AIUBOA89", "ACCOUNT-12345");
        System.out.println("Bought the album: " + album);

        // Enter the contest
        storeClient.enterAlbumGiveaway(context, "kevin@workiva.com", "Kevin");

        // Close the transport
        transport.close();

        // Close the NATS client
        conn.close();
    }
}
