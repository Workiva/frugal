
import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'dart:typed_data';
import 'package:collection/collection.dart';
import 'package:thrift/thrift.dart';
import 'package:frugal_test/frugal_test.dart';
import 'package:frugal/frugal.dart';

typedef Future FutureFunction();

class FTest {
  final int errorCode;
  final String name;
  final FutureFunction func;

  FTest(this.errorCode, this.name, this.func);
}

class FTestError extends Error {
  final actual;
  final expected;

  FTestError(this.actual, this.expected);

  String toString() => '\n\nUNEXPECTED ERROR \n$actual != \n$expected\n\n';
}

FContext ctx;

List<FTest> createTests(FFrugalTestClient client) {
  List<FTest> tests = [];

  var xtruct = new Xtruct()
    ..string_thing = 'Zero'
    ..byte_thing = 1
    ..i32_thing = -3
    ..i64_thing = -5;

  tests.add(new FTest(1, 'testVoid', () async {
    ctx = new FContext(correlationId: 'testVoid');
    await client.testVoid(ctx);
  }));

  tests.add(new FTest(1, 'testString', () async {
    ctx = new FContext(correlationId: 'testString');
    var input = 'Testå∫ç';
    var result = await client.testString(ctx, input);
    if (result != input) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testBool', () async {
    ctx = new FContext(correlationId: 'testBool');
    var input = true;
    var result = await client.testBool(ctx, input);
    if (result != input) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testByte', () async {
    ctx = new FContext(correlationId: 'testByte');
    var input = 64;
    var result = await client.testByte(ctx, input);
    if (result != input) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testI32', () async {
    ctx = new FContext(correlationId: 'testI32');
    var input = 2147483647;
    var result = await client.testI32(ctx, input);
    if (result != input) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testI64', () async {
    ctx = new FContext(correlationId: 'testI64');
    var input = 9223372036854775807;
    var result = await client.testI64(ctx, input);
    if (result != input) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testDouble', () async {
    ctx = new FContext(correlationId: 'testDouble');
    var input = 3.1415926;
    var result = await client.testDouble(ctx, input);
    if (result != input) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testBinary', () async {
    ctx = new FContext(correlationId: 'testBinary');
    var utf8Codec = const Utf8Codec();
    var input = utf8Codec.encode('foo');
    var result = await client.testBinary(ctx, input);
    var equality = const ListEquality();
    if (!equality.equals(result, input)) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testStruct', () async {
    ctx = new FContext(correlationId: 'testStruct');
    var result = await client.testStruct(ctx, xtruct);
    if ('$result' != '$xtruct') throw new FTestError(result, xtruct);
  }));

  tests.add(new FTest(1, 'testNest', () async {
    ctx = new FContext(correlationId: 'testNest');
    var input = new Xtruct2()
    ..byte_thing = 1
    ..struct_thing = xtruct
    ..i32_thing = -3;

    stdout.write("testNest(${input})");
    var result = await client.testNest(ctx, input);
    if ('$result' != '$input') throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testMap', () async {
    ctx = new FContext(correlationId: 'testMap');
    Map<int, int> input = {1: -10, 2: -9, 3: -8, 4: -7, 5: -6};

    var result = await client.testMap(ctx, input);
    var equality = const MapEquality();
    if (!equality.equals(result, input)) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testSet', () async {
    ctx = new FContext(correlationId: 'testSet');
    var input = new Set.from([-2, -1, 0, 1, 2]);
    var result = await client.testSet(ctx, input);
    var equality = const SetEquality();
    if (!equality.equals(result, input)) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testList', () async {
    ctx = new FContext(correlationId: 'testList');
    var input = [-2, -1, 0, 1, 2];
    var result = await client.testList(ctx, input);
    var equality = const ListEquality();
    if (!equality.equals(result, input)) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testEnum', () async {
    ctx = new FContext(correlationId: 'testEnum');
    await _testEnum(client, Numberz.ONE);
    await _testEnum(client, Numberz.TWO);
    await _testEnum(client, Numberz.THREE);
    await _testEnum(client, Numberz.FIVE);
    await _testEnum(client, Numberz.EIGHT);
  }));

  tests.add(new FTest(1, 'testTypedef', () async {
    ctx = new FContext(correlationId: 'testTypedef');
    var input = 309858235082523;
    var result = await client.testTypedef(ctx, input);
    if (result != input) throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testMapMap', () async {
    ctx = new FContext(correlationId: 'testMapMap');
    Map<int, Map<int, int>> result = await client.testMapMap(ctx, 1);
    if (result.isEmpty || result[result.keys.first].isEmpty) {
      throw new FTestError(result, 'Map<int, Map<int, int>>');
    }
  }));

  tests.add(new FTest(1, 'testUppercaseMethod', () async {
    ctx = new FContext(correlationId: 'testUppercaseMethod');
    var input = true;
    var result = await client.testUppercaseMethod(ctx, input);
    if (result != input) throw new FTestError(result, input);
  }));


  tests.add(new FTest(1, 'testInsanity', () async {
    ctx = new FContext(correlationId: 'testInsanity');
    var input = new Insanity();
    input.userMap = {Numberz.FIVE: 5000};
    input.xtructs = [xtruct];

    Map<int, Map<int, Insanity>> result = await client.testInsanity(ctx, input);
    if (result.isEmpty || result[1].isEmpty) {
      throw new FTestError(result, input);
    }
  }));

  tests.add(new FTest(1, 'testMulti', () async {
    ctx = new FContext(correlationId: 'testMulti');
    var input = new Xtruct()
    ..string_thing = 'Hello2'
    ..byte_thing = 123
    ..i32_thing = 456
    ..i64_thing = 789;

    var result = await client.testMulti(ctx, input.byte_thing, input.i32_thing,
    input.i64_thing, {1: 'one'}, Numberz.EIGHT, 5678);
    if ('$result' != '$input') throw new FTestError(result, input);
  }));

  tests.add(new FTest(1, 'testException', () async {
    ctx = new FContext(correlationId: 'testException');
    try {
      await client.testException(ctx, 'Xception');
    } on Xception catch (exception) {
      return;
    }

    throw new FTestError(null, 'Xception');
  }));

  tests.add(new FTest(1, 'testMultiException', () async {
    ctx = new FContext(correlationId: 'testMultiException');
    try {
      await client.testMultiException(ctx, 'Xception2', 'foo');
    } on Xception2 catch (exception2) {
      return;
    }

    throw new FTestError(null, 'Xception2');
  }));

  tests.add(new FTest(1, 'testOneway', () async {
    ctx = new FContext(correlationId: 'testOneway');
    await client.testOneway(ctx, 1);
  }));

  tests.add(new FTest(1, 'testUncheckedException', () async {
    ctx = new FContext(correlationId: 'testUncheckedException');
    try {
      await client.testUncaughtException(ctx);
    } on TApplicationError catch (e) {

      if (e.type != FrugalTApplicationErrorType.INTERNAL_ERROR ||
      !e.message.contains('An uncaught error')){
        throw new FTestError(e, TApplicationError(FrugalTApplicationErrorType.INTERNAL_ERROR));
      }
    }}));

  tests.add(new FTest(1, 'testUncheckedTApplicationException', () async {
    ctx = new FContext(correlationId: 'testUncheckedTApplicationException');
    try {
      await client.testUncheckedTApplicationException(ctx);
    } on TApplicationError catch (e) {
      int expectedErrorType = 400;
      if (e.type != expectedErrorType ||
      !e.message.contains('Unchecked TApplicationException')) {
        throw new FTestError(e, TApplicationError(expectedErrorType));
      }
    }}));

  tests.add(new FTest(1, 'testRequestTooLarge', () async {
    ctx = new FContext(correlationId: 'testRequestTooLarge');
    var request = new Uint8List(1024*1024);
    try {
      await client.testRequestTooLarge(ctx, request);
    } on TTransportError catch (e) {
      if (e.type != FrugalTTransportErrorType.REQUEST_TOO_LARGE) {
        throw new FTestError(e, FrugalTTransportErrorType.REQUEST_TOO_LARGE);
      }
    }
  }));

  tests.add(new FTest(1, 'testResponseTooLarge', () async {
    ctx = new FContext(correlationId: 'testResponseTooLarge');
    var request = new Uint8List(1);
    try {
      await client.testResponseTooLarge(ctx, request);
    } on TTransportError catch (e) {
      if (e.type != FrugalTTransportErrorType.RESPONSE_TOO_LARGE) {
        throw new FTestError(e, FrugalTTransportErrorType.RESPONSE_TOO_LARGE);
      }
    }
  }));

  return tests;
}

Future _testEnum(FFrugalTestClient client, int input) async {
  var result = await client.testEnum(ctx, input);
  if (result != input) throw new FTestError(result, input);
  if (!(result is Numberz)) throw new FTestError(result.runtimeType.toString(), "Numberz");
}
