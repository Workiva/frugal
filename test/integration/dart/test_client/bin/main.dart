/// Licensed to the Apache Software Foundation (ASF) under one
/// or more contributor license agreements. See the NOTICE file
/// distributed with this work for additional information
/// regarding copyright ownership. The ASF licenses this file
/// to you under the Apache License, Version 2.0 (the
/// 'License'); you may not use this file except in compliance
/// with the License. You may obtain a copy of the License at
///
/// http://www.apache.org/licenses/LICENSE-2.0
///
/// Unless required by applicable law or agreed to in writing,
/// software distributed under the License is distributed on an
/// 'AS IS' BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
/// KIND, either express or implied. See the License for the
/// specific language governing permissions and limitations
/// under the License.

import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:args/args.dart';
import 'package:collection/collection.dart';
import 'package:thrift/thrift.dart';
import 'package:frugal_test/frugal_test.dart';
import 'package:frugal/frugal.dart';
import 'package:w_transport/w_transport.dart' as wt;
import 'package:w_transport/w_transport_vm.dart' show configureWTransportForVM;

typedef Future FutureFunction();

class TTest {
  final int errorCode;
  final String name;
  final FutureFunction func;

  TTest(this.errorCode, this.name, this.func);
}

class TTestError extends Error {
  final actual;
  final expected;

  TTestError(this.actual, this.expected);

  String toString() => '\n\nUNEXPECTED ERROR \n$actual != \n$expected\n\n';
}

List<TTest> _tests;
FFrugalTestClient client;
bool verbose;
FContext ctx;

main(List<String> args) async {
  configureWTransportForVM();
  ArgResults results = _parseArgs(args);

  if (results == null) {
    exit(1);
  }

  verbose = results['verbose'] == true;

  await _initTestClient(
      host: results['host'],
      port: int.parse(results['port']),
      transportType: results['transport'],
      protocolType: results['protocol']).catchError((e) {
    stdout.writeln('Error:');
    stdout.writeln('$e');
    if (e is Error) {
      stdout.writeln('${e.stackTrace}');
    }
    exit(1);
  });

  // run tests
  int result = 0;
  _tests = _createTests();

  for (TTest test in _tests) {
    if (verbose) stdout.write('${test.name}... ');
    try {
      await test.func();
      if (verbose) stdout.writeln('success!');
    } catch (e) {
      stdout.writeln(e.toString());
      result = result | test.errorCode;
    }
  }

  exit(result);
}

ArgResults _parseArgs(List<String> args) {
  var parser = new ArgParser();
  parser.addOption('host', defaultsTo: 'localhost', help: 'The server host');
  parser.addOption('port', defaultsTo: '9090', help: 'The port to connect to');
  parser.addOption('transport',
      defaultsTo: 'http',
      allowed: ['http'],
      help: 'The transport name',
      allowedHelp: {
        'http': 'http transport'
      });
  parser.addOption('protocol',
      defaultsTo: 'binary',
      allowed: ['binary', 'compact', 'json'],
      help: 'The protocol name',
      allowedHelp: {
        'binary': 'TBinaryProtocol',
        'compact': 'TCompactProtocol',
        'json': 'TJsonProtocol'
      });
  parser.addFlag('verbose', defaultsTo: false);

  ArgResults results;
  try {
    results = parser.parse(args);
  } catch (e) {
    stdout.writeln('$e\n');
  }

  if (results == null) stdout.write(parser.usage);

  return results;
}

TProtocolFactory getProtocolFactory(String protocolType) {
  if (protocolType == 'binary') {
    return new TBinaryProtocolFactory();
  } else if (protocolType == 'compact') {
    return new TCompactProtocolFactory();
  } else if (protocolType == 'json') {
    return new TJsonProtocolFactory();
  }

  throw new ArgumentError.value(protocolType);
}

Future _initTestClient(
    {String host, int port, String transportType, String protocolType}) async {

  FProtocolFactory fProtocolFactory = null;
  FTransport fTransport = null;
  ctx = new FContext();

//  Nats is not available without the SDK in dart, so HTTP is the only transport we can test
  var uri = Uri.parse('http://$host:$port');
  var config = new FHttpConfig(uri, {});
  fTransport = new FHttpClientTransport(new wt.Client(), config);
  await fTransport.open();

  fProtocolFactory = new FProtocolFactory(getProtocolFactory(protocolType));
  client = new FFrugalTestClient(fTransport, fProtocolFactory);
}

List<TTest> _createTests() {
  List<TTest> tests = [];

  var xtruct = new Xtruct()
    ..string_thing = 'Zero'
    ..byte_thing = 1
    ..i32_thing = -3
    ..i64_thing = -5;

  tests.add(new TTest(1, 'testVoid', () async {
    stdout.write("testVoid()");
    await client.testVoid(ctx);
    stdout.write(" = void \n");
  }));

  tests.add(new TTest(1, 'testString', () async {
    var input = 'Test';
    stdout.write("testString(${input})");
    var result = await client.testString(ctx, input);
    if (result != input) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testBool', () async {
    var input = true;
    stdout.write("testBool(${input})");
    var result = await client.testBool(ctx, input);
    if (result != input) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testByte', () async {
    var input = 64;
    stdout.write("testByte(${input})");
    var result = await client.testByte(ctx, input);
    if (result != input) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testI32', () async {
    var input = 2147483647;
    stdout.write("testI32(${input})");
    var result = await client.testI32(ctx, input);
    if (result != input) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testI64', () async {
    var input = 9223372036854775807;
    stdout.write("testI64(${input})");
    var result = await client.testI64(ctx, input);
    if (result != input) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testDouble', () async {
    var input = 3.1415926;
    stdout.write("testDouble(${input})");
    var result = await client.testDouble(ctx, input);
    if (result != input) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testBinary', () async {
    var utf8Codec = const Utf8Codec();
    var input = utf8Codec.encode('foo');
    stdout.write("testBinary(${input})");
    var result = await client.testBinary(ctx, input);
    var equality = const ListEquality();
    if (!equality.equals(result, input)) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testStruct', () async {
    stdout.write("testStruct()");
    var result = await client.testStruct(ctx, xtruct);
    if ('$result' != '$xtruct') throw new TTestError(result, xtruct);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testNest', () async {
    var input = new Xtruct2()
      ..byte_thing = 1
      ..struct_thing = xtruct
      ..i32_thing = -3;

    stdout.write("testNest(${input})");
    var result = await client.testNest(ctx, input);
    if ('$result' != '$input') throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testMap', () async {
    Map<int, int> input = {1: -10, 2: -9, 3: -8, 4: -7, 5: -6};

    stdout.write("testMap(${input})");
    var result = await client.testMap(ctx, input);
    var equality = const MapEquality();
    if (!equality.equals(result, input)) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testSet', () async {
    var input = new Set.from([-2, -1, 0, 1, 2]);
    stdout.write("testSet(${input})");
    var result = await client.testSet(ctx, input);
    var equality = const SetEquality();
    if (!equality.equals(result, input)) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testList', () async {
    var input = [-2, -1, 0, 1, 2];
    stdout.write("testList(${input})");
    var result = await client.testList(ctx, input);
    var equality = const ListEquality();
    if (!equality.equals(result, input)) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testEnum', () async {
    await _testEnum(Numberz.ONE);
    await _testEnum(Numberz.TWO);
    await _testEnum(Numberz.THREE);
    await _testEnum(Numberz.FIVE);
    await _testEnum(Numberz.EIGHT);
  }));

  tests.add(new TTest(1, 'testTypedef', () async {
    var input = 309858235082523;
    stdout.write("testTypedef(${input})");
    var result = await client.testTypedef(ctx, input);
    if (result != input) throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testMapMap', () async {
    stdout.write("testMapMap(ctx, 1)");
    Map<int, Map<int, int>> result = await client.testMapMap(ctx, 1);
    if (result.isEmpty || result[result.keys.first].isEmpty) {
      throw new TTestError(result, 'Map<int, Map<int, int>>');
    }
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testInsanity', () async {
    var input = new Insanity();
    input.userMap = {Numberz.FIVE: 5000};
    input.xtructs = [xtruct];

    stdout.write("testInsanity(${input})");
    Map<int, Map<int, Insanity>> result = await client.testInsanity(ctx, input);
    if (result.isEmpty || result[1].isEmpty) {
      throw new TTestError(result, input);
    }
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testMulti', () async {
    var input = new Xtruct()
      ..string_thing = 'Hello2'
      ..byte_thing = 123
      ..i32_thing = 456
      ..i64_thing = 789;

    stdout.write("testMulti(${input})");
    var result = await client.testMulti(ctx, input.byte_thing, input.i32_thing,
        input.i64_thing, {1: 'one'}, Numberz.EIGHT, 5678);
    if ('$result' != '$input') throw new TTestError(result, input);
    stdout.write(" = ${result} \n");
  }));

  tests.add(new TTest(1, 'testException', () async {
    stdout.write("testException(Xception)");
    try {
      await client.testException(ctx, 'Xception');
    } on Xception catch (exception) {
      stdout.write(" = ${exception} \n");
      return;
    }

    throw new TTestError(null, 'Xception');
  }));

  tests.add(new TTest(1, 'testMultiException', () async {
    stdout.write("testMultiException(Xception2, foo)");
    try {
      await client.testMultiException(ctx, 'Xception2', 'foo');
    } on Xception2 catch (exception2) {
      stdout.write(" = ${exception2} \n");
      return;
    }

    throw new TTestError(null, 'Xception2');
  }));

  return tests;
}

Future _testEnum(int input) async {
  stdout.write("testEnum(${input})");
  var result = await client.testEnum(ctx, input);
  if (result != input) throw new TTestError(result, input);
  stdout.write(" = ${result} \n");
}
