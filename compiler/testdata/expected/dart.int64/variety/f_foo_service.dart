// Autogenerated by Frugal Compiler (3.16.26)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING



// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:fixnum/fixnum.dart' as fixnum;
import 'package:logging/logging.dart' as logging;
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;
import 'package:w_common/disposable.dart' as disposable;

import 'package:actual_base_dart/actual_base_dart.dart' as t_actual_base_dart;
import 'package:validStructs/validStructs.dart' as t_validStructs;
import 'package:ValidTypes/ValidTypes.dart' as t_ValidTypes;
import 'package:subdir_include_ns/subdir_include_ns.dart' as t_subdir_include_ns;
import 'package:variety/variety.dart' as t_variety;


/// This is a thrift service. Frugal will generate bindings that include
/// a frugal Context for each service call.
abstract class FFoo extends t_actual_base_dart.FBaseFoo {
  /// Ping the server.
  /// Deprecated: don't use this; use "something else"
  @deprecated
  Future ping(frugal.FContext ctx);

  /// Blah the server.
  Future<fixnum.Int64> blah(frugal.FContext ctx, int num, String str, t_variety.Event event);

  /// oneway methods don't receive a response from the server.
  Future oneWay(frugal.FContext ctx, fixnum.Int64 id, Map<int, String> req);

  Future<Uint8List> bin_method(frugal.FContext ctx, Uint8List bin, String str);

  Future<fixnum.Int64> param_modifiers(frugal.FContext ctx, int opt_num, int default_num, int req_num);

  Future<List<fixnum.Int64>> underlying_types_test(frugal.FContext ctx, List<fixnum.Int64> list_type, Set<fixnum.Int64> set_type);

  Future<t_validStructs.Thing> getThing(frugal.FContext ctx);

  Future<int> getMyInt(frugal.FContext ctx);

  Future<t_subdir_include_ns.A> use_subdir_struct(frugal.FContext ctx, t_subdir_include_ns.A a);

  Future<String> sayHelloWith(frugal.FContext ctx, String newMessage);

  Future<String> whatDoYouSay(frugal.FContext ctx, String messageArgs);

  Future<String> sayAgain(frugal.FContext ctx, String messageResult);
}

FFooClient fFooClientFactory(frugal.FServiceProvider provider, {List<frugal.Middleware> middleware}) =>
    FFooClient(provider, middleware);

/// This is a thrift service. Frugal will generate bindings that include
/// a frugal Context for each service call.
// The below ignore statement is only needed to workaround https://github.com/dart-lang/sdk/issues/29751, which is fixed on Dart 2.8.0 and later.
// Dart versions before 2.8.0 need this ignore to analyze properly.
// ignore: private_collision_in_mixin_application
class FFooClient extends t_actual_base_dart.FBaseFooClient with disposable.Disposable implements FFoo {
  static final logging.Logger _frugalLog = logging.Logger('Foo');
  Map<String, frugal.FMethod> _methods = {};

  FFooClient(frugal.FServiceProvider provider, [List<frugal.Middleware> middleware])
      : this._provider = provider,
        super(provider, middleware) {
    _transport = provider.transport;
    _protocolFactory = provider.protocolFactory;
    var combined = middleware ?? [];
    combined.addAll(provider.middleware);
    this._methods['ping'] = frugal.FMethod(this._ping, 'Foo', 'ping', combined);
    this._methods['blah'] = frugal.FMethod(this._blah, 'Foo', 'blah', combined);
    this._methods['oneWay'] = frugal.FMethod(this._oneWay, 'Foo', 'oneWay', combined);
    this._methods['bin_method'] = frugal.FMethod(this._bin_method, 'Foo', 'bin_method', combined);
    this._methods['param_modifiers'] = frugal.FMethod(this._param_modifiers, 'Foo', 'param_modifiers', combined);
    this._methods['underlying_types_test'] = frugal.FMethod(this._underlying_types_test, 'Foo', 'underlying_types_test', combined);
    this._methods['getThing'] = frugal.FMethod(this._getThing, 'Foo', 'getThing', combined);
    this._methods['getMyInt'] = frugal.FMethod(this._getMyInt, 'Foo', 'getMyInt', combined);
    this._methods['use_subdir_struct'] = frugal.FMethod(this._use_subdir_struct, 'Foo', 'use_subdir_struct', combined);
    this._methods['sayHelloWith'] = frugal.FMethod(this._sayHelloWith, 'Foo', 'sayHelloWith', combined);
    this._methods['whatDoYouSay'] = frugal.FMethod(this._whatDoYouSay, 'Foo', 'whatDoYouSay', combined);
    this._methods['sayAgain'] = frugal.FMethod(this._sayAgain, 'Foo', 'sayAgain', combined);
  }

  frugal.FServiceProvider _provider;
  frugal.FTransport _transport;
  frugal.FProtocolFactory _protocolFactory;

  @override
  Future<Null> onDispose() async {
    if (_provider is disposable.Disposable && !_provider.isOrWillBeDisposed)  {
      return _provider.dispose();
    }
    return null;
  }

  /// Ping the server.
  /// Deprecated: don't use this; use "something else"
  @deprecated
  @override
  Future ping(frugal.FContext ctx) {
    _frugalLog.warning("Call to deprecated function 'Foo.ping'");
    return this._methods['ping']([ctx]);
  }

  Future _ping(frugal.FContext ctx) async {
    final args = Ping_args();
    final message = frugal.prepareMessage(ctx, 'ping', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = Ping_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
  }
  /// Blah the server.
  @override
  Future<fixnum.Int64> blah(frugal.FContext ctx, int num, String str, t_variety.Event event) {
    return this._methods['blah']([ctx, num, str, event]).then((value) => value as fixnum.Int64);
  }

  Future<fixnum.Int64> _blah(frugal.FContext ctx, int num, String str, t_variety.Event event) async {
    final args = blah_args();
    args.num = num;
    args.str = str;
    args.event = event;
    final message = frugal.prepareMessage(ctx, 'blah', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = blah_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    if (result.awe != null) {
      throw result.awe;
    }
    if (result.api != null) {
      throw result.api;
    }
    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'blah failed: unknown result'
    );
  }
  /// oneway methods don't receive a response from the server.
  @override
  Future oneWay(frugal.FContext ctx, fixnum.Int64 id, Map<int, String> req) {
    return this._methods['oneWay']([ctx, id, req]);
  }

  Future _oneWay(frugal.FContext ctx, fixnum.Int64 id, Map<int, String> req) async {
    final args = oneWay_args();
    args.id = id;
    args.req = req;
    final message = frugal.prepareMessage(ctx, 'oneWay', args, thrift.TMessageType.ONEWAY, _protocolFactory, _transport.requestSizeLimit);
    await _transport.oneway(ctx, message);
  }

  @override
  Future<Uint8List> bin_method(frugal.FContext ctx, Uint8List bin, String str) {
    return this._methods['bin_method']([ctx, bin, str]).then((value) => value as Uint8List);
  }

  Future<Uint8List> _bin_method(frugal.FContext ctx, Uint8List bin, String str) async {
    final args = bin_method_args();
    args.bin = bin;
    args.str = str;
    final message = frugal.prepareMessage(ctx, 'bin_method', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = bin_method_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    if (result.api != null) {
      throw result.api;
    }
    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'bin_method failed: unknown result'
    );
  }
  @override
  Future<fixnum.Int64> param_modifiers(frugal.FContext ctx, int opt_num, int default_num, int req_num) {
    return this._methods['param_modifiers']([ctx, opt_num, default_num, req_num]).then((value) => value as fixnum.Int64);
  }

  Future<fixnum.Int64> _param_modifiers(frugal.FContext ctx, int opt_num, int default_num, int req_num) async {
    final args = param_modifiers_args();
    args.opt_num = opt_num;
    args.default_num = default_num;
    args.req_num = req_num;
    final message = frugal.prepareMessage(ctx, 'param_modifiers', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = param_modifiers_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'param_modifiers failed: unknown result'
    );
  }
  @override
  Future<List<fixnum.Int64>> underlying_types_test(frugal.FContext ctx, List<fixnum.Int64> list_type, Set<fixnum.Int64> set_type) {
    return this._methods['underlying_types_test']([ctx, list_type, set_type]).then((value) => value as List<fixnum.Int64>);
  }

  Future<List<fixnum.Int64>> _underlying_types_test(frugal.FContext ctx, List<fixnum.Int64> list_type, Set<fixnum.Int64> set_type) async {
    final args = underlying_types_test_args();
    args.list_type = list_type;
    args.set_type = set_type;
    final message = frugal.prepareMessage(ctx, 'underlying_types_test', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = underlying_types_test_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'underlying_types_test failed: unknown result'
    );
  }
  @override
  Future<t_validStructs.Thing> getThing(frugal.FContext ctx) {
    return this._methods['getThing']([ctx]).then((value) => value as t_validStructs.Thing);
  }

  Future<t_validStructs.Thing> _getThing(frugal.FContext ctx) async {
    final args = getThing_args();
    final message = frugal.prepareMessage(ctx, 'getThing', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = getThing_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'getThing failed: unknown result'
    );
  }
  @override
  Future<int> getMyInt(frugal.FContext ctx) {
    return this._methods['getMyInt']([ctx]).then((value) => value as int);
  }

  Future<int> _getMyInt(frugal.FContext ctx) async {
    final args = getMyInt_args();
    final message = frugal.prepareMessage(ctx, 'getMyInt', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = getMyInt_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'getMyInt failed: unknown result'
    );
  }
  @override
  Future<t_subdir_include_ns.A> use_subdir_struct(frugal.FContext ctx, t_subdir_include_ns.A a) {
    return this._methods['use_subdir_struct']([ctx, a]).then((value) => value as t_subdir_include_ns.A);
  }

  Future<t_subdir_include_ns.A> _use_subdir_struct(frugal.FContext ctx, t_subdir_include_ns.A a) async {
    final args = use_subdir_struct_args();
    args.a = a;
    final message = frugal.prepareMessage(ctx, 'use_subdir_struct', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = use_subdir_struct_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'use_subdir_struct failed: unknown result'
    );
  }
  @override
  Future<String> sayHelloWith(frugal.FContext ctx, String newMessage) {
    return this._methods['sayHelloWith']([ctx, newMessage]).then((value) => value as String);
  }

  Future<String> _sayHelloWith(frugal.FContext ctx, String newMessage) async {
    final args = sayHelloWith_args();
    args.newMessage = newMessage;
    final message = frugal.prepareMessage(ctx, 'sayHelloWith', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = sayHelloWith_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'sayHelloWith failed: unknown result'
    );
  }
  @override
  Future<String> whatDoYouSay(frugal.FContext ctx, String messageArgs) {
    return this._methods['whatDoYouSay']([ctx, messageArgs]).then((value) => value as String);
  }

  Future<String> _whatDoYouSay(frugal.FContext ctx, String messageArgs) async {
    final args = whatDoYouSay_args();
    args.messageArgs = messageArgs;
    final message = frugal.prepareMessage(ctx, 'whatDoYouSay', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = whatDoYouSay_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'whatDoYouSay failed: unknown result'
    );
  }
  @override
  Future<String> sayAgain(frugal.FContext ctx, String messageResult) {
    return this._methods['sayAgain']([ctx, messageResult]).then((value) => value as String);
  }

  Future<String> _sayAgain(frugal.FContext ctx, String messageResult) async {
    final args = sayAgain_args();
    args.messageResult = messageResult;
    final message = frugal.prepareMessage(ctx, 'sayAgain', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = sayAgain_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'sayAgain failed: unknown result'
    );
  }
}

// ignore: camel_case_types
class Ping_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('Ping_args');



  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class Ping_result extends frugal.FGeneratedArgsResultBase {


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class blah_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('blah_args');
  static final thrift.TField _NUM_FIELD_DESC = thrift.TField('num', thrift.TType.I32, 1);
  static final thrift.TField _STR_FIELD_DESC = thrift.TField('Str', thrift.TType.STRING, 2);
  static final thrift.TField _EVENT_FIELD_DESC = thrift.TField('event', thrift.TType.STRUCT, 3);

  int num;
  String str;
  t_variety.Event event;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    oprot.writeFieldBegin(_NUM_FIELD_DESC);
    oprot.writeI32(this.num);
    oprot.writeFieldEnd();
    if (this.str != null) {
      oprot.writeFieldBegin(_STR_FIELD_DESC);
      oprot.writeString(this.str);
      oprot.writeFieldEnd();
    }
    if (this.event != null) {
      oprot.writeFieldBegin(_EVENT_FIELD_DESC);
      this.event.write(oprot);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class blah_result extends frugal.FGeneratedArgsResultBase {
  fixnum.Int64 success;
  t_variety.AwesomeException awe;
  t_actual_base_dart.api_exception api;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.I64) {
            this.success = iprot.readInt64();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case 1:
          if (field.type == thrift.TType.STRUCT) {
            this.awe = t_variety.AwesomeException();
            awe.read(iprot);
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case 2:
          if (field.type == thrift.TType.STRUCT) {
            this.api = t_actual_base_dart.api_exception();
            api.read(iprot);
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class oneWay_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('oneWay_args');
  static final thrift.TField _ID_FIELD_DESC = thrift.TField('id', thrift.TType.I64, 1);
  static final thrift.TField _REQ_FIELD_DESC = thrift.TField('req', thrift.TType.MAP, 2);

  fixnum.Int64 id;
  Map<int, String> req;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    oprot.writeFieldBegin(_ID_FIELD_DESC);
    oprot.writeInt64(this.id);
    oprot.writeFieldEnd();
    if (this.req != null) {
      oprot.writeFieldBegin(_REQ_FIELD_DESC);
      final temp = this.req;
      oprot.writeMapBegin(thrift.TMap(thrift.TType.I32, thrift.TType.STRING, temp.length));
      for(var elem83 in temp.keys) {
        oprot.writeI32(elem83);
        oprot.writeString(temp[elem83]);
      }
      oprot.writeMapEnd();
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class bin_method_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('bin_method_args');
  static final thrift.TField _BIN_FIELD_DESC = thrift.TField('bin', thrift.TType.STRING, 1);
  static final thrift.TField _STR_FIELD_DESC = thrift.TField('Str', thrift.TType.STRING, 2);

  Uint8List bin;
  String str;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (this.bin != null) {
      oprot.writeFieldBegin(_BIN_FIELD_DESC);
      oprot.writeBinary(this.bin);
      oprot.writeFieldEnd();
    }
    if (this.str != null) {
      oprot.writeFieldBegin(_STR_FIELD_DESC);
      oprot.writeString(this.str);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class bin_method_result extends frugal.FGeneratedArgsResultBase {
  Uint8List success;
  t_actual_base_dart.api_exception api;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.STRING) {
            this.success = iprot.readBinary();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case 1:
          if (field.type == thrift.TType.STRUCT) {
            this.api = t_actual_base_dart.api_exception();
            api.read(iprot);
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class param_modifiers_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('param_modifiers_args');
  static final thrift.TField _OPT_NUM_FIELD_DESC = thrift.TField('opt_num', thrift.TType.I32, 1);
  static final thrift.TField _DEFAULT_NUM_FIELD_DESC = thrift.TField('default_num', thrift.TType.I32, 2);
  static final thrift.TField _REQ_NUM_FIELD_DESC = thrift.TField('req_num', thrift.TType.I32, 3);

  int opt_num;
  int default_num;
  int req_num;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    oprot.writeFieldBegin(_OPT_NUM_FIELD_DESC);
    oprot.writeI32(this.opt_num);
    oprot.writeFieldEnd();
    oprot.writeFieldBegin(_DEFAULT_NUM_FIELD_DESC);
    oprot.writeI32(this.default_num);
    oprot.writeFieldEnd();
    oprot.writeFieldBegin(_REQ_NUM_FIELD_DESC);
    oprot.writeI32(this.req_num);
    oprot.writeFieldEnd();
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
    // check for required fields
    if (this.req_num == null) {
      throw thrift.TProtocolError(thrift.TProtocolErrorType.INVALID_DATA, "Required field 'req_num' was not present in struct param_modifiers_args");
    }
  }
}
// ignore: camel_case_types
class param_modifiers_result extends frugal.FGeneratedArgsResultBase {
  fixnum.Int64 success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.I64) {
            this.success = iprot.readInt64();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class underlying_types_test_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('underlying_types_test_args');
  static final thrift.TField _LIST_TYPE_FIELD_DESC = thrift.TField('list_type', thrift.TType.LIST, 1);
  static final thrift.TField _SET_TYPE_FIELD_DESC = thrift.TField('set_type', thrift.TType.SET, 2);

  List<fixnum.Int64> list_type;
  Set<fixnum.Int64> set_type;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (this.list_type != null) {
      oprot.writeFieldBegin(_LIST_TYPE_FIELD_DESC);
      final temp = this.list_type;
      oprot.writeListBegin(thrift.TList(thrift.TType.I64, temp.length));
      for(var elem84 in temp) {
        oprot.writeInt64(elem84);
      }
      oprot.writeListEnd();
      oprot.writeFieldEnd();
    }
    if (this.set_type != null) {
      oprot.writeFieldBegin(_SET_TYPE_FIELD_DESC);
      final temp = this.set_type;
      oprot.writeSetBegin(thrift.TSet(thrift.TType.I64, temp.length));
      for(var elem85 in temp) {
        oprot.writeInt64(elem85);
      }
      oprot.writeSetEnd();
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class underlying_types_test_result extends frugal.FGeneratedArgsResultBase {
  List<fixnum.Int64> success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.LIST) {
            thrift.TList elem86 = iprot.readListBegin();
            final elem89 = <fixnum.Int64>[];
            for(int elem88 = 0; elem88 < elem86.length; ++elem88) {
              fixnum.Int64 elem87 = iprot.readInt64();
              elem89.add(elem87);
            }
            iprot.readListEnd();
            this.success = elem89;
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class getThing_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('getThing_args');



  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class getThing_result extends frugal.FGeneratedArgsResultBase {
  t_validStructs.Thing success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.STRUCT) {
            this.success = t_validStructs.Thing();
            success.read(iprot);
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class getMyInt_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('getMyInt_args');



  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class getMyInt_result extends frugal.FGeneratedArgsResultBase {
  int success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.I32) {
            this.success = iprot.readI32();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class use_subdir_struct_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('use_subdir_struct_args');
  static final thrift.TField _A_FIELD_DESC = thrift.TField('a', thrift.TType.STRUCT, 1);

  t_subdir_include_ns.A a;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (this.a != null) {
      oprot.writeFieldBegin(_A_FIELD_DESC);
      this.a.write(oprot);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class use_subdir_struct_result extends frugal.FGeneratedArgsResultBase {
  t_subdir_include_ns.A success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.STRUCT) {
            this.success = t_subdir_include_ns.A();
            success.read(iprot);
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class sayHelloWith_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('sayHelloWith_args');
  static final thrift.TField _NEW_MESSAGE_FIELD_DESC = thrift.TField('newMessage', thrift.TType.STRING, 1);

  String newMessage;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (this.newMessage != null) {
      oprot.writeFieldBegin(_NEW_MESSAGE_FIELD_DESC);
      oprot.writeString(this.newMessage);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class sayHelloWith_result extends frugal.FGeneratedArgsResultBase {
  String success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.STRING) {
            this.success = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class whatDoYouSay_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('whatDoYouSay_args');
  static final thrift.TField _MESSAGE_ARGS_FIELD_DESC = thrift.TField('messageArgs', thrift.TType.STRING, 1);

  String messageArgs;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (this.messageArgs != null) {
      oprot.writeFieldBegin(_MESSAGE_ARGS_FIELD_DESC);
      oprot.writeString(this.messageArgs);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class whatDoYouSay_result extends frugal.FGeneratedArgsResultBase {
  String success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.STRING) {
            this.success = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
// ignore: camel_case_types
class sayAgain_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('sayAgain_args');
  static final thrift.TField _MESSAGE_RESULT_FIELD_DESC = thrift.TField('messageResult', thrift.TType.STRING, 1);

  String messageResult;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (this.messageResult != null) {
      oprot.writeFieldBegin(_MESSAGE_RESULT_FIELD_DESC);
      oprot.writeString(this.messageResult);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class sayAgain_result extends frugal.FGeneratedArgsResultBase {
  String success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.STRING) {
            this.success = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          thrift.TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    validate();
  }

  validate() {
  }
}
