// Autogenerated by Frugal Compiler (10.0.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal_test/frugal_test.dart' as t_frugal_test;

class Bonk implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('Bonk');
  static final thrift.TField _MESSAGE_FIELD_DESC = thrift.TField('message', thrift.TType.STRING, 1);
  static final thrift.TField _TYPE_FIELD_DESC = thrift.TField('type', thrift.TType.I32, 2);

  String _message;
  static const int MESSAGE = 1;
  int _type = 0;
  static const int TYPE = 2;

  bool __isset_type = false;

  String get message => this._message;

  set message(String message) {
    this._message = message;
  }

  bool isSetMessage() => this.message != null;

  unsetMessage() {
    this.message = null;
  }

  int get type => this._type;

  set type(int type) {
    this._type = type;
    this.__isset_type = true;
  }

  bool isSetType() => this.__isset_type;

  unsetType() {
    this.__isset_type = false;
  }

  @override
  getFieldValue(int fieldID) {
    switch (fieldID) {
      case MESSAGE:
        return this.message;
      case TYPE:
        return this.type;
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  @override
  setFieldValue(int fieldID, Object value) {
    switch (fieldID) {
      case MESSAGE:
        if (value == null) {
          unsetMessage();
        } else {
          this.message = value as String;
        }
        break;

      case TYPE:
        if (value == null) {
          unsetType();
        } else {
          this.type = value as int;
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
      case MESSAGE:
        return isSetMessage();
      case TYPE:
        return isSetType();
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
        case MESSAGE:
          if (field.type == thrift.TType.STRING) {
            this.message = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case TYPE:
          if (field.type == thrift.TType.I32) {
            this.type = iprot.readI32();
            this.__isset_type = true;
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
    if (this.message != null) {
      oprot.writeFieldBegin(_MESSAGE_FIELD_DESC);
      oprot.writeString(this.message);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldBegin(_TYPE_FIELD_DESC);
    oprot.writeI32(this.type);
    oprot.writeFieldEnd();
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  @override
  String toString() {
    StringBuffer ret = StringBuffer('Bonk(');

    ret.write('message:');
    if (this.message == null) {
      ret.write('null');
    } else {
      ret.write(this.message);
    }

    ret.write(', ');
    ret.write('type:');
    ret.write(this.type);

    ret.write(')');

    return ret.toString();
  }

  @override
  bool operator ==(Object o) {
    if (o is Bonk) {
      return this.message == o.message &&
        this.type == o.type;
    }
    return false;
  }

  @override
  int get hashCode {
    var value = 17;
    value = (value * 31) ^ this.message.hashCode;
    value = (value * 31) ^ this.type.hashCode;
    return value;
  }

  Bonk clone({
    String message,
    int type,
  }) {
    return Bonk()
      ..message = message ?? this.message
      ..type = type ?? this.type;
  }

  validate() {
  }
}
