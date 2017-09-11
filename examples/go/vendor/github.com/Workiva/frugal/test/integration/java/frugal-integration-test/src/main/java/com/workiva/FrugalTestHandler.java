package com.workiva;

import com.workiva.frugal.FContext;
import frugal.test.*;
import org.apache.thrift.TException;
import org.apache.thrift.TApplicationException;
import frugal.test.Numberz;

import java.nio.ByteBuffer;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;


public class FrugalTestHandler implements FFrugalTest.Iface {

        // Each RPC handler "test___" accepts a value of type ___ and returns the same value (where applicable).
        // The client then asserts that the returned value is equal to the value sent.
        @Override
        public void testVoid(FContext ctx) throws TException {
        }

        @Override
        public String testString(FContext ctx, String thing) throws TException {
            return thing;
        }

        @Override
        public boolean testBool(FContext ctx, boolean thing) throws TException {
            return thing;
        }

        @Override
        public byte testByte(FContext ctx, byte thing) throws TException {
            return thing;
        }

        @Override
        public int testI32(FContext ctx, int thing) throws TException {
            return thing;
        }

        @Override
        public long testI64(FContext ctx, long thing) throws TException {
            return thing;
        }

        @Override
        public double testDouble(FContext ctx, double thing) throws TException {
            return thing;
        }

        @Override
        public ByteBuffer testBinary(FContext ctx, ByteBuffer thing) throws TException {
            return thing;
        }

        @Override
        public Xtruct testStruct(FContext ctx, Xtruct thing) throws TException {
            return thing;
        }

        @Override
        public Xtruct2 testNest(FContext ctx, Xtruct2 thing) throws TException {
            return thing;
        }

        @Override
        public Map<Integer, Integer> testMap(FContext ctx, Map<Integer, Integer> thing) throws TException {
            return thing;
        }

        @Override
        public Map<String, String> testStringMap(FContext ctx, Map<String, String> thing) throws TException {
            return thing;
        }

        @Override
        public Set<Integer> testSet(FContext ctx, Set<Integer> thing) throws TException {
            return thing;
        }

        @Override
        public List<Integer> testList(FContext ctx, List<Integer> thing) throws TException {
            return thing;
        }

        @Override
        public Numberz testEnum(FContext ctx, Numberz thing) throws TException {
            return thing;
        }

        @Override
        public long testTypedef(FContext ctx, long thing) throws TException {
            return thing;
        }

        @Override
        public Map<Integer, Map<Integer, Integer>> testMapMap(FContext ctx, int hello) throws TException {
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
        public boolean TestUppercaseMethod(FContext ctx, boolean thing) throws TException {
            return thing;
        }

        @Override
        public Map<Long, Map<Numberz, Insanity>> testInsanity(FContext ctx, Insanity argument) throws TException {
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
            Xtruct r = new Xtruct();

            r.string_thing = "Hello2";
            r.byte_thing = arg0;
            r.i32_thing = arg1;
            r.i64_thing = arg2;

            return r;
        }

        @Override
        public void testException(FContext ctx, String arg) throws TException {
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

        // This doesn't really make the same sense to check in Java
        // as in other languages, because we rethrow any caught runtime exceptions.
        @Override
        public void testUncaughtException(FContext ctx) throws TException {
            throw new TException("An uncaught error which will be caught in Java");
        }

        @Override
        public void testUncheckedTApplicationException(FContext ctx) throws TException {
            throw new TApplicationException(400, "Unchecked TApplicationException");
        }

        @Override
        public Xtruct testMultiException(FContext ctx, String arg0, String arg1) throws TException {
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
        public void testRequestTooLarge(FContext ctx, java.nio.ByteBuffer request) throws TException {
             throw new TApplicationException(400, "testRequestTooLarge should" +
                     " never be sucessfully called.");
        }

        @Override
        public java.nio.ByteBuffer testResponseTooLarge(FContext ctx, java.nio.ByteBuffer request) {
            java.nio.ByteBuffer response = ByteBuffer.allocate(1024*1024);
            return response;
        }

        @Override
        public void testOneway(FContext ctx, int secondsToSleep) throws TException {
        }
}
