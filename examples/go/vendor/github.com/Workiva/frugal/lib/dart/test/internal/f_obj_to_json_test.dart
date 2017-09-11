import "package:test/test.dart";
import "package:thrift/thrift.dart";

import "../../lib/src/frugal.dart";

void main() {
  group('fObjToJson', () {
    test('Serializes a TBase object', () {
      String json = fObjToJson(new _Edge()..label = "foo");
      expect(json, '{"1":{"str":"foo"}}');
    });

    test('Serializes an FContext', () {
      FContext ctx = new FContext(correlationId: "cid");
      String json = fObjToJson(ctx);
      var opId = ctx.requestHeader("_opid");
      expect(json, '{"_cid":"cid","_opid":"$opId","_timeout":"5000"}');
    });

    test('Serializes a normal object', () {
      String json = fObjToJson("foo");
      expect(json, '"foo"');
    });
  });
}

class _Edge implements TBase {
  static final TStruct _STRUCT_DESC = new TStruct("Edge");
  static final TField _LABEL_FIELD_DESC = new TField("label", TType.STRING, 1);

  String _label;
  static const int LABEL = 1;

  Edge() {}

  // label
  String get label => this._label;

  set label(String label) {
    this._label = label;
  }

  bool isSetLabel() => this.label != null;

  unsetLabel() {
    this.label = null;
  }

  getFieldValue(int fieldID) {
    switch (fieldID) {
      case LABEL:
        return this.label;
      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  setFieldValue(int fieldID, Object value) {
    switch (fieldID) {
      case LABEL:
        if (value == null) {
          unsetLabel();
        } else {
          this.label = value;
        }
        break;
      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  // Returns true if field corresponding to fieldID is set (has been assigned a value) and false otherwise
  bool isSet(int fieldID) {
    switch (fieldID) {
      case LABEL:
        return isSetLabel();
      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  read(TProtocol iprot) {
    TField field;
    iprot.readStructBegin();
    while (true) {
      field = iprot.readFieldBegin();
      if (field.type == TType.STOP) {
        break;
      }
      switch (field.id) {
        case LABEL:
          if (field.type == TType.STRING) {
            this.label = iprot.readString();
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
    if (this.label != null) {
      oprot.writeFieldBegin(_LABEL_FIELD_DESC);
      oprot.writeString(this.label);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  String toString() {
    StringBuffer ret = new StringBuffer("Edge(");

    ret.write("label:");
    if (this.label == null) {
      ret.write("null");
    } else {
      ret.write(this.label);
    }

    ret.write(")");

    return ret.toString();
  }

  validate() {
    // check for required fields
    // check that fields of type enum have valid values
  }
}
