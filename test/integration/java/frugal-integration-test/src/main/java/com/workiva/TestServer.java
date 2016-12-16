/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package com.workiva;

import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.server.FNatsServer;
import com.workiva.frugal.server.FServer;
import com.workiva.frugal.server.FServlet;
import com.workiva.frugal.transport.FPublisherTransportFactory;
import com.workiva.frugal.transport.FNatsPublisherTransport;
import com.workiva.frugal.transport.FNatsSubscriberTransport;
import com.workiva.frugal.transport.FSubscriberTransportFactory;
import frugal.test.Event;
import frugal.test.EventsPublisher;
import frugal.test.EventsSubscriber;
import frugal.test.FFrugalTest;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import io.nats.client.Constants;
import org.apache.catalina.Context;
import org.apache.catalina.startup.Tomcat;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TProtocolFactory;

import java.io.File;
import java.io.IOException;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;


public class TestServer {

    public static boolean middlewareCalled = false;

    public static void main(String[] args) {
        try {
            // default testing parameters, overwritten in Python runner
            int port = 9090;
            String transport_type = "stateless";
            String protocol_type = "binary";

            try {
                for (String arg : args) {
                    if (arg.startsWith("--port")) {
                        port = Integer.valueOf(arg.split("=")[1]);
                    } else if (arg.startsWith("--port")) {
                        port = Integer.parseInt(arg.split("=")[1]);
                    } else if (arg.startsWith("--protocol")) {
                        protocol_type = arg.split("=")[1];
                    } else if (arg.startsWith("--transport")) {
                        transport_type = arg.split("=")[1];
                    } else if (arg.equals("--help")) {
                        System.out.println("Allowed options:");
                        System.out.println("  --help\t\t\tProduce help message");
                        System.out.println("  --port=arg (=" + port + ")\tPort number to connect");
                        System.out.println("  --protocol=arg (=" + protocol_type + ")\tProtocol: binary, json, compact");
                        System.out.println("  --transport=arg (=" + transport_type + ")\tTransport: stateless");
                        System.exit(0);
                    }
                }
            } catch (Exception e) {
                System.err.println("Can not parse arguments! See --help");
                System.exit(1);
            }

            TProtocolFactory protocolFactory = utils.whichProtocolFactory(protocol_type);
            FProtocolFactory fProtocolFactory = new FProtocolFactory(protocolFactory);

            ConnectionFactory cf = new ConnectionFactory("nats://localhost:4222");
            Connection conn = cf.createConnection();

            List<String> validTransports = new ArrayList<>();
            validTransports.add("stateless");

            if (!validTransports.contains(transport_type)) {
                throw new Exception("Unknown transport type! " + transport_type);
            }

            // Start subscriber for pub/sub test
            new Subscriber(fProtocolFactory, port).run();

            FFrugalTest.Iface handler = new FrugalTestHandler();
            CountDownLatch called = new CountDownLatch(1);
            FFrugalTest.Processor processor = new FFrugalTest.Processor(handler, new ServerMiddleware(called));
            FServer server = null;
            Tomcat tomcat = null;
            switch (transport_type) {
                case "stateless":
                    server = new FNatsServer.Builder(
                            conn,
                            processor,
                            fProtocolFactory,
                            new String[]{Integer.toString(port)}).build();
                    break;
                case "http":
                    tomcat = new Tomcat();
                    tomcat.setPort(port);
                    String servletName = "testServlet";

                    Context ctx = tomcat.addContext("/", new File(".").getAbsolutePath());

                    Tomcat.addServlet(ctx, servletName, new FServlet(processor, fProtocolFactory));

                    ctx.addServletMappingDecoded("/*", servletName);
                    break;
        }

        // Start a healthcheck server for the cross language tests
        try {
            new HealthCheck(port);
        } catch (IOException e) {
            System.out.println(e.getMessage());
        }

        // Start server in separate thread
        if (tomcat != null) {
            RunTomcat serverThread = new RunTomcat(tomcat);
            serverThread.start();
        } else {
            RunServer serverThread = new RunServer(server, transport_type);
            serverThread.start();
        }

        // Wait for the middleware to be invoked, fail if it exceeds the longest client timeout (currently 20 sec)
        if (called.await(20, TimeUnit.SECONDS)) {
            System.out.println("Server middleware called successfully");
        } else {
            System.out.println("Server middleware not called within 20 seconds");
            System.exit(1);
        }

    } catch(
    Exception x)

    {
        x.printStackTrace();
    }
}


private static class RunTomcat extends Thread {
        Tomcat tomcat;

        RunTomcat(Tomcat tomcat) {
            this.tomcat = tomcat;
        }

        public void run() {
            System.out.println("Starting HTTP server...");
            try {
                tomcat.start();
                tomcat.getServer().await();
            } catch (Exception e) {
                System.out.printf("Exception starting server %s\n", e);
            }
        }
    }


private static class RunServer extends Thread {
    FServer server;
    String transport_type;

    RunServer(FServer server, String transport_type) {
        this.server = server;
        this.transport_type = transport_type;
    }

    public void run() {
        System.out.println("Starting " + transport_type + " server...");
        try {
            server.serve();
        } catch (Exception e) {
            System.out.printf("Exception starting server %s\n", e);
        }
    }
}


private static class ServerMiddleware implements ServiceMiddleware {
    CountDownLatch called;

    ServerMiddleware(CountDownLatch called) {
        this.called = called;
    }

    @Override
    public <T> InvocationHandler<T> apply(T next) {
        return new InvocationHandler<T>(next) {
            @Override
            public Object invoke(Method method, Object receiver, Object[] args) throws Throwable {
                Object[] subArgs = Arrays.copyOfRange(args, 1, args.length);
                System.out.printf("%s(%s)\n", method.getName(), Arrays.toString(subArgs));
                if (method.getName().equals("testOneway")) {

                    called.countDown();
                }
                return method.invoke(receiver, args);
            }
        };
    }
}


/*
Subscriber subscribes to "port-'call'" and upon receipt, publishes to "port-'response'".
The corresponding publisher in the client code publishes to "port-'call'" and subscribes
and awaits a response on "port-'response'".
*/
private static class Subscriber implements Runnable {

    FProtocolFactory protocolFactory;
    int port;

    Subscriber(FProtocolFactory protocolFactory, int port) {
        this.protocolFactory = protocolFactory;
        this.port = port;
    }

    public void run() {
        ConnectionFactory cf = new ConnectionFactory(Constants.DEFAULT_URL);
        try {
            Connection conn = cf.createConnection();
            FPublisherTransportFactory publisherFactory = new FNatsPublisherTransport.Factory(conn);
            FSubscriberTransportFactory subscriberFactory = new FNatsSubscriberTransport.Factory(conn);
            FScopeProvider provider = new FScopeProvider(publisherFactory, subscriberFactory, protocolFactory);
            EventsSubscriber.Iface subscriber = new EventsSubscriber.Client(provider);
            try {
                subscriber.subscribeEventCreated("*", "*", "call", Integer.toString(port), (context, event) -> {
                    System.out.format("received " + context + " : " + event);
                    EventsPublisher.Iface publisher = new EventsPublisher.Client(provider);
                    try {
                        publisher.open();
                        String preamble = context.getRequestHeader(utils.PREAMBLE_HEADER);
                        if (preamble == null || "".equals(preamble)) {
                            System.out.println("Client did not provide preamble header");
                            return;
                        }
                        String ramble = context.getRequestHeader(utils.RAMBLE_HEADER);
                        if (ramble == null || "".equals(ramble)) {
                            System.out.println("Client did not provide ramble header");
                            return;
                        }
                        event = new Event(1, "received call");
                        publisher.publishEventCreated(new FContext("Call"), preamble, ramble, "response", Integer.toString(port), event);

                    } catch (TException e) {
                        System.out.println("Error opening publisher to respond" + e.getMessage());
                    }
                });
            } catch (TException e) {
                System.out.println("Error subscribing" + e.getMessage());
            }
            System.out.println("Subscriber started...");

        } catch (TimeoutException | IOException e) {
            System.out.println("Error connecting to nats" + e.getMessage());
        }
    }
}

}
