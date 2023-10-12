// Autogenerated by Frugal Compiler (3.17.3)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:thrift/thrift.dart' as thrift;
import 'package:actual_base_dart/actual_base_dart.dart' as t_actual_base_dart;

class thing implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('thing');
  static final thrift.TField _AN_ID_FIELD_DESC = thrift.TField('an_id', thrift.TType.I32, 1);
  static final thrift.TField _A_STRING_FIELD_DESC = thrift.TField('a_string', thrift.TType.STRING, 2);

  int? _an_id = 0;
  static const int AN_ID = 1;
  String? _a_string;
  static const int A_STRING = 2;

  bool __isset_an_id = false;

  int? get an_id => this._an_id;

  set an_id(int? an_id) {
    this._an_id = an_id;
    this.__isset_an_id = true;
  }

  bool isSetAn_id() => this.__isset_an_id;

  unsetAn_id() {
    this.__isset_an_id = false;
  }

  String? get a_string => this._a_string;

  set a_string(String? a_string) {
    this._a_string = a_string;
  }

  bool isSetA_string() => this.a_string != null;

  unsetA_string() {
    this.a_string = null;
  }

  @override
  getFieldValue(int fieldID) {
    switch (fieldID) {
      case AN_ID:
        return this.an_id;
      case A_STRING:
        return this.a_string;
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  @override
  setFieldValue(int fieldID, Object? value) {
    switch (fieldID) {
      case AN_ID:
        if (value == null) {
          unsetAn_id();
        } else {
          this.an_id = value as int?;
        }
        break;

      case A_STRING:
        this.a_string = value as String?;
        break;

      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  // Returns true if the field corresponding to fieldID is set (has been assigned a value) and false otherwise
  @override
  bool isSet(int fieldID) {
    switch (fieldID) {
      case AN_ID:
        return isSetAn_id();
    }
    return getFieldValue(fieldID) != null;
  }

  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case AN_ID:
          if (field.type == thrift.TType.I32) {
            this.an_id = iprot.readI32();
            this.__isset_an_id = true;
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case A_STRING:
          if (field.type == thrift.TType.STRING) {
            this.a_string = iprot.readString();
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
    final elem180 = an_id!;
    oprot.writeFieldBegin(_AN_ID_FIELD_DESC);
    oprot.writeI32(elem180);
    oprot.writeFieldEnd();
    final elem181 = a_string;
    if (elem181 != null) {
      oprot.writeFieldBegin(_A_STRING_FIELD_DESC);
      oprot.writeString(elem181);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  @override
  String toString() {
    StringBuffer ret = StringBuffer('thing(');

    ret.write('an_id:');
    ret.write(this.an_id);

    ret.write(', ');
    ret.write('a_string:');
    if (this.a_string == null) {
      ret.write('null');
    } else {
      ret.write(this.a_string);
    }

    ret.write(')');

    return ret.toString();
  }

  @override
  bool operator ==(Object o) {
    if (o is thing) {
      return this.an_id == o.an_id &&
        this.a_string == o.a_string;
    }
    return false;
  }

  @override
  int get hashCode {
    var value = 17;
    value = (value * 31) ^ this.an_id.hashCode;
    value = (value * 31) ^ this.a_string.hashCode;
    return value;
  }

  thing clone({
    int? an_id,
    String? a_string,
  }) {
    return thing()
      ..an_id = an_id ?? this.an_id
      ..a_string = a_string ?? this.a_string;
  }

  validate() {
  }
}
