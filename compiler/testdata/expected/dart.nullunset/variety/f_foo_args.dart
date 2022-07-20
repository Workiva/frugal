// Autogenerated by Frugal Compiler (3.15.5)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:thrift/thrift.dart' as thrift;
import 'package:variety/variety.dart' as t_variety;
import 'package:actual_base_dart/actual_base_dart.dart' as t_actual_base_dart;
import 'package:intermediate_include/intermediate_include.dart' as t_intermediate_include;
import 'package:validStructs/validStructs.dart' as t_validStructs;
import 'package:ValidTypes/ValidTypes.dart' as t_ValidTypes;
import 'package:subdir_include_ns/subdir_include_ns.dart' as t_subdir_include_ns;

class FooArgs implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('FooArgs');
  static final thrift.TField _NEW_MESSAGE_FIELD_DESC = thrift.TField('newMessage', thrift.TType.STRING, 1);
  static final thrift.TField _MESSAGE_ARGS_FIELD_DESC = thrift.TField('messageArgs', thrift.TType.STRING, 2);
  static final thrift.TField _MESSAGE_RESULT_FIELD_DESC = thrift.TField('messageResult', thrift.TType.STRING, 3);

  String newMessage;
  static const int NEWMESSAGE = 1;
  String messageArgs;
  static const int MESSAGEARGS = 2;
  String messageResult;
  static const int MESSAGERESULT = 3;


  bool isSetNewMessage() => this.newMessage != null;

  unsetNewMessage() {
    this.newMessage = null;
  }

  bool isSetMessageArgs() => this.messageArgs != null;

  unsetMessageArgs() {
    this.messageArgs = null;
  }

  bool isSetMessageResult() => this.messageResult != null;

  unsetMessageResult() {
    this.messageResult = null;
  }

  @override
  getFieldValue(int fieldID) {
    switch (fieldID) {
      case NEWMESSAGE:
        return this.newMessage;
      case MESSAGEARGS:
        return this.messageArgs;
      case MESSAGERESULT:
        return this.messageResult;
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  @override
  setFieldValue(int fieldID, Object value) {
    switch (fieldID) {
      case NEWMESSAGE:
        if (value == null) {
          unsetNewMessage();
        } else {
          this.newMessage = value as String;
        }
        break;

      case MESSAGEARGS:
        if (value == null) {
          unsetMessageArgs();
        } else {
          this.messageArgs = value as String;
        }
        break;

      case MESSAGERESULT:
        if (value == null) {
          unsetMessageResult();
        } else {
          this.messageResult = value as String;
        }
        break;

      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  // Returns true if the field corresponding to fieldID is set (has been assigned a value) and false otherwise
  @override
  bool isSet(int fieldID) {
    switch (fieldID) {
      case NEWMESSAGE:
        return isSetNewMessage();
      case MESSAGEARGS:
        return isSetMessageArgs();
      case MESSAGERESULT:
        return isSetMessageResult();
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
        case NEWMESSAGE:
          if (field.type == thrift.TType.STRING) {
            this.newMessage = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case MESSAGEARGS:
          if (field.type == thrift.TType.STRING) {
            this.messageArgs = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case MESSAGERESULT:
          if (field.type == thrift.TType.STRING) {
            this.messageResult = iprot.readString();
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

  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (isSetNewMessage()) {
      oprot.writeFieldBegin(_NEW_MESSAGE_FIELD_DESC);
      oprot.writeString(this.newMessage);
      oprot.writeFieldEnd();
    }
    if (isSetMessageArgs()) {
      oprot.writeFieldBegin(_MESSAGE_ARGS_FIELD_DESC);
      oprot.writeString(this.messageArgs);
      oprot.writeFieldEnd();
    }
    if (isSetMessageResult()) {
      oprot.writeFieldBegin(_MESSAGE_RESULT_FIELD_DESC);
      oprot.writeString(this.messageResult);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  @override
  String toString() {
    StringBuffer ret = StringBuffer('FooArgs(');

    ret.write('newMessage:');
    if (this.newMessage == null) {
      ret.write('null');
    } else {
      ret.write(this.newMessage);
    }

    ret.write(', ');
    ret.write('messageArgs:');
    if (this.messageArgs == null) {
      ret.write('null');
    } else {
      ret.write(this.messageArgs);
    }

    ret.write(', ');
    ret.write('messageResult:');
    if (this.messageResult == null) {
      ret.write('null');
    } else {
      ret.write(this.messageResult);
    }

    ret.write(')');

    return ret.toString();
  }

  @override
  bool operator ==(Object o) {
    if (o is FooArgs) {
      return this.newMessage == o.newMessage &&
        this.messageArgs == o.messageArgs &&
        this.messageResult == o.messageResult;
    }
    return false;
  }

  @override
  int get hashCode {
    var value = 17;
    value = (value * 31) ^ this.newMessage.hashCode;
    value = (value * 31) ^ this.messageArgs.hashCode;
    value = (value * 31) ^ this.messageResult.hashCode;
    return value;
  }

  FooArgs clone({
    String newMessage,
    String messageArgs,
    String messageResult,
  }) {
    return FooArgs()
      ..newMessage = newMessage ?? this.newMessage
      ..messageArgs = messageArgs ?? this.messageArgs
      ..messageResult = messageResult ?? this.messageResult;
  }

  validate() {
  }
}
