package examples;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.transport.FNatsScopeTransport;
import com.workiva.frugal.transport.FScopeTransportFactory;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;
import v1.music.Album;
import v1.music.AlbumWinnersSubscriber;

import java.io.IOException;
import java.util.concurrent.TimeoutException;

/**
 * Create a NATS PubSub subscriber.
 */
public class NatsSubscriber {

    public static void main(String[] args) throws TException, IOException, TimeoutException {
        // Specify the protocol used for serializing requests.
        // The protocol stack must match the protocol stack of the publisher.
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());

        // Create a NATS client (using default options for local dev)
        ConnectionFactory cf = new ConnectionFactory(ConnectionFactory.DEFAULT_URL);
        Connection conn = cf.createConnection();

        // Create the pubsub scope transport and provider, given the NATs connection and protocol
        FScopeTransportFactory factory = new FNatsScopeTransport.Factory(conn);
        FScopeProvider provider = new FScopeProvider(factory, protocolFactory);

        // Subscribe to winner announcements
        AlbumWinnersSubscriber subscriber = new AlbumWinnersSubscriber(provider);
        subscriber.subscribeWinner(new AlbumWinnersSubscriber.WinnerHandler() {
            @Override
            public void onWinner(FContext ctx, Album album) {
                System.out.println("You won! " + album);
            }
        });
        System.out.println("Subscriber started...");
    }
}
