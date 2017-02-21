// Autogenerated by Frugal Compiler (2.0.4)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

import 'dart:typed_data' show Uint8List;
import 'package:thrift/thrift.dart' as thrift;
import 'package:variety/variety.dart' as t_variety;
import 'package:actual_base_dart/actual_base_dart.dart' as t_actual_base_dart;
import 'package:validStructs/validStructs.dart' as t_validStructs;
import 'package:ValidTypes/ValidTypes.dart' as t_ValidTypes;
import 'package:subdir_include_ns/subdir_include_ns.dart' as t_subdir_include_ns;

class EventWrapper implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = new thrift.TStruct("EventWrapper");
  static final thrift.TField _ID_FIELD_DESC = new thrift.TField("ID", thrift.TType.I64, 1);
  static final thrift.TField _EV_FIELD_DESC = new thrift.TField("Ev", thrift.TType.STRUCT, 2);
  static final thrift.TField _EVENTS_FIELD_DESC = new thrift.TField("Events", thrift.TType.LIST, 3);
  static final thrift.TField _EVENTS2_FIELD_DESC = new thrift.TField("Events2", thrift.TType.SET, 4);
  static final thrift.TField _EVENT_MAP_FIELD_DESC = new thrift.TField("EventMap", thrift.TType.MAP, 5);
  static final thrift.TField _NUMS_FIELD_DESC = new thrift.TField("Nums", thrift.TType.LIST, 6);
  static final thrift.TField _ENUMS_FIELD_DESC = new thrift.TField("Enums", thrift.TType.LIST, 7);
  static final thrift.TField _A_BOOL_FIELD_FIELD_DESC = new thrift.TField("aBoolField", thrift.TType.BOOL, 8);
  static final thrift.TField _A_UNION_FIELD_DESC = new thrift.TField("a_union", thrift.TType.STRUCT, 9);
  static final thrift.TField _TYPEDEF_OF_TYPEDEF_FIELD_DESC = new thrift.TField("typedefOfTypedef", thrift.TType.STRING, 10);

  int _iD;
  static const int ID = 1;
  t_variety.Event _ev;
  static const int EV = 2;
  List<t_variety.Event> _events;
  static const int EVENTS = 3;
  Set<t_variety.Event> _events2;
  static const int EVENTS2 = 4;
  Map<int, t_variety.Event> _eventMap;
  static const int EVENTMAP = 5;
  List<List<int>> _nums;
  static const int NUMS = 6;
  List<int> _enums;
  static const int ENUMS = 7;
  bool _aBoolField = false;
  static const int ABOOLFIELD = 8;
  t_variety.TestingUnions _a_union;
  static const int A_UNION = 9;
  String _typedefOfTypedef;
  static const int TYPEDEFOFTYPEDEF = 10;

  bool __isset_iD = false;
  bool __isset_aBoolField = false;

  EventWrapper() {
  }

  int get iD => this._iD;

  set iD(int iD) {
    this._iD = iD;
    this.__isset_iD = true;
  }

  bool isSetID() => this.__isset_iD;

  unsetID() {
    this.__isset_iD = false;
  }

  t_variety.Event get ev => this._ev;

  set ev(t_variety.Event ev) {
    this._ev = ev;
  }

  bool isSetEv() => this.ev != null;

  unsetEv() {
    this.ev = null;
  }

  List<t_variety.Event> get events => this._events;

  set events(List<t_variety.Event> events) {
    this._events = events;
  }

  bool isSetEvents() => this.events != null;

  unsetEvents() {
    this.events = null;
  }

  Set<t_variety.Event> get events2 => this._events2;

  set events2(Set<t_variety.Event> events2) {
    this._events2 = events2;
  }

  bool isSetEvents2() => this.events2 != null;

  unsetEvents2() {
    this.events2 = null;
  }

  Map<int, t_variety.Event> get eventMap => this._eventMap;

  set eventMap(Map<int, t_variety.Event> eventMap) {
    this._eventMap = eventMap;
  }

  bool isSetEventMap() => this.eventMap != null;

  unsetEventMap() {
    this.eventMap = null;
  }

  List<List<int>> get nums => this._nums;

  set nums(List<List<int>> nums) {
    this._nums = nums;
  }

  bool isSetNums() => this.nums != null;

  unsetNums() {
    this.nums = null;
  }

  List<int> get enums => this._enums;

  set enums(List<int> enums) {
    this._enums = enums;
  }

  bool isSetEnums() => this.enums != null;

  unsetEnums() {
    this.enums = null;
  }

  bool get aBoolField => this._aBoolField;

  set aBoolField(bool aBoolField) {
    this._aBoolField = aBoolField;
    this.__isset_aBoolField = true;
  }

  bool isSetABoolField() => this.__isset_aBoolField;

  unsetABoolField() {
    this.__isset_aBoolField = false;
  }

  t_variety.TestingUnions get a_union => this._a_union;

  set a_union(t_variety.TestingUnions a_union) {
    this._a_union = a_union;
  }

  bool isSetA_union() => this.a_union != null;

  unsetA_union() {
    this.a_union = null;
  }

  String get typedefOfTypedef => this._typedefOfTypedef;

  set typedefOfTypedef(String typedefOfTypedef) {
    this._typedefOfTypedef = typedefOfTypedef;
  }

  bool isSetTypedefOfTypedef() => this.typedefOfTypedef != null;

  unsetTypedefOfTypedef() {
    this.typedefOfTypedef = null;
  }

  getFieldValue(int fieldID) {
    switch (fieldID) {
      case ID:
        return this.iD;
      case EV:
        return this.ev;
      case EVENTS:
        return this.events;
      case EVENTS2:
        return this.events2;
      case EVENTMAP:
        return this.eventMap;
      case NUMS:
        return this.nums;
      case ENUMS:
        return this.enums;
      case ABOOLFIELD:
        return this.aBoolField;
      case A_UNION:
        return this.a_union;
      case TYPEDEFOFTYPEDEF:
        return this.typedefOfTypedef;
      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  setFieldValue(int fieldID, Object value) {
    switch(fieldID) {
      case ID:
        if(value == null) {
          unsetID();
        } else {
          this.iD = value as int;
        }
        break;

      case EV:
        if(value == null) {
          unsetEv();
        } else {
          this.ev = value as t_variety.Event;
        }
        break;

      case EVENTS:
        if(value == null) {
          unsetEvents();
        } else {
          this.events = value as List<t_variety.Event>;
        }
        break;

      case EVENTS2:
        if(value == null) {
          unsetEvents2();
        } else {
          this.events2 = value as Set<t_variety.Event>;
        }
        break;

      case EVENTMAP:
        if(value == null) {
          unsetEventMap();
        } else {
          this.eventMap = value as Map<int, t_variety.Event>;
        }
        break;

      case NUMS:
        if(value == null) {
          unsetNums();
        } else {
          this.nums = value as List<List<int>>;
        }
        break;

      case ENUMS:
        if(value == null) {
          unsetEnums();
        } else {
          this.enums = value as List<int>;
        }
        break;

      case ABOOLFIELD:
        if(value == null) {
          unsetABoolField();
        } else {
          this.aBoolField = value as bool;
        }
        break;

      case A_UNION:
        if(value == null) {
          unsetA_union();
        } else {
          this.a_union = value as t_variety.TestingUnions;
        }
        break;

      case TYPEDEFOFTYPEDEF:
        if(value == null) {
          unsetTypedefOfTypedef();
        } else {
          this.typedefOfTypedef = value as String;
        }
        break;

      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  // Returns true if the field corresponding to fieldID is set (has been assigned a value) and false otherwise
  bool isSet(int fieldID) {
    switch(fieldID) {
      case ID:
        return isSetID();
      case EV:
        return isSetEv();
      case EVENTS:
        return isSetEvents();
      case EVENTS2:
        return isSetEvents2();
      case EVENTMAP:
        return isSetEventMap();
      case NUMS:
        return isSetNums();
      case ENUMS:
        return isSetEnums();
      case ABOOLFIELD:
        return isSetABoolField();
      case A_UNION:
        return isSetA_union();
      case TYPEDEFOFTYPEDEF:
        return isSetTypedefOfTypedef();
      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  read(thrift.TProtocol iprot) {
    thrift.TField field;
    iprot.readStructBegin();
    while(true) {
      field = iprot.readFieldBegin();
      if(field.type == thrift.TType.STOP) {
        break;
      }
      switch(field.id) {
        case ID:
          if(field.type == thrift.TType.I64) {
            iD = iprot.readI64();
            this.__isset_iD = true;
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case EV:
          if(field.type == thrift.TType.STRUCT) {
            ev = new t_variety.Event();
            ev.read(iprot);
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case EVENTS:
          if(field.type == thrift.TType.LIST) {
            thrift.TList elem21 = iprot.readListBegin();
            events = new List<t_variety.Event>();
            for(int elem23 = 0; elem23 < elem21.length; ++elem23) {
              t_variety.Event elem22 = new t_variety.Event();
              elem22.read(iprot);
              events.add(elem22);
            }
            iprot.readListEnd();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case EVENTS2:
          if(field.type == thrift.TType.SET) {
            thrift.TSet elem24 = iprot.readSetBegin();
            events2 = new Set<t_variety.Event>();
            for(int elem26 = 0; elem26 < elem24.length; ++elem26) {
              t_variety.Event elem25 = new t_variety.Event();
              elem25.read(iprot);
              events2.add(elem25);
            }
            iprot.readSetEnd();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case EVENTMAP:
          if(field.type == thrift.TType.MAP) {
            thrift.TMap elem27 = iprot.readMapBegin();
            eventMap = new Map<int, t_variety.Event>();
            for(int elem29 = 0; elem29 < elem27.length; ++elem29) {
              int elem30 = iprot.readI64();
              t_variety.Event elem28 = new t_variety.Event();
              elem28.read(iprot);
              eventMap[elem30] = elem28;
            }
            iprot.readMapEnd();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case NUMS:
          if(field.type == thrift.TType.LIST) {
            thrift.TList elem31 = iprot.readListBegin();
            nums = new List<List<int>>();
            for(int elem36 = 0; elem36 < elem31.length; ++elem36) {
              thrift.TList elem33 = iprot.readListBegin();
              List<int> elem32 = new List<int>();
              for(int elem35 = 0; elem35 < elem33.length; ++elem35) {
                int elem34 = iprot.readI32();
                elem32.add(elem34);
              }
              iprot.readListEnd();
              nums.add(elem32);
            }
            iprot.readListEnd();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case ENUMS:
          if(field.type == thrift.TType.LIST) {
            thrift.TList elem37 = iprot.readListBegin();
            enums = new List<int>();
            for(int elem39 = 0; elem39 < elem37.length; ++elem39) {
              int elem38 = iprot.readI32();
              enums.add(elem38);
            }
            iprot.readListEnd();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case ABOOLFIELD:
          if(field.type == thrift.TType.BOOL) {
            aBoolField = iprot.readBool();
            this.__isset_aBoolField = true;
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case A_UNION:
          if(field.type == thrift.TType.STRUCT) {
            a_union = new t_variety.TestingUnions();
            a_union.read(iprot);
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case TYPEDEFOFTYPEDEF:
          if(field.type == thrift.TType.STRING) {
            typedefOfTypedef = iprot.readString();
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

    // check for required fields of primitive type, which can't be checked in the validate method
    validate();
  }

  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if(isSetID()) {
      oprot.writeFieldBegin(_ID_FIELD_DESC);
      oprot.writeI64(iD);
      oprot.writeFieldEnd();
    }
    if(this.ev != null) {
      oprot.writeFieldBegin(_EV_FIELD_DESC);
      ev.write(oprot);
      oprot.writeFieldEnd();
    }
    if(this.events != null) {
      oprot.writeFieldBegin(_EVENTS_FIELD_DESC);
      oprot.writeListBegin(new thrift.TList(thrift.TType.STRUCT, events.length));
      for(var elem40 in events) {
        elem40.write(oprot);
      }
      oprot.writeListEnd();
      oprot.writeFieldEnd();
    }
    if(this.events2 != null) {
      oprot.writeFieldBegin(_EVENTS2_FIELD_DESC);
      oprot.writeSetBegin(new thrift.TSet(thrift.TType.STRUCT, events2.length));
      for(var elem41 in events2) {
        elem41.write(oprot);
      }
      oprot.writeSetEnd();
      oprot.writeFieldEnd();
    }
    if(this.eventMap != null) {
      oprot.writeFieldBegin(_EVENT_MAP_FIELD_DESC);
      oprot.writeMapBegin(new thrift.TMap(thrift.TType.I64, thrift.TType.STRUCT, eventMap.length));
      for(var elem42 in eventMap.keys) {
        oprot.writeI64(elem42);
        eventMap[elem42].write(oprot);
      }
      oprot.writeMapEnd();
      oprot.writeFieldEnd();
    }
    if(this.nums != null) {
      oprot.writeFieldBegin(_NUMS_FIELD_DESC);
      oprot.writeListBegin(new thrift.TList(thrift.TType.LIST, nums.length));
      for(var elem43 in nums) {
        oprot.writeListBegin(new thrift.TList(thrift.TType.I32, elem43.length));
        for(var elem44 in elem43) {
          oprot.writeI32(elem44);
        }
        oprot.writeListEnd();
      }
      oprot.writeListEnd();
      oprot.writeFieldEnd();
    }
    if(this.enums != null) {
      oprot.writeFieldBegin(_ENUMS_FIELD_DESC);
      oprot.writeListBegin(new thrift.TList(thrift.TType.I32, enums.length));
      for(var elem45 in enums) {
        oprot.writeI32(elem45);
      }
      oprot.writeListEnd();
      oprot.writeFieldEnd();
    }
    oprot.writeFieldBegin(_A_BOOL_FIELD_FIELD_DESC);
    oprot.writeBool(aBoolField);
    oprot.writeFieldEnd();
    if(this.a_union != null) {
      oprot.writeFieldBegin(_A_UNION_FIELD_DESC);
      a_union.write(oprot);
      oprot.writeFieldEnd();
    }
    if(this.typedefOfTypedef != null) {
      oprot.writeFieldBegin(_TYPEDEF_OF_TYPEDEF_FIELD_DESC);
      oprot.writeString(typedefOfTypedef);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  String toString() {
    StringBuffer ret = new StringBuffer("EventWrapper(");

    if(isSetID()) {
      ret.write("iD:");
      ret.write(this.iD);
    }

    ret.write(", ");
    ret.write("ev:");
    if(this.ev == null) {
      ret.write("null");
    } else {
      ret.write(this.ev);
    }

    ret.write(", ");
    ret.write("events:");
    if(this.events == null) {
      ret.write("null");
    } else {
      ret.write(this.events);
    }

    ret.write(", ");
    ret.write("events2:");
    if(this.events2 == null) {
      ret.write("null");
    } else {
      ret.write(this.events2);
    }

    ret.write(", ");
    ret.write("eventMap:");
    if(this.eventMap == null) {
      ret.write("null");
    } else {
      ret.write(this.eventMap);
    }

    ret.write(", ");
    ret.write("nums:");
    if(this.nums == null) {
      ret.write("null");
    } else {
      ret.write(this.nums);
    }

    ret.write(", ");
    ret.write("enums:");
    if(this.enums == null) {
      ret.write("null");
    } else {
      ret.write(this.enums);
    }

    ret.write(", ");
    ret.write("aBoolField:");
    ret.write(this.aBoolField);

    ret.write(", ");
    ret.write("a_union:");
    if(this.a_union == null) {
      ret.write("null");
    } else {
      ret.write(this.a_union);
    }

    ret.write(", ");
    ret.write("typedefOfTypedef:");
    if(this.typedefOfTypedef == null) {
      ret.write("null");
    } else {
      ret.write(this.typedefOfTypedef);
    }

    ret.write(")");

    return ret.toString();
  }

  validate() {
    // check for required fields
    if(ev == null) {
      throw new thrift.TProtocolError(thrift.TProtocolErrorType.INVALID_DATA, "Required field 'ev' was not present in struct EventWrapper");
    }
    // check that fields of type enum have valid values
  }
}
