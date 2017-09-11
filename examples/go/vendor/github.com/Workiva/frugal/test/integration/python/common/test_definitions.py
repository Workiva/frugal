# -*- coding: utf-8 -*-

import six

from frugal.exceptions import TTransportExceptionType

from frugal_test.ttypes import Xception, Insanity, Xception2
from frugal_test.f_FrugalTest import Xtruct, Xtruct2, Numberz

from thrift.Thrift import TApplicationException
from thrift.transport.TTransport import TTransportException


def rpc_test_definitions(transport):
    """
    Defines and returns shared tests for all python frugal implementations.

    :return: list of tuples of the form (key, value) with:
        key=rpc name
        value= dict with array of args and expected test result
    """
    tests = []

    tests.append(('testVoid', dict(args=None, expected_result=None)))

    thing = u"thing\u2016"
    tests.append(('testString', dict(args=[thing], expected_result=thing)))

    thing2 = "thingå∫ç"
    tests.append(('testString', dict(args=[thing2], expected_result=u"thingå∫ç")))

    tests.append(('testBool', dict(args=[True], expected_result=True)))

    byte = 42
    tests.append(('testByte', dict(args=[byte], expected_result=byte)))

    i32 = 4242
    tests.append(('testI32', dict(args=[i32], expected_result=i32)))

    i64 = 424242
    tests.append(('testI64', dict(args=[i64], expected_result=i64)))

    dbl = 42.42
    tests.append(('testDouble', dict(args=[dbl], expected_result=dbl)))

    binary = b'101010'
    tests.append(('testBinary', dict(args=[binary], expected_result=binary)))

    struct = Xtruct()
    struct.string_thing = thing
    struct.byte_thing = byte
    struct.i32_thing = i32
    struct.i64_thing = i64
    tests.append(('testStruct', dict(args=[struct], expected_result=struct)))

    struct2 = Xtruct2()
    struct2.struct_thing = struct
    struct2.byte_thing = 0
    struct2.i32_thing = 0
    tests.append(('testNest', dict(args=[struct2], expected_result=struct2)))

    dictionary = {1: 2, 3: 4, 5: 42}
    tests.append(('testMap', dict(args=[dictionary], expected_result=dictionary)))

    string_map = {u"\u2018a\u2019": u"\u20182\u2019", u"\u2018b\u2019": u"\u2018blah\u2019", u"\u2018kraken\u2019": u"\u2018thing\u2019"}
    tests.append(('testStringMap', dict(args=[string_map], expected_result=string_map)))

    set = {1, 2, 2, 42}
    tests.append(('testSet', dict(args=[set], expected_result=set)))

    list = [1, 2, 42]
    tests.append(('testList', dict(args=[list], expected_result=list)))

    enum = Numberz(2)
    tests.append(('testEnum', dict(args=[enum], expected_result=enum)))

    type_def = 42
    tests.append(('testTypedef', dict(args=[type_def], expected_result=type_def)))

    d = {4: 4, 3: 3, 2: 2, 1: 1}
    e = {-4: -4, -3: -3, -2: -2, -1: -1}
    mapmap = {-4: e, 4: d}
    tests.append(('testMapMap', dict(args=[42], expected_result=mapmap)))

    tests.append(('TestUppercaseMethod', dict(args=[True], expected_result=True)))

    truck1 = Xtruct("Goodbye4", 4, 4, 4)
    truck2 = Xtruct("Hello2", 2, 2, 2)
    insanity = Insanity()
    insanity.userMap = {Numberz.FIVE: 5, Numberz.EIGHT: 8}
    insanity.xtructs = [truck1, truck2]
    expected_result = {1:
        {2: Insanity(
            xtructs=[Xtruct(string_thing='Goodbye4', byte_thing=4, i32_thing=4, i64_thing=4),
                     Xtruct(string_thing='Hello2', byte_thing=2, i32_thing=2, i64_thing=2)],
            userMap={8: 8, 5: 5}),
            3: Insanity(
                xtructs=[Xtruct(string_thing='Goodbye4', byte_thing=4, i32_thing=4, i64_thing=4),
                         Xtruct(string_thing='Hello2', byte_thing=2, i32_thing=2, i64_thing=2)],
                userMap={8: 8, 5: 5})}, 2: {}}
    tests.append(('testInsanity', dict(args=[insanity], expected_result=expected_result)))

    multi = Xtruct()
    multi.string_thing = "Hello2"
    multi.byte_thing = byte
    multi.i32_thing = i32
    multi.i64_thing = i64
    tests.append(('testMulti', dict(args=[byte, i32, i64, {1: "blah", 2: thing}, Numberz.EIGHT, 24], expected_result=multi)))

    tests.append(('testException', dict(
        args=['Xception'],
        expected_result=Xception(errorCode=1001, message='Xception')
    )))

    struct_thing = Xtruct()
    struct_thing.string_thing = 'This is an Xception2'
    struct_thing.byte_thing = 0
    struct_thing.i32_thing = 0
    struct_thing.i64_thing = 0
    e = Xception2(errorCode=2002, struct_thing=struct_thing)
    tests.append(('testMultiException', dict(
        args=['Xception2', 'ignoreme'],
        expected_result=e
    )))
    e = TApplicationException(TApplicationException.INTERNAL_ERROR, 'An uncaught error')
    tests.append(('testUncaughtException', dict(
        args=[],
        expected_result=e
    )))

    e = TApplicationException(400, 'Unchecked TApplicationException')
    tests.append(('testUncheckedTApplicationException', dict(
        args=[],
        expected_result=e
    )))

    e = TTransportException(TTransportExceptionType.REQUEST_TOO_LARGE)
    tests.append(('testRequestTooLarge', dict(
        args=[six.binary_type(b'0' * (1024 * 1024))],
        expected_result=e
    )))

    e = TTransportException(TTransportExceptionType.RESPONSE_TOO_LARGE)
    tests.append(('testResponseTooLarge', dict(
        args=[six.binary_type(b'0' * 4)],
        expected_result=e
    )))

    return tests
