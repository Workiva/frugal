import com.workiva.frugal.FContext;
import com.workiva.frugal.FProtocolFactory;
import com.workiva.frugal.FScopeProvider;
import com.workiva.frugal.FServiceProvider;
import com.workiva.frugal.server.FNatsServer;
import com.workiva.frugal.server.FServer;
import com.workiva.frugal.transport.*;
import io.nats.client.*;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.transport.TTransportException;

import java.io.IOException;
import java.util.concurrent.TimeoutException;

public class Main {

    public static void main(String[] args) throws IOException, TimeoutException, TException {
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());
        FTransportFactory transportFactory = new FMuxTransport.Factory(5);
        ConnectionFactory cf = new ConnectionFactory(Constants.DEFAULT_URL);
        Connection conn = cf.createConnection();

        if (args.length > 0) {
            runSubscriber(conn, protocolFactory);
            runServer(conn, transportFactory, protocolFactory);
        } else {
            runPublisher(conn, protocolFactory);
            runClient(conn, transportFactory, protocolFactory);
        }
    }

    private static void handleClient(FFoo.Client client) {
        FContext ctx = new FContext();
        try {
            client.ping(ctx);
        } catch (TException e) {
            System.out.println("ping error: " + e.getMessage());
        }

        Event event = new Event(42, "hello, world!");
        ctx = new FContext();
        try {
            long result = client.blah(ctx, 100, "awesomesauce", event);
            System.out.println("blah = " + result);
            System.out.println(ctx.getResponseHeader("foo"));
            System.out.println(ctx.getResponseHeaders());
        } catch (AwesomeException e) {
            System.out.println("blah error: " + e.getMessage());
        } catch (TException e) {
            e.printStackTrace();
        }
    }

    private static void runServer(Connection conn, FTransportFactory transportFactory, FProtocolFactory protocolFactory) throws TException {
        FFoo.Iface handler = new FooHandler();
        FFoo.Processor processor = new FFoo.Processor(handler);
        FServer server = new FNatsServer(conn, "foo", 60000, processor, transportFactory, protocolFactory);
        System.out.println("Starting nats server on 'foo'");
        server.serve();
    }

    private static void runClient(Connection conn, FTransportFactory transportFactory, FProtocolFactory protocolFactory) throws TTransportException, TimeoutException {
        FTransport transport = transportFactory.getTransport(TNatsServiceTransport.client(conn, "foo", 60000));
        transport.open();
        try {
            handleClient(new FFoo.Client(new FServiceProvider(transport, protocolFactory)));
        } finally {
            transport.close();
        }
    }

    private static void runSubscriber(Connection conn, FProtocolFactory protocolFactory) throws TException {
        FScopeTransportFactory factory = new FNatsScopeTransport.Factory(conn);
        FScopeProvider provider = new FScopeProvider(factory, protocolFactory);
        EventsSubscriber subscriber = new EventsSubscriber(provider);
        subscriber.subscribeEventCreated("barUser", new EventsSubscriber.EventCreatedHandler() {
            @Override
            public void onEventCreated(FContext ctx, Event req) {
                System.out.println("received " + req);
            }
        });
        System.out.println("Subscriber started...");
    }

    private static void runPublisher(Connection conn, FProtocolFactory protocolFactory) throws TException {
        FScopeTransportFactory factory = new FNatsScopeTransport.Factory(conn);
        FScopeProvider provider = new FScopeProvider(factory, protocolFactory);
        EventsPublisher publisher = new EventsPublisher(provider);
        publisher.open();
        Event event = new Event(42, "hello, world!");
        publisher.publishEventCreated(new FContext(), "barUser", event);
        System.out.println("Published event");
        publisher.close();
    }

    private static class FooHandler implements FFoo.Iface {

        @Override
        public void ping(FContext ctx) throws TException {
            System.out.format("ping(%s)\n", ctx);
        }

        @Override
        public long blah(FContext ctx, int num, String Str, Event event) throws TException, AwesomeException {
            System.out.format("blah(%s, %d, %s %s)\n", ctx, num, Str, event);
            return 42;
        }
    }

}
