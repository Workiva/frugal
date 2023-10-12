// Autogenerated by Frugal Compiler (3.17.5)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:thrift/thrift.dart' as thrift;
import 'package:v1_music/v1_music.dart' as t_v1_music;

/// Comments (with an @ symbol) will be added to generated code.
class Track implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('Track');
  static final thrift.TField _TITLE_FIELD_DESC = thrift.TField('title', thrift.TType.STRING, 1);
  static final thrift.TField _ARTIST_FIELD_DESC = thrift.TField('artist', thrift.TType.STRING, 2);
  static final thrift.TField _PUBLISHER_FIELD_DESC = thrift.TField('publisher', thrift.TType.STRING, 3);
  static final thrift.TField _COMPOSER_FIELD_DESC = thrift.TField('composer', thrift.TType.STRING, 4);
  static final thrift.TField _DURATION_FIELD_DESC = thrift.TField('duration', thrift.TType.DOUBLE, 5);
  static final thrift.TField _PRO_FIELD_DESC = thrift.TField('pro', thrift.TType.I32, 6);

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
  /// [t_v1_music.PerfRightsOrg]
  int _pro;
  static const int PRO = 6;

  bool __isset_duration = false;
  bool __isset_pro = false;

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

  /// [t_v1_music.PerfRightsOrg]
  int get pro => this._pro;

  /// [t_v1_music.PerfRightsOrg]
  set pro(int pro) {
    this._pro = pro;
    this.__isset_pro = true;
  }

  bool isSetPro() => this.__isset_pro;

  unsetPro() {
    this.__isset_pro = false;
  }

  @override
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
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  @override
  setFieldValue(int fieldID, Object value) {
    switch (fieldID) {
      case TITLE:
        this.title = value as String;
        break;

      case ARTIST:
        this.artist = value as String;
        break;

      case PUBLISHER:
        this.publisher = value as String;
        break;

      case COMPOSER:
        this.composer = value as String;
        break;

      case DURATION:
        if (value == null) {
          unsetDuration();
        } else {
          this.duration = value as double;
        }
        break;

      case PRO:
        if (value == null) {
          unsetPro();
        } else {
          this.pro = value as int;
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
      case DURATION:
        return isSetDuration();
      case PRO:
        return isSetPro();
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
        case TITLE:
          if (field.type == thrift.TType.STRING) {
            this.title = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case ARTIST:
          if (field.type == thrift.TType.STRING) {
            this.artist = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case PUBLISHER:
          if (field.type == thrift.TType.STRING) {
            this.publisher = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case COMPOSER:
          if (field.type == thrift.TType.STRING) {
            this.composer = iprot.readString();
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case DURATION:
          if (field.type == thrift.TType.DOUBLE) {
            this.duration = iprot.readDouble();
            this.__isset_duration = true;
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case PRO:
          if (field.type == thrift.TType.I32) {
            this.pro = iprot.readI32();
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

    validate();
  }

  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    final elem0 = title;
    if (elem0 != null) {
      oprot.writeFieldBegin(_TITLE_FIELD_DESC);
      oprot.writeString(elem0);
      oprot.writeFieldEnd();
    }
    final elem1 = artist;
    if (elem1 != null) {
      oprot.writeFieldBegin(_ARTIST_FIELD_DESC);
      oprot.writeString(elem1);
      oprot.writeFieldEnd();
    }
    final elem2 = publisher;
    if (elem2 != null) {
      oprot.writeFieldBegin(_PUBLISHER_FIELD_DESC);
      oprot.writeString(elem2);
      oprot.writeFieldEnd();
    }
    final elem3 = composer;
    if (elem3 != null) {
      oprot.writeFieldBegin(_COMPOSER_FIELD_DESC);
      oprot.writeString(elem3);
      oprot.writeFieldEnd();
    }
    final elem4 = duration;
    oprot.writeFieldBegin(_DURATION_FIELD_DESC);
    oprot.writeDouble(elem4);
    oprot.writeFieldEnd();
    final elem5 = pro;
    oprot.writeFieldBegin(_PRO_FIELD_DESC);
    oprot.writeI32(elem5);
    oprot.writeFieldEnd();
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  @override
  String toString() {
    StringBuffer ret = StringBuffer('Track(');

    ret.write('title:');
    if (this.title == null) {
      ret.write('null');
    } else {
      ret.write(this.title);
    }

    ret.write(', ');
    ret.write('artist:');
    if (this.artist == null) {
      ret.write('null');
    } else {
      ret.write(this.artist);
    }

    ret.write(', ');
    ret.write('publisher:');
    if (this.publisher == null) {
      ret.write('null');
    } else {
      ret.write(this.publisher);
    }

    ret.write(', ');
    ret.write('composer:');
    if (this.composer == null) {
      ret.write('null');
    } else {
      ret.write(this.composer);
    }

    ret.write(', ');
    ret.write('duration:');
    ret.write(this.duration);

    ret.write(', ');
    ret.write('pro:');
    String pro_name = t_v1_music.PerfRightsOrg.VALUES_TO_NAMES[this.pro];
    if (pro_name != null) {
      ret.write(pro_name);
      ret.write(' (');
    }
    ret.write(this.pro);
    if (pro_name != null) {
      ret.write(')');
    }

    ret.write(')');

    return ret.toString();
  }

  @override
  bool operator ==(Object o) {
    if (o is Track) {
      return this.title == o.title &&
        this.artist == o.artist &&
        this.publisher == o.publisher &&
        this.composer == o.composer &&
        this.duration == o.duration &&
        this.pro == o.pro;
    }
    return false;
  }

  @override
  int get hashCode {
    var value = 17;
    value = (value * 31) ^ this.title.hashCode;
    value = (value * 31) ^ this.artist.hashCode;
    value = (value * 31) ^ this.publisher.hashCode;
    value = (value * 31) ^ this.composer.hashCode;
    value = (value * 31) ^ this.duration.hashCode;
    value = (value * 31) ^ this.pro.hashCode;
    return value;
  }

  Track clone({
    String title,
    String artist,
    String publisher,
    String composer,
    double duration,
    int pro,
  }) {
    return Track()
      ..title = title ?? this.title
      ..artist = artist ?? this.artist
      ..publisher = publisher ?? this.publisher
      ..composer = composer ?? this.composer
      ..duration = duration ?? this.duration
      ..pro = pro ?? this.pro;
  }

  validate() {
    // check that fields of type enum have valid values
    if (isSetPro() && !t_v1_music.PerfRightsOrg.VALID_VALUES.contains(this.pro)) {
      throw thrift.TProtocolError(thrift.TProtocolErrorType.INVALID_DATA, "The field 'pro' has been assigned the invalid value ${this.pro}");
    }
  }
}
