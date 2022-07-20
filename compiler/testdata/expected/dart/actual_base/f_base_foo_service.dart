// Autogenerated by Frugal Compiler (3.15.5)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING



// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:logging/logging.dart' as logging;
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;
import 'package:w_common/disposable.dart' as disposable;

import 'package:actual_base_dart/actual_base_dart.dart' as t_actual_base_dart;


abstract class FBaseFoo {
  Future basePing(frugal.FContext ctx);
}

FBaseFooClient fBaseFooClientFactory(frugal.FServiceProvider provider, {List<frugal.Middleware> middleware}) =>
    FBaseFooClient(provider, middleware);

class FBaseFooClient extends disposable.Disposable implements FBaseFoo {
  static final logging.Logger _frugalLog = logging.Logger('BaseFoo');
  Map<String, frugal.FMethod> _methods;

  FBaseFooClient(frugal.FServiceProvider provider, [List<frugal.Middleware> middleware])
      : this._provider = provider {
    _transport = provider.transport;
    _protocolFactory = provider.protocolFactory;
    var combined = middleware ?? [];
    combined.addAll(provider.middleware);
    this._methods = {};
    this._methods['basePing'] = frugal.FMethod(this._basePing, 'BaseFoo', 'basePing', combined);
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

  @override
  Future basePing(frugal.FContext ctx) {
    return this._methods['basePing']([ctx]);
  }

  Future _basePing(frugal.FContext ctx) async {
    final args = basePing_args();
    final message = frugal.prepareMessage(ctx, 'basePing', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = basePing_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
  }
}

// ignore: camel_case_types
class basePing_args implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('basePing_args');



  @override
  getFieldValue(int fieldID) {
    switch (fieldID) {
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  @override
  setFieldValue(int fieldID, Object value) {
    switch (fieldID) {
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  // Returns true if the field corresponding to fieldID is set (has been assigned a value) and false otherwise
  @override
  bool isSet(int fieldID) {
    switch (fieldID) {
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

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

  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  @override
  String toString() {
    StringBuffer ret = StringBuffer('basePing_args(');

    ret.write(')');

    return ret.toString();
  }

  @override
  bool operator ==(Object o) {
    return o is basePing_args;
  }

  @override
  int get hashCode {
    var value = 17;
    return value;
  }

  basePing_args clone() {
    return basePing_args();
  }

  validate() {
  }
}
// ignore: camel_case_types
class basePing_result implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('basePing_result');



  @override
  getFieldValue(int fieldID) {
    switch (fieldID) {
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  @override
  setFieldValue(int fieldID, Object value) {
    switch (fieldID) {
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  // Returns true if the field corresponding to fieldID is set (has been assigned a value) and false otherwise
  @override
  bool isSet(int fieldID) {
    switch (fieldID) {
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

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

  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  @override
  String toString() {
    StringBuffer ret = StringBuffer('basePing_result(');

    ret.write(')');

    return ret.toString();
  }

  @override
  bool operator ==(Object o) {
    return o is basePing_result;
  }

  @override
  int get hashCode {
    var value = 17;
    return value;
  }

  basePing_result clone() {
    return basePing_result();
  }

  validate() {
  }
}
