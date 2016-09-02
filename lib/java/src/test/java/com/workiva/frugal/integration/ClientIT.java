package com.workiva.frugal.integration;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.transport.FHttpTransport;
import com.workiva.frugal.transport.FNatsScopeTransport;
import com.workiva.frugal.transport.FNatsTransport;
import com.workiva.frugal.transport.FScopeTransportFactory;
import com.workiva.frugal.transport.FTransport;
import frugal.test.*;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransportException;

import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.TimeUnit;

/**
 * Test Java client for frugal.
 * This makes a variety of requests to enable testing for both performance and
 * correctness of the output.
 */
public class ClientIT {

    public static void main(String[] args) throws Exception {
        // default testing parameters, overwritten in Python runner
        String host = "localhost";
        Integer port = 9090;
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
                case "stateless":
                    fTransport = FNatsTransport.of(conn, Integer.toString(port));
                    break;
            }
        } catch (Exception x) {
            x.printStackTrace();
            System.exit(1);
        }

        try {
            fTransport.open();
        } catch (TTransportException e) {
            e.printStackTrace();
            System.out.println("Failed to open fTransport: " + e.getMessage());
            System.exit(1);
        }

        FFrugalTest.Client testClient = new FFrugalTest.Client(fTransport, new FProtocolFactory(protocolFactory));

        Insanity insane = new Insanity();
        FContext context = new FContext("");

        int returnCode = 0;
        try {
            /**
             * VOID TEST
             */

            try {
                System.out.print("testVoid()");
                testClient.testVoid(context);
                System.out.print(" = void\n");
            } catch (TApplicationException tax) {
                tax.printStackTrace();
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * STRING TEST
             */
            System.out.print("testString(\"Test\")");
            String s = testClient.testString(context, "Test");
            System.out.print(" = \"" + s + "\"\n");
            if (!s.equals("Test")) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * BYTE TEST
             */
            System.out.print("testByte(1)");
            byte i8 = testClient.testByte(context, (byte) 1);
            System.out.print(" = " + i8 + "\n");
            if (i8 != 1) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * I32 TEST
             */
            System.out.print("testI32(-1)");
            int i32 = testClient.testI32(context, -1);
            System.out.print(" = " + i32 + "\n");
            if (i32 != -1) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * I64 TEST
             */
            System.out.print("testI64(-34359738368)");
            long i64 = testClient.testI64(context, -34359738368L);
            System.out.print(" = " + i64 + "\n");
            if (i64 != -34359738368L) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * DOUBLE TEST
             */
            System.out.print("testDouble(-5.325098235)");
            double dub = testClient.testDouble(context, -5.325098235);
            System.out.print(" = " + dub + "\n");
            if (Math.abs(dub - (-5.325098235)) > 0.001) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * BINARY TEST
             */
            try {
                System.out.print("testBinary(-128...127) = ");
                // There currently a mismatch between java and go that will cause test failures if the length of this array is not divisible by 4
                // TODO: Use commented line in lieu of modified line once a fix is in Thrift
                // byte[] data = new byte[]{-128, -127, -126, -125, -124, -123, -122, -121, -120, -119, -118, -117, -116, -115, -114, -113, -112, -111, -110, -109, -108, -107, -106, -105, -104, -103, -102, -101, -100, -99, -98, -97, -96, -95, -94, -93, -92, -91, -90, -89, -88, -87, -86, -85, -84, -83, -82, -81, -80, -79, -78, -77, -76, -75, -74, -73, -72, -71, -70, -69, -68, -67, -66, -65, -64, -63, -62, -61, -60, -59, -58, -57, -56, -55, -54, -53, -52, -51, -50, -49, -48, -47, -46, -45, -44, -43, -42, -41, -40, -39, -38, -37, -36, -35, -34, -33, -32, -31, -30, -29, -28, -27, -26, -25, -24, -23, -22, -21, -20, -19, -18, -17, -16, -15, -14, -13, -12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127};
                byte[] data = new byte[]{-127, -126, -125, -124, -123, -122, -121, -120, -119, -118, -117, -116, -115, -114, -113, -112, -111, -110, -109, -108, -107, -106, -105, -104, -103, -102, -101, -100, -99, -98, -97, -96, -95, -94, -93, -92, -91, -90, -89, -88, -87, -86, -85, -84, -83, -82, -81, -80, -79, -78, -77, -76, -75, -74, -73, -72, -71, -70, -69, -68, -67, -66, -65, -64, -63, -62, -61, -60, -59, -58, -57, -56, -55, -54, -53, -52, -51, -50, -49, -48, -47, -46, -45, -44, -43, -42, -41, -40, -39, -38, -37, -36, -35, -34, -33, -32, -31, -30, -29, -28, -27, -26, -25, -24, -23, -22, -21, -20, -19, -18, -17, -16, -15, -14, -13, -12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127};
                ByteBuffer bin = testClient.testBinary(context, ByteBuffer.wrap(data));

                bin.mark();
                byte[] bytes = new byte[bin.limit() - bin.position()];
                bin.get(bytes);
                bin.reset();
                System.out.print("{");
                boolean first = true;
                for (byte aByte : bytes) {
                    if (first)
                        first = false;
                    else
                        System.out.print(", ");
                    System.out.print(aByte);
                }
                System.out.println("}");
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
            System.out.print("testStruct({\"Zero\", 1, -3, -5})");
            Xtruct out = new Xtruct();
            out.string_thing = "Zero";
            out.byte_thing = (byte) 1;
            out.i32_thing = -3;
            out.i64_thing = -5;
            Xtruct in = testClient.testStruct(context, out);
            System.out.print(" = {" + "\"" +
                    in.string_thing + "\"," +
                    in.byte_thing + ", " +
                    in.i32_thing + ", " +
                    in.i64_thing + "}\n");

            if (!in.equals(out)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * NESTED STRUCT TEST
             */
            System.out.print("testNest({1, {\"Zero\", 1, -3, -5}), 5}");
            Xtruct2 out2 = new Xtruct2();
            out2.byte_thing = (short) 1;
            out2.struct_thing = out;
            out2.i32_thing = 5;
            Xtruct2 in2 = testClient.testNest(context, out2);
            in = in2.struct_thing;
            System.out.print(" = {" + in2.byte_thing + ", {" + "\"" +
                    in.string_thing + "\", " +
                    in.byte_thing + ", " +
                    in.i32_thing + ", " +
                    in.i64_thing + "}, " +
                    in2.i32_thing + "}\n");
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
            System.out.print("testMap({");
            boolean first = true;
            for (int key : mapout.keySet()) {
                if (first) {
                    first = false;
                } else {
                    System.out.print(", ");
                }
                System.out.print(key + " => " + mapout.get(key));
            }
            System.out.print("})");
            Map<Integer, Integer> mapin = testClient.testMap(context, mapout);
            System.out.print(" = {");
            first = true;
            for (int key : mapin.keySet()) {
                if (first) {
                    first = false;
                } else {
                    System.out.print(", ");
                }
                System.out.print(key + " => " + mapout.get(key));
            }
            System.out.print("}\n");

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
                for (String key : smapout.keySet()) {
                    if (first) {
                        first = false;
                    } else {
                        System.out.print(", ");
                    }
                    System.out.print(key + " => " + smapout.get(key));
                }
                System.out.print("})");
                Map<String, String> smapin = testClient.testStringMap(context, smapout);
                System.out.print(" = {");
                first = true;
                for (String key : smapin.keySet()) {
                    if (first) {
                        first = false;
                    } else {
                        System.out.print(", ");
                    }
                    System.out.print(key + " => " + smapout.get(key));
                }
                System.out.print("}\n");
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
            System.out.print("testSet({");
            first = true;
            for (int elem : setout) {
                if (first) {
                    first = false;
                } else {
                    System.out.print(", ");
                }
                System.out.print(elem);
            }
            System.out.print("})");
            Set<Integer> setin = testClient.testSet(context, setout);
            System.out.print(" = {");
            first = true;
            for (int elem : setin) {
                if (first) {
                    first = false;
                } else {
                    System.out.print(", ");
                }
                System.out.print(elem);
            }
            System.out.print("}\n");
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
            System.out.print("testList({");
            first = true;
            for (int elem : listout) {
                if (first) {
                    first = false;
                } else {
                    System.out.print(", ");
                }
                System.out.print(elem);
            }
            System.out.print("})");
            List<Integer> listin = testClient.testList(context, listout);
            System.out.print(" = {");
            first = true;
            for (int elem : listin) {
                if (first) {
                    first = false;
                } else {
                    System.out.print(", ");
                }
                System.out.print(elem);
            }
            System.out.print("}\n");
            if (!listout.equals(listin)) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * ENUM TEST
             */
            System.out.print("testEnum(ONE)");
            Numberz ret = testClient.testEnum(context, Numberz.ONE);
            System.out.print(" = " + ret + "\n");
            if (ret != Numberz.ONE) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            System.out.print("testEnum(TWO)");
            ret = testClient.testEnum(context, Numberz.TWO);
            System.out.print(" = " + ret + "\n");
            if (ret != Numberz.TWO) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            System.out.print("testEnum(THREE)");
            ret = testClient.testEnum(context, Numberz.THREE);
            System.out.print(" = " + ret + "\n");
            if (ret != Numberz.THREE) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            System.out.print("testEnum(FIVE)");
            ret = testClient.testEnum(context, Numberz.FIVE);
            System.out.print(" = " + ret + "\n");
            if (ret != Numberz.FIVE) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            System.out.print("testEnum(EIGHT)");
            ret = testClient.testEnum(context, Numberz.EIGHT);
            System.out.print(" = " + ret + "\n");
            if (ret != Numberz.EIGHT) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * TYPEDEF TEST
             */
            System.out.print("testTypedef(309858235082523)");
            long uid = testClient.testTypedef(context, 309858235082523L);
            System.out.print(" = " + uid + "\n");
            if (uid != 309858235082523L) {
                returnCode |= 1;
                System.out.println("*** FAILURE ***\n");
            }

            /**
             * NESTED MAP TEST
             */
            System.out.print("testMapMap(1)");
            Map<Integer, Map<Integer, Integer>> mm =
                    testClient.testMapMap(context, 1);
            System.out.print(" = {");
            for (int key : mm.keySet()) {
                System.out.print(key + " => {");
                Map<Integer, Integer> m2 = mm.get(key);
                for (int k2 : m2.keySet()) {
                    System.out.print(k2 + " => " + m2.get(k2) + ", ");
                }
                System.out.print("}, ");
            }
            System.out.print("}\n");
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

                System.out.print("testInsanity()");
                Map<Long, Map<Numberz, Insanity>> whoa =
                        testClient.testInsanity(context, insane);
                System.out.print(" = {");
                for (long key : whoa.keySet()) {
                    Map<Numberz, Insanity> val = whoa.get(key);
                    System.out.print(key + " => {");

                    for (Numberz k2 : val.keySet()) {
                        Insanity v2 = val.get(k2);
                        System.out.print(k2 + " => {");
                        Map<Numberz, Long> userMap = v2.userMap;
                        System.out.print("{");
                        if (userMap != null) {
                            for (Numberz k3 : userMap.keySet()) {
                                System.out.print(k3 + " => " + userMap.get(k3) + ", ");
                            }
                        }
                        System.out.print("}, ");

                        List<Xtruct> xtructs = v2.xtructs;
                        System.out.print("{");
                        if (xtructs != null) {
                            for (Xtruct x : xtructs) {
                                System.out.print("{" + "\"" + x.string_thing + "\", " + x.byte_thing + ", " + x.i32_thing + ", " + x.i64_thing + "}, ");
                            }
                        }
                        System.out.print("}");

                        System.out.print("}, ");
                    }
                    System.out.print("}, ");
                }
                System.out.print("}\n");
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
                System.out.print("testException(\"Xception\") =>");
                testClient.testException(context, "Xception");
                System.out.print("  void\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (Xception e) {
                System.out.printf("  {%d, \"%s\"}\n", e.errorCode, e.message);
            }

            try {
                System.out.print("testException(\"success\") =>");
                testClient.testException(context, "success");
                System.out.print("  void\n");
            } catch (Exception e) {
                System.out.printf("  exception\n*** FAILURE ***\n");
                returnCode |= 1;
            }


            /**
             * MULTI EXCEPTION TEST
             */

            try {
                System.out.printf("testMultiException(\"Xception\", \"test 1\") =>");
                testClient.testMultiException(context, "Xception", "test 1");
                System.out.print("  result\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (Xception e) {
                System.out.printf("  {%d, \"%s\"}\n", e.errorCode, e.message);
            }

            try {
                System.out.printf("testMultiException(\"Xception2\", \"test 2\") =>");
                testClient.testMultiException(context, "Xception2", "test 2");
                System.out.print("  result\n*** FAILURE ***\n");
                returnCode |= 1;
            } catch (Xception2 e) {
                System.out.printf("  {%d, {\"%s\"}}\n", e.errorCode, e.struct_thing.string_thing);
            }

            try {
                System.out.print("testMultiException(\"success\", \"test 3\") =>");
                Xtruct result;
                result = testClient.testMultiException(context, "success", "test 3");
                System.out.printf("  {{\"%s\"}}\n", result.string_thing);
            } catch (Exception e) {
                System.out.printf("  exception\n*** FAILURE ***\n");
                returnCode |= 1;
            }

            /**
             * ONEWAY TEST
             */
            System.out.print("testOneWay(1)...");
            try {
                testClient.testOneway(context, 1);
                System.out.print(" - no error returned\n");
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

        System.exit(returnCode);
    }

}