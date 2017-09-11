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

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TTransportExceptionType;
import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.provider.FServiceProvider;
import com.workiva.frugal.transport.FHttpTransport;
import com.workiva.frugal.transport.FTransport;
import com.workiva.frugal.transport.FNatsTransport;
import com.workiva.frugal.transport.FPublisherTransportFactory;
import com.workiva.frugal.transport.FNatsPublisherTransport;
import com.workiva.frugal.transport.FNatsSubscriberTransport;
import com.workiva.frugal.transport.FSubscriberTransportFactory;
import com.workiva.utils;
import frugal.test.*;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransportException;

import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.util.*;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.TimeUnit;

import static com.workiva.utils.whichProtocolFactory;

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
                    System.out.println("  --transport=arg (=" + transport_type + ")\n\t\t\t\tTransport: stateless, http");
                    System.out.println("  --protocol=arg (=" + protocol_type + ")\tProtocol: binary, json, compact");
                    System.exit(0);
                }
            }
        } catch (Exception x) {
            System.err.println("Can not parse arguments! See --help");
            System.err.println("Exception parsing arguments: " + x);
            System.exit(1);
        }
        TProtocolFactory protocolFactory = whichProtocolFactory(protocol_type);

        List<String> validTransports = new ArrayList<>();
        validTransports.add("stateless");
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
                    // Set request and response size limit to 1mb
                    int maxSize = 1048576;
                    FHttpTransport.Builder httpTransport = new FHttpTransport.Builder(httpClient, url).withRequestSizeLimit(maxSize).withResponseSizeLimit(maxSize);
                    fTransport = httpTransport.build();
                    fTransport.open();
                    break;
                case "stateless":
                    fTransport = FNatsTransport.of(conn, "frugal.foo.bar.rpc." + Integer.toString(port));
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

        FFrugalTest.Client testClient = new FFrugalTest.Client(new FServiceProvider(fTransport, new FProtocolFactory(protocolFactory)), new ClientMiddleware());

        Insanity insane = new Insanity();
        FContext context;

        int returnCode = 0;
        try {
            /**
             * VOID TEST
             */

            try {
                context = new FContext("testVoid");
                testClient.testVoid(context);
            } catch (TApplicationException tax) {
                tax.printStackTrace();
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * STRING TEST
             */
            context = new FContext("testString");
            String strInput = "Testå∫ç";
            String s = testClient.testString(context, strInput);
            if (!s.equals(strInput)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * BOOL TESTS
             */
            context = new FContext("testBool");
            boolean bl = testClient.testBool(context, true);
            if (!bl) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            bl = testClient.testBool(context, false);
            if (bl) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * BYTE TEST
             */
            context = new FContext("testByte");
            byte i8 = testClient.testByte(context, (byte) 1);
            if (i8 != 1) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * I32 TEST
             */
            context = new FContext("testI32");
            int i32 = testClient.testI32(context, -1);
            if (i32 != -1) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * I64 TEST
             */
            context = new FContext("testI64");
            long i64 = testClient.testI64(context, -34359738368L);
            if (i64 != -34359738368L) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * DOUBLE TEST
             */
            context = new FContext("testDouble");
            double dub = testClient.testDouble(context, -5.325098235);
            if (Math.abs(dub - (-5.325098235)) > 0.001) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * BINARY TEST
             */
            context = new FContext("testBinary");
            try {
                // verify the byte[] is able to be encoded as UTF-8 to avoid deserialization errors in clients
                byte[] data = "foo".getBytes("UTF-8");
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
            context = new FContext("testStruct");
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
            context = new FContext("testXtruct2");
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
            context = new FContext("testMap");
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
            context = new FContext("testStringMap");
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
            context = new FContext("testSet");
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
            context = new FContext("testList");
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
             * Note: No enum type check. Java is statically typed
             */
            context = new FContext("testEnum");
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
            context = new FContext("testTypedef");
            long uid = testClient.testTypedef(context, 309858235082523L);
            if (uid != 309858235082523L) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * NESTED MAP TEST
             */
            context = new FContext("testMapMap");
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
             * BOOL TESTS
             */
            context = new FContext("TestUppercaseMethod");
            boolean uppercase = testClient.TestUppercaseMethod(context, true);
            if (!uppercase) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * INSANITY TEST
             */
            context = new FContext("testInsanity");
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
             * UNCHECKED EXCEPTION TEST
             */
            context = new FContext("testUncaughtException");
            try {
                testClient.testUncaughtException(context);
                System.out.print("  void\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (TApplicationException e) {
                boolean failed = false;
                String expectedMessage = "An uncaught error";
                int expectedErrorType = 6;
                if (e.getType() != expectedErrorType){
                    System.out.printf("  Expected type %d, got type %d\n", expectedErrorType, e.getType());
                    failed = true;
                }
                if (!e.getMessage().contains(expectedMessage)){
                    System.out.printf("  Expected message %s, got message %s\n", expectedMessage, e.getMessage());
                    failed = true;
                }
                if (failed){
                    System.out.print("  void\n*** FAILURE ***\n");
                    returnCode |= 1;
                } else {
                    System.out.printf("  {\"%s\"}\n", e);
                }

                if (e.getType() != expectedErrorType){
                    System.out.printf("  Expected type %d, got type %d\n", expectedErrorType, e.getType());
                    System.out.printf("  void\n*** FAILURE ***\n");
                    returnCode |= 1;
                } else {
                    System.out.printf("  {\"%s\"}\n", e);
                }
            }


            /**
             * UNCHECKED TAPPLICATION EXCEPTION TEST
             */
            context = new FContext("testUncheckedTApplicationException");
            try {
                testClient.testUncheckedTApplicationException(context);
                System.out.print("  void\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (TApplicationException e) {
                boolean failed = false;
                String expectedMessage = "Unchecked TApplicationException";
                int expectedErrorType = 400;
                if (e.getType() != expectedErrorType){
                    System.out.printf("  Expected type %d, got type %d\n", expectedErrorType, e.getType());
                    failed = true;
                }
                if (!e.getMessage().contains(expectedMessage)){
                    System.out.printf("  Expected message %s, got message %s\n", expectedMessage, e.getMessage());
                    failed = true;
                }
                if (failed){
                    System.out.print("  void\n*** FAILURE ***\n");
                    returnCode |= 1;
                } else {
                    System.out.printf("  {\"%s\"}\n", e);
                }
            }

            /**
             * EXECPTION TEST
             */
            context = new FContext("testException");
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
            context = new FContext("testMultiException");
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
             * REQUEST TOO LARGE TEST
             */
            context = new FContext("testRequestTooLarge");
            try {
                java.nio.ByteBuffer request = ByteBuffer
                        .allocate(1024*1024);
                testClient.testRequestTooLarge(context, request);
                System.out.print("\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (TTransportException e) {
                System.out.println("Failed Request Too Large");
                System.out.println(e);
                if (e.getType() != TTransportExceptionType.REQUEST_TOO_LARGE) {
                    System.out.printf("  Unexpected exception %s\n", e);
                    System.out.print("\n*** FAILURE ***\n");
                    returnCode |= 1;
                }
            }


            /**
             * RESPONSE TOO LARGE TEST
             */
            context = new FContext("testResponseTooLarge");
            java.nio.ByteBuffer response;
            try {
                java.nio.ByteBuffer request = ByteBuffer.allocate(1);
                response = testClient.testResponseTooLarge(context, request);
                System.out.print("  result\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (TTransportException e) {
                TTransportException expectedException = new TTransportException(TTransportExceptionType.RESPONSE_TOO_LARGE);
                if (e.getType() != expectedException.getType()) {
                    System.out.printf("  Unexpected exception %s\n", e);
                    System.out.print("  result\n*** FAILURE ***\n");
                    returnCode |= 1;
                }
            }

            /**
             * ONEWAY TEST
             */
            context = new FContext("testOneway");
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
            FPublisherTransportFactory publisherFactory = new FNatsPublisherTransport.Factory(conn);
            FSubscriberTransportFactory subscriberFactory = new FNatsSubscriberTransport.Factory(conn);
            FScopeProvider provider = new FScopeProvider(publisherFactory, subscriberFactory, new FProtocolFactory(protocolFactory));

            String preamble = "foo";
            String ramble = "bar";
            EventsSubscriber.Iface subscriber = new EventsSubscriber.Client(provider);
            subscriber.subscribeEventCreated(preamble, ramble, "response", Integer.toString(port), (ctx, event) -> {
                System.out.println("Response received " + event);
                queue.add(1);
            });

            EventsPublisher.Iface publisher = new EventsPublisher.Client(provider);
            publisher.open();
            Event event = new Event(1, "Sending Call");
            FContext ctx = new FContext("Call");
            ctx.addRequestHeader(utils.PREAMBLE_HEADER, preamble);
            ctx.addRequestHeader(utils.RAMBLE_HEADER, ramble);
            publisher.publishEventCreated(ctx, preamble, ramble, "call", Integer.toString(port), event);
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
