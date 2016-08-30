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

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.server.FNatsServer;
import com.workiva.frugal.server.FServer;
import com.workiva.frugal.server.FStatelessNatsServer;
import com.workiva.frugal.transport.FMuxTransport;
import com.workiva.frugal.transport.FNatsScopeTransport;
import com.workiva.frugal.transport.FScopeTransportFactory;
import com.workiva.frugal.transport.FTransportFactory;
import frugal.test.*;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import io.nats.client.Constants;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TProtocolFactory;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.*;
import java.util.concurrent.TimeoutException;


public class TestServer {

    public static void main(String [] args) {
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
                        System.out.println("  --transport=arg (=" + transport_type + ")\tTransport: stateless, stateful, stateless-stateful");
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
            validTransports.add("stateful");
            validTransports.add("stateless-stateful");

            if (!validTransports.contains(transport_type)) {
                throw new Exception("Unknown transport type! " + transport_type);
            }

            // Start pub/sub in a separate thread
            new Subscriber(fProtocolFactory, port).run();

            FFrugalTest.Iface handler = new TestServerHandler();
            FFrugalTest.Processor processor = new FFrugalTest.Processor(handler);
            FServer server = null;
            switch (transport_type) {
                case "stateless":
                    server = new FStatelessNatsServer.Builder(
                            conn,
                            processor,
                            fProtocolFactory,
                            Integer.toString(port)).build();
                    break;
                case "stateful":
                case "stateless-stateful":
                    FTransportFactory fTransportFactory = new FMuxTransport.Factory(2);
                    server = new FNatsServer(
                            conn,
                            Integer.toString(port),
                            10000,
                            processor,
                            fTransportFactory,
                            fProtocolFactory);
                    break;
            }

            // Start a healthcheck server for the cross language tests
            try {
                HealthCheck healthcheck = new HealthCheck(port);
            } catch (IOException e) {
                System.out.println(e.getMessage());
            }

            System.out.println("Starting " + transport_type + " server...");
            server.serve();

        } catch (Exception x) {
            x.printStackTrace();
        }
    }

    private static class TestServerHandler implements FFrugalTest.Iface {

//      Each RPC handler "test___" accepts a value of type ___ and returns the same value (where applicable).
//      The client then asserts that the returned value is equal to the value sent.
        @Override
        public void testVoid(FContext ctx) throws TException {
            System.out.format("testVoid(%s)\n", ctx);
        }

        @Override
        public String testString(FContext ctx, String thing) throws TException {
            System.out.format("testString(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public boolean testBool(FContext ctx, boolean thing) throws TException {
            System.out.format("testBool(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public byte testByte(FContext ctx, byte thing) throws TException {
            System.out.format("testByte(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public int testI32(FContext ctx, int thing) throws TException {
            System.out.format("testI32(%s, %d)\n", ctx, thing);
            return thing;
        }

        @Override
        public long testI64(FContext ctx, long thing) throws TException {
            System.out.format("testI64(%s, %d)\n", ctx, thing);
            return thing;
        }

        @Override
        public double testDouble(FContext ctx, double thing) throws TException {
            System.out.format("testDouble(%s, %f)\n", ctx, thing);
            return thing;
        }

        @Override
        public ByteBuffer testBinary(FContext ctx, ByteBuffer thing) throws TException {
            System.out.format("testBinary(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public Xtruct testStruct(FContext ctx, Xtruct thing) throws TException {
            System.out.format("testStruct(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public Xtruct2 testNest(FContext ctx, Xtruct2 thing) throws TException {
            System.out.format("testNest(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public Map<Integer, Integer> testMap(FContext ctx, Map<Integer, Integer> thing) throws TException {
            System.out.format("testMap(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public Map<String, String> testStringMap(FContext ctx, Map<String, String> thing) throws TException {
            System.out.format("testStringMap(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public Set<Integer> testSet(FContext ctx, Set<Integer> thing) throws TException {
            System.out.format("testSet(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public List<Integer> testList(FContext ctx, List<Integer> thing) throws TException {
            System.out.format("testList(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public Numberz testEnum(FContext ctx, Numberz thing) throws TException {
            System.out.format("testEnum(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public long testTypedef(FContext ctx, long thing) throws TException {
            System.out.format("testTypedef(%s, %s)\n", ctx, thing);
            return thing;
        }

        @Override
        public Map<Integer, Map<Integer, Integer>> testMapMap(FContext ctx, int hello) throws TException {
            System.out.format("testMapMap(%s, %d)\n", ctx, hello);

            Map<Integer, Integer> mp1 = new HashMap<>();
            mp1.put(-4,-4);
            mp1.put(-3,-3);
            mp1.put(-2,-2);
            mp1.put(-1,-1);

            Map<Integer, Integer> mp2 = new HashMap<>();
            mp2.put(4,4);
            mp2.put(3,3);
            mp2.put(2,2);
            mp2.put(1,1);

            Map<Integer, Map<Integer, Integer>> rMapMap = new HashMap<>();
            rMapMap.put(-4, mp1);
            rMapMap.put(4, mp2);
            return rMapMap;
        }

        @Override
        public Map<Long, Map<Numberz, Insanity>> testInsanity(FContext ctx, Insanity argument) throws TException {
            System.out.format("testInsanity(%s, %s)\n", ctx, argument);

            Map<Numberz, Insanity> mp1 = new HashMap<>();
            mp1.put(Numberz.findByValue(2), argument);
            mp1.put(Numberz.findByValue(3), argument);

            Map<Numberz, Insanity> mp2 = new HashMap<>();

            Map<Long, Map<Numberz, Insanity>> returnInsanity = new HashMap<>();
            returnInsanity.put((long) 1, mp1);
            returnInsanity.put((long) 2, mp2);

            return returnInsanity;
        }

        @Override
        public Xtruct testMulti(FContext ctx, byte arg0, int arg1, long arg2, Map<Short, String> arg3, Numberz arg4, long arg5) throws TException {
            System.out.format("testMulti(%s, %s, %s, %s, %s, %s\n", ctx, arg1, arg2, arg3, arg4, arg5);
            Xtruct r = new Xtruct();

            r.string_thing = "Hello2";
            r.byte_thing = arg0;
            r.i32_thing = arg1;
            r.i64_thing = arg2;

            return r;
        }

        @Override
        public void testException(FContext ctx, String arg) throws TException {
            System.out.format("testException(%s, %s)\n", ctx, arg);
            switch (arg) {
                case "Xception":
                    Xception e = new Xception();
                    e.errorCode = 1001;
                    e.message = arg;
                    throw e;
                case "TException":
                    throw new TException("Just TException");
                default:
            }
        }

        @Override
        public Xtruct testMultiException(FContext ctx, String arg0, String arg1) throws TException {
            System.out.format("testMultiException(%s, %s, %s)\n", ctx, arg0, arg1);
            switch (arg0) {
                case "Xception":
                    Xception e = new Xception();
                    e.errorCode = 1001;
                    e.message = "This is an Xception";
                    throw e;
                case "Xception2":
                    Xception2 e2 = new Xception2();
                    e2.errorCode = 2002;
                    e2.struct_thing = new Xtruct();
                    e2.struct_thing.string_thing = "This is an Xception2";
                    throw e2;
                default:
                    Xtruct r = new Xtruct();
                    r.string_thing = arg1;
                    return r;
            }
        }

        @Override
        public void testOneway(FContext ctx, int secondsToSleep) throws TException {
            System.out.format("testOneway(%s, %d)\n", ctx, secondsToSleep);
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
                FScopeTransportFactory factory = new FNatsScopeTransport.Factory(conn);
                FScopeProvider provider = new FScopeProvider(factory, protocolFactory);
                EventsSubscriber subscriber = new EventsSubscriber(provider);
                try {
                    subscriber.subscribeEventCreated(Integer.toString(port)+"-call", (context, event) -> {
                        System.out.format("received " + context + " : " + event);
                        EventsPublisher publisher = new EventsPublisher(provider);
                        try {
                            publisher.open();
                            event = new Event(1, "received call");
                            publisher.publishEventCreated(new FContext("Call"), Integer.toString(port)+"-response", event);

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
