// Autogenerated by Frugal Compiler (3.16.3)
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

class AwesomeException extends Error implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('AwesomeException');
  static final thrift.TField _ID_FIELD_DESC = thrift.TField('ID', thrift.TType.I64, 1);
  static final thrift.TField _REASON_FIELD_DESC = thrift.TField('Reason', thrift.TType.STRING, 2);
  static final thrift.TField _DEPR_FIELD_DESC = thrift.TField('depr', thrift.TType.BOOL, 3);

  /// ID is a unique identifier for an awesome exception.
  int iD;
  static const int ID = 1;
  /// Reason contains the error message.
  String reason;
  static const int REASON = 2;
  /// Deprecated: use something else
  @deprecated
  bool depr;
  static const int DEPR = 3;


  bool isSetID() => this.iD != null;

  unsetID() {
    this.iD = null;
  }

  bool isSetReason() => this.reason != null;

  unsetReason() {
    this.reason = null;
  }

  @deprecated  bool isSetDepr() => this.depr != null;

  unsetDepr() {
    // ignore: deprecated_member_use
    this.depr = null;
  }

  @override
  getFieldValue(int fieldID) {
    switch (fieldID) {
      case ID:
        return this.iD;
      case REASON:
        return this.reason;
      case DEPR:
        // ignore: deprecated_member_use
        return this.depr;
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  @override
  setFieldValue(int fieldID, Object value) {
    switch (fieldID) {
      case ID:
        if (value == null) {
          unsetID();
        } else {
          this.iD = value as int;
        }
        break;

      case REASON:
        if (value == null) {
          unsetReason();
        } else {
          this.reason = value as String;
        }
        break;

      case DEPR:
        if (value == null) {
          unsetDepr();
        } else {
          // ignore: deprecated_member_use
          this.depr = value as bool;
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
      case ID:
        return isSetID();
      case REASON:
        return isSetReason();
      case DEPR:
        // ignore: deprecated_member_use
        return isSetDepr();
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
        case ID:
          if (field.type == thrift.TType.I64) {
            this.iD = iprot.readI64();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case REASON:
          if (field.type == thrift.TType.STRING) {
            this.reason = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case DEPR:
          if (field.type == thrift.TType.BOOL) {
            // ignore: deprecated_member_use
            this.depr = iprot.readBool();
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
    if (isSetID()) {
      oprot.writeFieldBegin(_ID_FIELD_DESC);
      oprot.writeI64(this.iD);
      oprot.writeFieldEnd();
    }
    if (isSetReason()) {
      oprot.writeFieldBegin(_REASON_FIELD_DESC);
      oprot.writeString(this.reason);
      oprot.writeFieldEnd();
    }
    // ignore: deprecated_member_use
    if (isSetDepr()) {
      oprot.writeFieldBegin(_DEPR_FIELD_DESC);
      // ignore: deprecated_member_use
      oprot.writeBool(this.depr);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  @override
  String toString() {
    StringBuffer ret = StringBuffer('AwesomeException(');

    ret.write('iD:');
    ret.write(this.iD);

    ret.write(', ');
    ret.write('reason:');
    if (this.reason == null) {
      ret.write('null');
    } else {
      ret.write(this.reason);
    }

    ret.write(', ');
    ret.write('depr:');
    // ignore: deprecated_member_use
    ret.write(this.depr);

    ret.write(')');

    return ret.toString();
  }

  @override
  bool operator ==(Object o) {
    if (o is AwesomeException) {
      return this.iD == o.iD &&
        this.reason == o.reason &&
        // ignore: deprecated_member_use
        this.depr == o.depr;
    }
    return false;
  }

  @override
  int get hashCode {
    var value = 17;
    value = (value * 31) ^ this.iD.hashCode;
    value = (value * 31) ^ this.reason.hashCode;
    // ignore: deprecated_member_use
    value = (value * 31) ^ this.depr.hashCode;
    return value;
  }

  AwesomeException clone({
    int iD,
    String reason,
    // ignore: deprecated_member_use
    bool depr,
  }) {
    return AwesomeException()
      ..iD = iD ?? this.iD
      ..reason = reason ?? this.reason
      // ignore: deprecated_member_use
      ..depr = depr ?? this.depr;
  }

  validate() {
  }
}
