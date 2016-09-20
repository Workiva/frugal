// Autogenerated by Frugal Compiler (1.18.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

library variety.src.f_event_wrapper;

import 'dart:typed_data' show Uint8List;
import 'package:thrift/thrift.dart';
import 'package:variety/variety.dart' as t_variety;
import 'package:actual_base_dart/actual_base_dart.dart' as t_actual_base_dart;

class EventWrapper implements TBase {
  static final TStruct _STRUCT_DESC = new TStruct("EventWrapper");
  static final TField _ID_FIELD_DESC = new TField("ID", TType.I64, 1);
  static final TField _EV_FIELD_DESC = new TField("Ev", TType.STRUCT, 2);
  static final TField _EVENTS_FIELD_DESC = new TField("Events", TType.LIST, 3);
  static final TField _EVENTS2_FIELD_DESC = new TField("Events2", TType.SET, 4);
  static final TField _EVENT_MAP_FIELD_DESC = new TField("EventMap", TType.MAP, 5);
  static final TField _NUMS_FIELD_DESC = new TField("Nums", TType.LIST, 6);
  static final TField _ENUMS_FIELD_DESC = new TField("Enums", TType.LIST, 7);
  static final TField _A_BOOL_FIELD_FIELD_DESC = new TField("aBoolField", TType.BOOL, 8);
  static final TField _A_UNION_FIELD_DESC = new TField("a_union", TType.STRUCT, 9);
  static final TField _TYPEDEF_OF_TYPEDEF_FIELD_DESC = new TField("typedefOfTypedef", TType.STRING, 10);

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
          this.iD = value;
        }
        break;

      case EV:
        if(value == null) {
          unsetEv();
        } else {
          this.ev = value;
        }
        break;

      case EVENTS:
        if(value == null) {
          unsetEvents();
        } else {
          this.events = value;
        }
        break;

      case EVENTS2:
        if(value == null) {
          unsetEvents2();
        } else {
          this.events2 = value;
        }
        break;

      case EVENTMAP:
        if(value == null) {
          unsetEventMap();
        } else {
          this.eventMap = value;
        }
        break;

      case NUMS:
        if(value == null) {
          unsetNums();
        } else {
          this.nums = value;
        }
        break;

      case ENUMS:
        if(value == null) {
          unsetEnums();
        } else {
          this.enums = value;
        }
        break;

      case ABOOLFIELD:
        if(value == null) {
          unsetABoolField();
        } else {
          this.aBoolField = value;
        }
        break;

      case A_UNION:
        if(value == null) {
          unsetA_union();
        } else {
          this.a_union = value;
        }
        break;

      case TYPEDEFOFTYPEDEF:
        if(value == null) {
          unsetTypedefOfTypedef();
        } else {
          this.typedefOfTypedef = value;
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

  read(TProtocol iprot) {
    TField field;
    iprot.readStructBegin();
    while(true) {
      field = iprot.readFieldBegin();
      if(field.type == TType.STOP) {
        break;
      }
      switch(field.id) {
        case ID:
          if(field.type == TType.I64) {
            iD = iprot.readI64();
            this.__isset_iD = true;
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case EV:
          if(field.type == TType.STRUCT) {
            ev = new t_variety.Event();
            ev.read(iprot);
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case EVENTS:
          if(field.type == TType.LIST) {
            TList elem25 = iprot.readListBegin();
            events = new List<t_variety.Event>();
            for(int elem27 = 0; elem27 < elem25.length; ++elem27) {
              t_variety.Event elem26 = new t_variety.Event();
              elem26.read(iprot);
              events.add(elem26);
            }
            iprot.readListEnd();
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case EVENTS2:
          if(field.type == TType.SET) {
            TSet elem28 = iprot.readSetBegin();
            events2 = new Set<t_variety.Event>();
            for(int elem30 = 0; elem30 < elem28.length; ++elem30) {
              t_variety.Event elem29 = new t_variety.Event();
              elem29.read(iprot);
              events2.add(elem29);
            }
            iprot.readSetEnd();
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case EVENTMAP:
          if(field.type == TType.MAP) {
            TMap elem31 = iprot.readMapBegin();
            eventMap = new Map<int, t_variety.Event>();
            for(int elem33 = 0; elem33 < elem31.length; ++elem33) {
              int elem34 = iprot.readI64();
              t_variety.Event elem32 = new t_variety.Event();
              elem32.read(iprot);
              eventMap[elem34] = elem32;
            }
            iprot.readMapEnd();
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case NUMS:
          if(field.type == TType.LIST) {
            TList elem35 = iprot.readListBegin();
            nums = new List<List<int>>();
            for(int elem40 = 0; elem40 < elem35.length; ++elem40) {
              TList elem37 = iprot.readListBegin();
              List<int> elem36 = new List<int>();
              for(int elem39 = 0; elem39 < elem37.length; ++elem39) {
                int elem38 = iprot.readI32();
                elem36.add(elem38);
              }
              iprot.readListEnd();
              nums.add(elem36);
            }
            iprot.readListEnd();
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case ENUMS:
          if(field.type == TType.LIST) {
            TList elem41 = iprot.readListBegin();
            enums = new List<int>();
            for(int elem43 = 0; elem43 < elem41.length; ++elem43) {
              int elem42 = iprot.readI32();
              enums.add(elem42);
            }
            iprot.readListEnd();
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case ABOOLFIELD:
          if(field.type == TType.BOOL) {
            aBoolField = iprot.readBool();
            this.__isset_aBoolField = true;
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case A_UNION:
          if(field.type == TType.STRUCT) {
            a_union = new t_variety.TestingUnions();
            a_union.read(iprot);
          } else {
            TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case TYPEDEFOFTYPEDEF:
          if(field.type == TType.STRING) {
            typedefOfTypedef = iprot.readString();
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
      oprot.writeListBegin(new TList(TType.STRUCT, events.length));
      for(var elem44 in events) {
        elem44.write(oprot);
      }
      oprot.writeListEnd();
      oprot.writeFieldEnd();
    }
    if(this.events2 != null) {
      oprot.writeFieldBegin(_EVENTS2_FIELD_DESC);
      oprot.writeSetBegin(new TSet(TType.STRUCT, events2.length));
      for(var elem45 in events2) {
        elem45.write(oprot);
      }
      oprot.writeSetEnd();
      oprot.writeFieldEnd();
    }
    if(this.eventMap != null) {
      oprot.writeFieldBegin(_EVENT_MAP_FIELD_DESC);
      oprot.writeMapBegin(new TMap(TType.I64, TType.STRUCT, eventMap.length));
      for(var elem46 in eventMap.keys) {
        oprot.writeI64(elem46);
        eventMap[elem46].write(oprot);
      }
      oprot.writeMapEnd();
      oprot.writeFieldEnd();
    }
    if(this.nums != null) {
      oprot.writeFieldBegin(_NUMS_FIELD_DESC);
      oprot.writeListBegin(new TList(TType.LIST, nums.length));
      for(var elem47 in nums) {
        oprot.writeListBegin(new TList(TType.I32, elem47.length));
        for(var elem48 in elem47) {
          oprot.writeI32(elem48);
        }
        oprot.writeListEnd();
      }
      oprot.writeListEnd();
      oprot.writeFieldEnd();
    }
    if(this.enums != null) {
      oprot.writeFieldBegin(_ENUMS_FIELD_DESC);
      oprot.writeListBegin(new TList(TType.I32, enums.length));
      for(var elem49 in enums) {
        oprot.writeI32(elem49);
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
      throw new TProtocolError(TProtocolErrorType.UNKNOWN, "Required field 'ev' was not present in struct EventWrapper");
    }
    // check that fields of type enum have valid values
  }
}
