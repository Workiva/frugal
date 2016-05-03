
import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.processor.FProcessorFactory;
import com.workiva.frugal.protocol.*;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.server.FNatsServer;
import com.workiva.frugal.server.FServer;
import com.workiva.frugal.transport.*;
import example.*;
import io.nats.client.*;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.transport.TTransportException;

import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.Map;
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
        try {
            client.ping(new FContext());
            System.out.println("ping()");
        } catch (TException e) {
            System.out.println("ping error: " + e.getMessage());
        }

        try {
            client.basePing(new FContext());
            System.out.println("basePing()");
        } catch (TException e) {
            System.out.println("basePing error: " + e.getMessage());
        }

        Event event = new Event(42, "hello, world!");
        FContext ctx = new FContext();
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
        FFoo.Processor processor = new FFoo.Processor(handler, new LoggingMiddleware());
        FServer server = new FNatsServer(conn, "foo", 20 * 1000, 2, new FProcessorFactory(processor), transportFactory, protocolFactory);
        System.out.println("Starting nats server on 'foo'");
        server.serve();
    }

    private static void runClient(Connection conn, FTransportFactory transportFactory, FProtocolFactory protocolFactory) throws TTransportException {
        FTransport transport = transportFactory.getTransport(TNatsServiceTransport.client(conn, "foo", 5000, 2));
        transport.open();
        try {
            handleClient(new FFoo.Client(transport, protocolFactory, new RetryMiddleware()));
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
            ctx.addResponseHeader("foo", "bar");
            return 42;
        }

        @Override
        public void oneWay(FContext ctx, long id, Map<Integer, String> req) throws TException {
            System.out.format("oneWay(%s, %d, %s)\n", ctx, id, req);
        }

        @Override
        public void basePing(FContext ctx) throws TException {
            System.out.format("basePing(%s)\n", ctx);
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

    private static class RetryMiddleware implements ServiceMiddleware {

        @Override
        public <T> InvocationHandler<T> apply(T next) {
            return new InvocationHandler<T>(next) {
                @Override
                public Object invoke(Method method, T receiver, Object[] args) throws Throwable {
                    Throwable ex = null;
                    for (int i = 0; i < 5; i++) {
                        try {
                            return method.invoke(receiver, args);
                        } catch (InvocationTargetException e) {
                            ex = e.getCause();
                            System.out.printf("%s.%s failed (%s), retrying...\n", method.getDeclaringClass().getName(),
                                    method.getName(), e.getCause());
                            Thread.sleep(500);
                        }
                    }
                    throw ex;
                }
            };
        }

    }

}
