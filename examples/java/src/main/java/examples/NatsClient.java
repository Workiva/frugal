package examples;

import com.workiva.frugal.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.provider.FServiceProvider;
import com.workiva.frugal.transport.FNatsTransport;
import com.workiva.frugal.transport.FTransport;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import v1.music.Album;
import v1.music.FStore;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;

import java.io.IOException;
import java.util.concurrent.TimeoutException;

/**
 * Creates a NATS client.
 */
public class NatsClient {

    public static void main(String[] args) throws IOException, TimeoutException, TException {
        // Specify the protocol used for serializing requests.
        // The protocol stack must match the protocol stack of the server.
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());

        // Create a NATS client (using default options for local dev)
        ConnectionFactory cf = new ConnectionFactory(ConnectionFactory.DEFAULT_URL);
        Connection conn = cf.createConnection();

        // Create a new transport that uses NATS for sending data.
        // The NATS transport will communicate using the music-service topic.
        FTransport transport = FNatsTransport.of(conn, NatsServer.SERVICE_SUBJECT);

        // Create an FServiceProvider and open it.
        FServiceProvider provider = new FServiceProvider(transport, protocolFactory);
        provider.open();

        // Create a new client for the music store
        FStore.Client storeClient = new FStore.Client(provider);

        // Request to buy an album
        Album album = storeClient.buyAlbum(new FContext("corr-id-1"), "ASIN-1290AIUBOA89", "ACCOUNT-12345");
        System.out.println("Bought the album: " + album);

        // Enter the contest
        storeClient.enterAlbumGiveaway(new FContext("corr-id-2"), "kevin@workiva.com", "Kevin");

        // Close the provider
        provider.close();

        // Close the NATS client
        conn.close();
    }
}