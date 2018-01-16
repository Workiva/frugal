// Autogenerated by Frugal Compiler (2.14.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

import 'dart:typed_data' show Uint8List;
import 'package:thrift/thrift.dart' as thrift;
import 'package:v1_music/v1_music.dart' as t_v1_music;

/// Comments (with an @ symbol) will be added to generated code.
class Track implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = new thrift.TStruct("Track");
  static final thrift.TField _TITLE_FIELD_DESC = new thrift.TField("title", thrift.TType.STRING, 1);
  static final thrift.TField _ARTIST_FIELD_DESC = new thrift.TField("artist", thrift.TType.STRING, 2);
  static final thrift.TField _PUBLISHER_FIELD_DESC = new thrift.TField("publisher", thrift.TType.STRING, 3);
  static final thrift.TField _COMPOSER_FIELD_DESC = new thrift.TField("composer", thrift.TType.STRING, 4);
  static final thrift.TField _DURATION_FIELD_DESC = new thrift.TField("duration", thrift.TType.DOUBLE, 5);
  static final thrift.TField _PRO_FIELD_DESC = new thrift.TField("pro", thrift.TType.I32, 6);

  String _title;
  static const int TITLE = 1;
  String _artist;
  static const int ARTIST = 2;
  String _publisher;
  static const int PUBLISHER = 3;
  String _composer;
  static const int COMPOSER = 4;
  double _duration = 0.0;
  static const int DURATION = 5;
  int _pro;
  static const int PRO = 6;

  bool __isset_duration = false;
  bool __isset_pro = false;

  Track() {
  }

  String get title => this._title;

  set title(String title) {
    this._title = title;
  }

  bool isSetTitle() => this.title != null;

  unsetTitle() {
    this.title = null;
  }

  String get artist => this._artist;

  set artist(String artist) {
    this._artist = artist;
  }

  bool isSetArtist() => this.artist != null;

  unsetArtist() {
    this.artist = null;
  }

  String get publisher => this._publisher;

  set publisher(String publisher) {
    this._publisher = publisher;
  }

  bool isSetPublisher() => this.publisher != null;

  unsetPublisher() {
    this.publisher = null;
  }

  String get composer => this._composer;

  set composer(String composer) {
    this._composer = composer;
  }

  bool isSetComposer() => this.composer != null;

  unsetComposer() {
    this.composer = null;
  }

  double get duration => this._duration;

  set duration(double duration) {
    this._duration = duration;
    this.__isset_duration = true;
  }

  bool isSetDuration() => this.__isset_duration;

  unsetDuration() {
    this.__isset_duration = false;
  }

  int get pro => this._pro;

  set pro(int pro) {
    this._pro = pro;
    this.__isset_pro = true;
  }

  bool isSetPro() => this.__isset_pro;

  unsetPro() {
    this.__isset_pro = false;
  }

  getFieldValue(int fieldID) {
    switch (fieldID) {
      case TITLE:
        return this.title;
      case ARTIST:
        return this.artist;
      case PUBLISHER:
        return this.publisher;
      case COMPOSER:
        return this.composer;
      case DURATION:
        return this.duration;
      case PRO:
        return this.pro;
      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  setFieldValue(int fieldID, Object value) {
    switch(fieldID) {
      case TITLE:
        if(value == null) {
          unsetTitle();
        } else {
          this.title = value as String;
        }
        break;

      case ARTIST:
        if(value == null) {
          unsetArtist();
        } else {
          this.artist = value as String;
        }
        break;

      case PUBLISHER:
        if(value == null) {
          unsetPublisher();
        } else {
          this.publisher = value as String;
        }
        break;

      case COMPOSER:
        if(value == null) {
          unsetComposer();
        } else {
          this.composer = value as String;
        }
        break;

      case DURATION:
        if(value == null) {
          unsetDuration();
        } else {
          this.duration = value as double;
        }
        break;

      case PRO:
        if(value == null) {
          unsetPro();
        } else {
          this.pro = value as int;
        }
        break;

      default:
        throw new ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  // Returns true if the field corresponding to fieldID is set (has been assigned a value) and false otherwise
  bool isSet(int fieldID) {
    switch(fieldID) {
      case TITLE:
        return isSetTitle();
      case ARTIST:
        return isSetArtist();
      case PUBLISHER:
        return isSetPublisher();
      case COMPOSER:
        return isSetComposer();
      case DURATION:
        return isSetDuration();
      case PRO:
        return isSetPro();
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
        case TITLE:
          if(field.type == thrift.TType.STRING) {
            title = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case ARTIST:
          if(field.type == thrift.TType.STRING) {
            artist = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case PUBLISHER:
          if(field.type == thrift.TType.STRING) {
            publisher = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case COMPOSER:
          if(field.type == thrift.TType.STRING) {
            composer = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case DURATION:
          if(field.type == thrift.TType.DOUBLE) {
            duration = iprot.readDouble();
            this.__isset_duration = true;
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case PRO:
          if(field.type == thrift.TType.I32) {
            pro = iprot.readI32();
            this.__isset_pro = true;
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
    if(this.title != null) {
      oprot.writeFieldBegin(_TITLE_FIELD_DESC);
      oprot.writeString(title);
      oprot.writeFieldEnd();
    }
    if(this.artist != null) {
      oprot.writeFieldBegin(_ARTIST_FIELD_DESC);
      oprot.writeString(artist);
      oprot.writeFieldEnd();
    }
    if(this.publisher != null) {
      oprot.writeFieldBegin(_PUBLISHER_FIELD_DESC);
      oprot.writeString(publisher);
      oprot.writeFieldEnd();
    }
    if(this.composer != null) {
      oprot.writeFieldBegin(_COMPOSER_FIELD_DESC);
      oprot.writeString(composer);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldBegin(_DURATION_FIELD_DESC);
    oprot.writeDouble(duration);
    oprot.writeFieldEnd();
    oprot.writeFieldBegin(_PRO_FIELD_DESC);
    oprot.writeI32(pro);
    oprot.writeFieldEnd();
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  String toString() {
    StringBuffer ret = new StringBuffer("Track(");

    ret.write("title:");
    if(this.title == null) {
      ret.write("null");
    } else {
      ret.write(this.title);
    }

    ret.write(", ");
    ret.write("artist:");
    if(this.artist == null) {
      ret.write("null");
    } else {
      ret.write(this.artist);
    }

    ret.write(", ");
    ret.write("publisher:");
    if(this.publisher == null) {
      ret.write("null");
    } else {
      ret.write(this.publisher);
    }

    ret.write(", ");
    ret.write("composer:");
    if(this.composer == null) {
      ret.write("null");
    } else {
      ret.write(this.composer);
    }

    ret.write(", ");
    ret.write("duration:");
    ret.write(this.duration);

    ret.write(", ");
    ret.write("pro:");
    String pro_name = t_v1_music.PerfRightsOrg.VALUES_TO_NAMES[this.pro];
    if(pro_name != null) {
      ret.write(pro_name);
      ret.write(" (");
    }
    ret.write(this.pro);
    if(pro_name != null) {
      ret.write(")");
    }

    ret.write(")");

    return ret.toString();
  }

  validate() {
    // check for required fields
    // check that fields of type enum have valid values
    if(isSetPro() && !t_v1_music.PerfRightsOrg.VALID_VALUES.contains(pro)) {
      throw new thrift.TProtocolError(thrift.TProtocolErrorType.INVALID_DATA, "The field 'pro' has been assigned the invalid value $pro");
    }
  }
}
