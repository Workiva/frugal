package examples;

import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.server.FServer;
import com.workiva.frugal.server.FStatelessNatsServer;
import com.workiva.frugal.transport.FNatsTransport;
import com.workiva.frugal.transport.FTransport;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import music.Album;
import music.FStore;
import music.PerfRightsOrg;
import music.PurchasingError;
import music.Track;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;

import java.io.IOException;
import java.lang.reflect.Method;
import java.util.UUID;
import java.util.concurrent.ThreadLocalRandom;
import java.util.concurrent.TimeoutException;

/**
 * Creates a NATS server listening for incoming requests.
 */
public class NatsServer {
    private static final double MIN_DURATION = 0;
    private static final double MAX_DURATION = 10000;
    public static final String SERVICE_SUBJECT = "music-service";

    public static void main(String[] args) throws IOException, TimeoutException, TException {
        // Specify the protocol used for serializing requests.
        // Clients must use the same protocol stack
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());

        // Create a NATS client (using default options for local dev)
        ConnectionFactory cf = new ConnectionFactory(ConnectionFactory.DEFAULT_URL);
        Connection conn = cf.createConnection();

        // Create and open a new transport that uses NATS for sending data.
        // The NATS transport will communicate using the music-service topic.
        FTransport transport = new FNatsTransport(conn, "music-service");
        transport.open();

        // Create a new server processor.
        // Incoming requests to the server are passed to the processor.
        // Results from the processor are returned back to the client.
        FStore.Processor processor = new FStore.Processor(new FStoreHandler(), new LoggingMiddleware());

        // Create a new music store server using the processor
        // The server can be configured using the Builder interface.
        FServer server =
                new FStatelessNatsServer.Builder(conn, processor, protocolFactory, SERVICE_SUBJECT)
                        .withQueueGroup(SERVICE_SUBJECT) // if set, all servers listen to the same queue group
                        .build();

        System.out.println("Starting nats server on " + SERVICE_SUBJECT);
        server.serve();
    }

    /**
     * A handler handles all incoming requests to the server.
     * The handler must satisfy the interface the server exposes.
     */
    private static class FStoreHandler implements FStore.Iface {

        /**
         * Return an album; always buy the same one.
         */
        @Override
        public Album buyAlbum(FContext ctx, String ASIN, String acct) throws TException, PurchasingError {
            Album album = new Album();
            album.setASIN(UUID.randomUUID().toString());
            album.setDuration(ThreadLocalRandom.current().nextDouble(MIN_DURATION, MAX_DURATION));
            album.addToTracks(
                    new Track(
                            "Coloring Book",
                            "Summer Friends",
                            "The Social Experiment",
                            "Chance the Rapper",
                            203,
                            PerfRightsOrg.ASCAP));
            return album;
        }

        @Override
        public boolean enterAlbumGiveaway(FContext ctx, String email, String name) throws TException {
            return true;
        }
    }

    private static class LoggingMiddleware implements ServiceMiddleware {

        @Override
        public <T> InvocationHandler<T> apply(T next) {
            return new InvocationHandler<T>(next) {
                @Override
                public Object invoke(Method method, Object receiver, Object[] args) throws Throwable {
                    System.out.printf("==== CALLING %s.%s ====\n", method.getDeclaringClass().getName(), method.getName());
                    Object ret = method.invoke(receiver, args);
                    System.out.printf("==== CALLED  %s.%s ====\n", method.getDeclaringClass().getName(), method.getName());
                    return ret;
                }
            };
        }

    }
}
