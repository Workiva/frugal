from frugal_test import ttypes, Xception, Insanity, Xception2, Event
from frugal_test.f_FrugalTest import Xtruct, Xtruct2, Numberz


def rpc_test_definitions():
    """
    Defines and returns shared tests for all python frugal implementations.

    :return: dictionary with:
        keys=rpc name
        values= dict with array of args and expected test result
    """
    tests = {}

    tests['testVoid'] = dict(args=None, expected_result=None)

    thing = "thing"
    tests['testString'] = dict(args=[thing], expected_result=thing)

    tests['testBool'] = dict(args=[True], expected_result=True)

    byte = 42
    tests['testByte'] = dict(args=[byte], expected_result=byte)

    i32 = 4242
    tests['testI32'] = dict(args=[i32], expected_result=i32)

    i64 = 424242
    tests['testI64'] = dict(args=[i64], expected_result=i64)

    dbl = 42.42
    tests['testDouble'] = dict(args=[dbl], expected_result=dbl)

    binary = b'101010'
    tests['testBinary'] = dict(args=[binary], expected_result=binary)

    struct = Xtruct()
    struct.string_thing = thing
    struct.byte_thing = byte
    struct.i32_thing = i32
    struct.i64_thing = i64
    tests['testStruct'] = dict(args=[struct], expected_result=struct)

    struct2 = Xtruct2()
    struct2.struct_thing = struct
    struct2.byte_thing = 0
    struct2.i32_thing = 0
    tests['testNest'] = dict(args=[struct2], expected_result=struct2)

    dictionary = {1: 2, 3: 4, 5: 42}
    tests['testMap'] = dict(args=[dictionary], expected_result=dictionary)

    string_map = {"a": "2", "b": "blah", "some": "thing"}
    tests['testStringMap'] = dict(args=[string_map], expected_result=string_map)

    set = {1, 2, 2, 42}
    tests['testSet'] = dict(args=[set], expected_result=set)

    list = [1, 2, 42]
    tests['testList'] = dict(args=[list], expected_result=list)

    enum = Numberz.TWO
    tests['testEnum'] = dict(args=[enum], expected_result=enum)

    type_def = 42
    tests['testTypedef'] = dict(args=[type_def], expected_result=type_def)

    d = {4: 4, 3: 3, 2: 2, 1: 1}
    e = {-4: -4, -3: -3, -2: -2, -1: -1}
    mapmap = {-4: e, 4: d}
    tests['testMapMap'] = dict(args=[42], expected_result=mapmap)

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
    tests['testInsanity'] = dict(args=[insanity], expected_result=expected_result)

    multi = Xtruct()
    multi.string_thing = "Hello2"
    multi.byte_thing = byte
    multi.i32_thing = i32
    multi.i64_thing = i64
    tests['testMulti'] = dict(args=[byte, i32, i64, {1: "blah", 2: thing}, Numberz.EIGHT, 24], expected_result=multi)

    tests['testException'] = dict(
        args=['Xception'],
        expected_result=Xception(errorCode=1001, message='Xception')
    )

    struct_thing = Xtruct()
    struct_thing.string_thing = 'This is an Xception2'
    struct_thing.byte_thing = 0
    struct_thing.i32_thing = 0
    struct_thing.i64_thing = 0
    e = Xception2(errorCode=2002, struct_thing=struct_thing)
    tests['testMultiException'] = dict(
        args=['Xception2', 'ignoreme'],
        expected_result=e
    )

    return tests
