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
import com.workiva.frugal.transport.FHttpTransport;
import com.workiva.frugal.transport.FTransport;
import com.workiva.frugal.transport.FTransportFactory;
import com.workiva.frugal.transport.TNatsServiceTransport;
import com.workiva.frugal.transport.FNatsTransport;
import com.workiva.frugal.transport.FMuxTransport;
import com.workiva.frugal.transport.FScopeTransportFactory;
import com.workiva.frugal.transport.FNatsScopeTransport;
import com.workiva.utils;
import frugal.test.*;
import frugal.test.Numberz;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransport;

import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.util.*;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.TimeUnit;

/**
 * Test Java client for frugal. This makes a variety of requests to enable testing for both performance and
 * correctness of the output.
 */
public class TestClient {

    public static boolean middlewareCalled = false;

    public static void main(String[] args) throws Exception {
        // default testing parameters, overwritten in Python runner
        String host = "localhost";
        int port = 9090;
        String protocol_type = "binary";
        String transport_type = "stateless";

        int socketTimeoutMs = 1000; // milliseconds
        ConnectionFactory cf = new ConnectionFactory("nats://localhost:4222");
        Connection conn = cf.createConnection();

        try {
            for (String arg : args) {
                if (arg.startsWith("--host")) {
                    host = arg.split("=")[1];
                } else if (arg.startsWith("--port")) {
                    port = Integer.valueOf(arg.split("=")[1]);
                } else if (arg.startsWith("--protocol")) {
                    protocol_type = arg.split("=")[1];
                } else if (arg.startsWith("--transport")) {
                    transport_type = arg.split("=")[1];
                } else if (arg.equals("--help")) {
                    System.out.println("Allowed options:");
                    System.out.println("  --help\t\t\tProduce help message");
                    System.out.println("  --host=arg (=" + host + ")\tHost to connect");
                    System.out.println("  --port=arg (=" + port + ")\tPort number to connect");
                    System.out.println("  --transport=arg (=" + transport_type + ")\n\t\t\t\tTransport: stateless, stateful, stateless-stateful, http");
                    System.out.println("  --protocol=arg (=" + protocol_type + ")\tProtocol: binary, json, compact");
                    System.exit(0);
                }
            }
        } catch (Exception x) {
            System.err.println("Can not parse arguments! See --help");
            System.err.println("Exception parsing arguments: " + x);
            System.exit(1);
        }
        TProtocolFactory protocolFactory = utils.whichProtocolFactory(protocol_type);

        List<String> validTransports = new ArrayList<>();
        validTransports.add("stateless");
        validTransports.add("stateful");
        validTransports.add("stateless-stateful");
        validTransports.add("http");

        if (!validTransports.contains(transport_type)) {
            throw new Exception("Unknown transport type! " + transport_type);
        }

        FTransport fTransport = null;

        try {
            switch (transport_type) {
                case "http":
                    String url = "http://" + host + ":" + port;
                    CloseableHttpClient httpClient = HttpClients.createDefault();
                    FHttpTransport.Builder httpTransport = new FHttpTransport.Builder(httpClient, url);
                    fTransport = httpTransport.build();
                    fTransport.open();
                    break;
                case "stateful":
                    TTransport tTransport = TNatsServiceTransport.client(conn, ""+ port, socketTimeoutMs, 3);
                    FTransportFactory fTransportFactory = new FMuxTransport.Factory(2);
                    fTransport = fTransportFactory.getTransport(tTransport);
                    break;
                case "stateless":
                case "stateless-stateful":
                    fTransport = new FNatsTransport(conn, Integer.toString(port));
                    break;
            }
        } catch (Exception x) {
            x.printStackTrace();
            System.exit(1);
        }

        try {
            fTransport.open();
        } catch (Exception e) {
            e.printStackTrace();
            System.out.println("Failed to open fTransport: " + e.getMessage());
            System.exit(1);
        }

        FFrugalTest.Client testClient = new FFrugalTest.Client(fTransport, new FProtocolFactory(protocolFactory), new ClientMiddleware());

        Insanity insane = new Insanity();
        FContext context = new FContext("");

        int returnCode = 0;
        try {
            /**
             * VOID TEST
             */

            try {
                testClient.testVoid(context);
            } catch (TApplicationException tax) {
                tax.printStackTrace();
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * STRING TEST
             */
            String s = testClient.testString(context, "Test");
            if (!s.equals("Test")) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * BYTE TEST
             */
            byte i8 = testClient.testByte(context, (byte) 1);
            if (i8 != 1) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * I32 TEST
             */
            int i32 = testClient.testI32(context, -1);
            if (i32 != -1) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * I64 TEST
             */
            long i64 = testClient.testI64(context, -34359738368L);
            if (i64 != -34359738368L) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * DOUBLE TEST
             */
            double dub = testClient.testDouble(context, -5.325098235);
            if (Math.abs(dub - (-5.325098235)) > 0.001) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * BINARY TEST
             */
            try {
                // There currently a mismatch between java and go that will cause test failures if the length of this array is not divisible by 4
                // TODO: Use commented line in lieu of modified line once a fix is in Thrift
                // byte[] data = new byte[]{-128, -127, -126, -125, -124, -123, -122, -121, -120, -119, -118, -117, -116, -115, -114, -113, -112, -111, -110, -109, -108, -107, -106, -105, -104, -103, -102, -101, -100, -99, -98, -97, -96, -95, -94, -93, -92, -91, -90, -89, -88, -87, -86, -85, -84, -83, -82, -81, -80, -79, -78, -77, -76, -75, -74, -73, -72, -71, -70, -69, -68, -67, -66, -65, -64, -63, -62, -61, -60, -59, -58, -57, -56, -55, -54, -53, -52, -51, -50, -49, -48, -47, -46, -45, -44, -43, -42, -41, -40, -39, -38, -37, -36, -35, -34, -33, -32, -31, -30, -29, -28, -27, -26, -25, -24, -23, -22, -21, -20, -19, -18, -17, -16, -15, -14, -13, -12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127};
                byte[] data = new byte[]{-127, -126, -125, -124, -123, -122, -121, -120, -119, -118, -117, -116, -115, -114, -113, -112, -111, -110, -109, -108, -107, -106, -105, -104, -103, -102, -101, -100, -99, -98, -97, -96, -95, -94, -93, -92, -91, -90, -89, -88, -87, -86, -85, -84, -83, -82, -81, -80, -79, -78, -77, -76, -75, -74, -73, -72, -71, -70, -69, -68, -67, -66, -65, -64, -63, -62, -61, -60, -59, -58, -57, -56, -55, -54, -53, -52, -51, -50, -49, -48, -47, -46, -45, -44, -43, -42, -41, -40, -39, -38, -37, -36, -35, -34, -33, -32, -31, -30, -29, -28, -27, -26, -25, -24, -23, -22, -21, -20, -19, -18, -17, -16, -15, -14, -13, -12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127};
                ByteBuffer bin = testClient.testBinary(context, ByteBuffer.wrap(data));

                bin.mark();
                byte[] bytes = new byte[bin.limit() - bin.position()];
                bin.get(bytes);
                bin.reset();

                if (!ByteBuffer.wrap(data).equals(bin)) {
                    returnCode |= 1;
                    System.out.println("*** FAILURE ***\n");
                }
            } catch (Exception ex) {
                returnCode |= 1;
                System.out.println("\n*** FAILURE ***\n");
                ex.printStackTrace(System.out);
            }


            /**
             * STRUCT TEST
             */
            Xtruct out = new Xtruct();
            out.string_thing = "Zero";
            out.byte_thing = (byte) 1;
            out.i32_thing = -3;
            out.i64_thing = -5;
            Xtruct in = testClient.testStruct(context, out);

            if (!in.equals(out)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * NESTED STRUCT TEST
             */
            Xtruct2 out2 = new Xtruct2();
            out2.byte_thing = (short) 1;
            out2.struct_thing = out;
            out2.i32_thing = 5;
            Xtruct2 in2 = testClient.testNest(context, out2);
            in = in2.struct_thing;

            if (!in2.equals(out2)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * MAP TEST
             */
            Map<Integer, Integer> mapout = new HashMap<>();
            for (int i = 0; i < 5; ++i) {
                mapout.put(i, i - 10);
            }
            Map<Integer, Integer> mapin = testClient.testMap(context, mapout);

            if (!mapout.equals(mapin)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * STRING MAP TEST
             */
            try {
                Map<String, String> smapout = new HashMap<>();
                smapout.put("a", "2");
                smapout.put("b", "blah");
                smapout.put("some", "thing");
                Map<String, String> smapin = testClient.testStringMap(context, smapout);
                if (!smapout.equals(smapin)) {
                    returnCode |= 1;
                    System.out.println("*** FAILURE ***\n");
                }
            } catch (Exception ex) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
                ex.printStackTrace(System.out);
            }

            /**
             * SET TEST
             */
            Set<Integer> setout = new HashSet<>();
            for (int i = -2; i < 3; ++i) {
                setout.add(i);
            }
            Set<Integer> setin = testClient.testSet(context, setout);
            if (!setout.equals(setin)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * LIST TEST
             */
            List<Integer> listout = new ArrayList<>();
            for (int i = -2; i < 3; ++i) {
                listout.add(i);
            }
            List<Integer> listin = testClient.testList(context, listout);
            if (!listout.equals(listin)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * ENUM TEST
             */
            Numberz ret = testClient.testEnum(context, Numberz.ONE);
            if (ret != Numberz.ONE) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            ret = testClient.testEnum(context, Numberz.TWO);
            if (ret != Numberz.TWO) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            ret = testClient.testEnum(context, Numberz.THREE);
            if (ret != Numberz.THREE) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            ret = testClient.testEnum(context, Numberz.FIVE);
            if (ret != Numberz.FIVE) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            ret = testClient.testEnum(context, Numberz.EIGHT);
            if (ret != Numberz.EIGHT) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * TYPEDEF TEST
             */
            long uid = testClient.testTypedef(context, 309858235082523L);
            if (uid != 309858235082523L) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * NESTED MAP TEST
             */
            Map<Integer, Map<Integer, Integer>> mm =
                    testClient.testMapMap(context, 1);
            if (mm.size() != 2 || !mm.containsKey(4) || !mm.containsKey(-4)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            } else {
                Map<Integer, Integer> m1 = mm.get(4);
                Map<Integer, Integer> m2 = mm.get(-4);
                if (m1.get(1) != 1 || m1.get(2) != 2 || m1.get(3) != 3 || m1.get(4) != 4 ||
                        m2.get(-1) != -1 || m2.get(-2) != -2 || m2.get(-3) != -3 || m2.get(-4) != -4) {
                    returnCode |= 1;
                    System.out.println("*** FAILURE ***\n");
                }
            }

            /**
             * INSANITY TEST
             */

            boolean insanityFailed = true;
            try {
                Xtruct hello = new Xtruct();
                hello.string_thing = "Hello2";
                hello.byte_thing = 2;
                hello.i32_thing = 2;
                hello.i64_thing = 2;

                Xtruct goodbye = new Xtruct();
                goodbye.string_thing = "Goodbye4";
                goodbye.byte_thing = (byte) 4;
                goodbye.i32_thing = 4;
                goodbye.i64_thing = (long) 4;

                insane.userMap = new HashMap<>();
                insane.userMap.put(Numberz.EIGHT, (long) 8);
                insane.userMap.put(Numberz.FIVE, (long) 5);
                insane.xtructs = new ArrayList<>();
                insane.xtructs.add(goodbye);
                insane.xtructs.add(hello);

                Map<Long, Map<Numberz, Insanity>> whoa =
                        testClient.testInsanity(context, insane);
                if (whoa.size() == 2 && whoa.containsKey(1L) && whoa.containsKey(2L)) {
                    Map<Numberz, Insanity> first_map = whoa.get(1L);
                    Map<Numberz, Insanity> second_map = whoa.get(2L);

                    if (first_map.size() == 2 &&
                            first_map.containsKey(Numberz.TWO) &&
                            first_map.containsKey(Numberz.THREE) &&
                            insane.equals(first_map.get(Numberz.TWO)) &&
                            insane.equals(first_map.get(Numberz.THREE))) {
                              insanityFailed = false;
                    }
                }
            } catch (Exception ex) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
                ex.printStackTrace(System.out);
                insanityFailed = false;
            }
            if (insanityFailed) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * EXECPTION TEST
             */
            try {
                testClient.testException(context, "Xception");
                System.out.print("  void\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (Xception e) {
                System.out.printf("  {%d, \"%s\"}\n", e.errorCode, e.message);
            }

            try {
                testClient.testException(context, "success");
            } catch (Exception e) {
                System.out.printf("  exception\n*** FAILURE ***\n");
                returnCode |= 1;
            }


            /**
             * MULTI EXCEPTION TEST
             */

            try {
                testClient.testMultiException(context, "Xception", "test 1");
                System.out.print("  result\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (Xception e) {
                System.out.printf("  {%d, \"%s\"}\n", e.errorCode, e.message);
            }

            try {
                testClient.testMultiException(context, "Xception2", "test 2");
                System.out.print("  result\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (Xception2 e) {
                System.out.printf("  {%d, {\"%s\"}}\n", e.errorCode, e.struct_thing.string_thing);
            }

            try {
                testClient.testMultiException(context, "success", "test 3");
            } catch (Exception e) {
                System.out.printf("  exception\n*** FAILURE ***\n");
                returnCode |= 1;
            }

            /**
             * ONEWAY TEST
             */
            try {
                testClient.testOneway(context, 1);
            } catch (Exception e) {
                System.out.print("  exception\n*** FAILURE ***\n");
                System.out.println(e);
                returnCode |= 1;
            }

            /**
             * PUB/SUB TEST
             * Publish a message, verify that a subscriber receives the message and publishes a response.
             * Verifies that scopes are correctly generated.
             */
            BlockingQueue<Integer> queue = new ArrayBlockingQueue<>(1);
            Object o = null;
            FScopeTransportFactory factory = new FNatsScopeTransport.Factory(conn);
            FScopeProvider provider = new FScopeProvider(factory,  new FProtocolFactory(protocolFactory));

            EventsSubscriber subscriber = new EventsSubscriber(provider);
            subscriber.subscribeEventCreated(Integer.toString(port)+"-response", (ctx, event) -> {
                System.out.println("Response received " + event);
                queue.add(1);
            });

            EventsPublisher publisher = new EventsPublisher(provider);
            publisher.open();
            Event event = new Event(1, "Sending Call");
            publisher.publishEventCreated(new FContext("Call"), Integer.toString(port)+"-call", event);
            System.out.print("Publishing...    ");

            try {
                o = queue.poll(3, TimeUnit.SECONDS);
            } catch (InterruptedException e){
                System.out.println("InterruptedException " + e);
            }

            if(o == null) {
                System.out.println("Pub/Sub response timed out!");
                returnCode = 1;
            }

        } catch (Exception x) {
            System.out.println("Exception: " + x);
            x.printStackTrace();
            returnCode |= 1;
        }

        if (middlewareCalled) {
            System.out.println("Middleware successfully called.");
        } else {
            System.out.println("Middleware never invoked!");
            returnCode = 1;
        }

        System.exit(returnCode);
    }

    public static class ClientMiddleware implements ServiceMiddleware {

        @Override
        public <T> InvocationHandler<T> apply(T next) {
            return new InvocationHandler<T>(next) {
                @Override
                public Object invoke(Method method, Object receiver, Object[] args) throws Throwable {
                    Object[] subArgs = Arrays.copyOfRange(args, 1, args.length);
                    System.out.printf("%s(%s) = ", method.getName(), Arrays.toString(subArgs));
                    middlewareCalled = true;
                    try {
                        Object ret = method.invoke(receiver, args);
                        System.out.printf("%s \n", ret);
                        return ret;
                    } catch (Exception e) {
                        throw e;
                    }
                }
            };
        }
    }

}