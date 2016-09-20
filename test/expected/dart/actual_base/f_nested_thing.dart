// Autogenerated by Frugal Compiler (1.18.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

library actual_base_dart.src.f_nested_thing;

import 'dart:typed_data' show Uint8List;
import 'package:thrift/thrift.dart';
import 'package:actual_base_dart/actual_base_dart.dart' as t_actual_base_dart;

class nested_thing implements TBase {
  static final TStruct _STRUCT_DESC = new TStruct("nested_thing");
  static final TField _THINGS_FIELD_DESC = new TField("things", TType.LIST, 1);

  List<t_actual_base_dart.thing> _things;
  static const int THINGS = 1;


  nested_thing() {
  }

  List<t_actual_base_dart.thing> get things => this._things;

  set things(List<t_actual_base_dart.thing> things) {
    this._things = things;
  }

  bool isSetThings() => this.things != null;

  unsetThings() {
    this.things = null;
  }

  getFieldValue(int fieldID) {
    switch (fieldID) {
      case THINGS:
        return this.things;
      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  setFieldValue(int fieldID, Object value) {
    switch(fieldID) {
      case THINGS:
        if(value == null) {
          unsetThings();
        } else {
          this.things = value;
        }
        break;

      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  // Returns true if the field corresponding to fieldID is set (has been assigned a value) and false otherwise
  bool isSet(int fieldID) {
    switch(fieldID) {
      case THINGS:
        return isSetThings();
      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  read(TProtocol iprot) {
    TField field;
    iprot.readStructBegin();
    while(true) {
      field = iprot.readFieldBegin();
      if(field.type == TType.STOP) {
        break;
      }
      switch(field.id) {
        case THINGS:
          if(field.type == TType.LIST) {
            TList elem0 = iprot.readListBegin();
            things = new List<t_actual_base_dart.thing>();
            for(int elem2 = 0; elem2 < elem0.length; ++elem2) {
              t_actual_base_dart.thing elem1 = new t_actual_base_dart.thing();
              elem1.read(iprot);
              things.add(elem1);
            }
            iprot.readListEnd();
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        default:
          TProtocolUtil.skip(iprot, field.type);
          break;
      }
      iprot.readFieldEnd();
    }
    iprot.readStructEnd();

    // check for required fields of primitive type, which can't be checked in the validate method
    validate();
  }

  write(TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if(this.things != null) {
      oprot.writeFieldBegin(_THINGS_FIELD_DESC);
      oprot.writeListBegin(new TList(TType.STRUCT, things.length));
      for(var elem3 in things) {
        elem3.write(oprot);
      }
      oprot.writeListEnd();
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  String toString() {
    StringBuffer ret = new StringBuffer("nested_thing(");

    ret.write("things:");
    if(this.things == null) {
      ret.write("null");
    } else {
      ret.write(this.things);
    }

    ret.write(")");

    return ret.toString();
  }

  validate() {
    // check for required fields
    // check that fields of type enum have valid values
  }
}
