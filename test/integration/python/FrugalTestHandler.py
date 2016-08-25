import time

from frugal_test.f_FrugalTest import Iface, Xtruct, Xception, Xception2, Insanity


class FrugalTestHandler(Iface):
    def testVoid(self, ctx):
        print "test_void({})".format(ctx)
        return

    def testString(self, ctx, thing):
        print "test_string({})".format(thing)
        return thing

    def testBool(self, ctx, thing):
        print "test_bool({})".format(thing)
        return thing

    def testByte(self, ctx, thing):
        print "test_byte({})".format(thing)
        return thing

    def testI32(self, ctx, thing):
        print "test_i32({})".format(thing)
        return thing

    def testI64(self, ctx, thing):
        print "test_i64({})".format(thing)
        return thing

    def testDouble(self, ctx, thing):
        print "test_double({})".format(thing)
        return thing

    def testBinary(self, ctx, thing):
        print "test_binary({})".format(thing)
        return thing

    def testStruct(self, ctx, thing):
        print "test_struct({})".format(thing)
        return thing

    def testNest(self, ctx, thing):
        print "test_nest({})".format(thing)
        return thing

    def testMap(self, ctx, thing):
        print "test_map({})".format(thing)
        return thing

    def testStringMap(self, ctx, thing):
        print "test_string_map({})".format(thing)
        return thing

    def testSet(self, ctx, thing):
        print "test_set({})".format(thing)
        return thing

    def testList(self, ctx, thing):
        print "test_list({})".format(thing)
        return thing

    def testEnum(self, ctx, thing):
        print "test_enum({})".format(thing)
        return thing

    def testMapMap(self, ctx, hello):
        print "test_map_map({})".format(hello)
        d = {4: 4, 3: 3, 2: 2, 1: 1}
        e = {-4: -4, -3: -3, -2: -2, -1: -1}
        mapmap = {-4: e, 4: d}
        return mapmap

    def testInsanity(self, ctx, argument):
        print "test_insanity({})".format(argument)
        return {1:
                {2: argument,
                 3: argument}, 2: {}}

    def testMulti(self, ctx, arg0, arg1, arg2, arg3, arg4, arg5):
        print "test_multi({}, {}, {}, {}, {}, {})".format(arg0, arg1, arg2, arg3, arg4, arg5)
        result = Xtruct()
        result.string_thing = "Hello2"
        result.byte_thing = arg0
        result.i32_thing = arg1
        result.i64_thing = arg2
        return result

    def testException(self, ctx, arg):
        print "test_exception({})".format(arg)
        if arg == "Xception":
            e = Xception()
            e.errorCode = 1001
            e.message = arg
            raise e
        elif arg == "TException":
            raise Xception.message("Just TException")
        return

    def testMultiException(self, ctx, arg0, arg1):
        print "test_multi_exception({}, {})".format(arg0, arg1)
        if arg0 == "Xception":
            e = Xception()
            e.errorCode = 1001
            e.message = "This is an Xception"
            raise e
        elif arg0 == "Xception2":
            e = Xception2()
            e.errorCode = 2002
            e.struct_thing = Xtruct()
            e.struct_thing.string_thing = "This is an Xception2"
            raise e
        else:
            r = Xtruct()
            r.string_thing = arg1
            return r

    def testOneway(self, ctx, seconds):
        print "test_oneway({}): Sleeping...".format(seconds)
        time.sleep(seconds)
        print "testOneway({}): done sleeping!".format(seconds)
        return
