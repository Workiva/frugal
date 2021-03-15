// Autogenerated by Frugal Compiler (10.0.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:thrift/thrift.dart' as thrift;
import 'package:v1_music/v1_music.dart' as t_v1_music;

/// The IDL provides set, list, and map types for representing collections
/// of data.  Our Album struct contains a list of Tracks.
class Album implements thrift.TBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('Album');
  static final thrift.TField _TRACKS_FIELD_DESC = thrift.TField('tracks', thrift.TType.LIST, 1);
  static final thrift.TField _DURATION_FIELD_DESC = thrift.TField('duration', thrift.TType.DOUBLE, 2);
  static final thrift.TField _ASIN_FIELD_DESC = thrift.TField('ASIN', thrift.TType.STRING, 3);

  List<t_v1_music.Track> _tracks;
  static const int TRACKS = 1;
  double _duration = 0.0;
  static const int DURATION = 2;
  String _aSIN;
  static const int ASIN = 3;

  bool __isset_duration = false;

  List<t_v1_music.Track> get tracks => this._tracks;

  set tracks(List<t_v1_music.Track> tracks) {
    this._tracks = tracks;
  }

  bool isSetTracks() => this.tracks != null;

  unsetTracks() {
    this.tracks = null;
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

  String get aSIN => this._aSIN;

  set aSIN(String aSIN) {
    this._aSIN = aSIN;
  }

  bool isSetASIN() => this.aSIN != null;

  unsetASIN() {
    this.aSIN = null;
  }

  @override
  getFieldValue(int fieldID) {
    switch (fieldID) {
      case TRACKS:
        return this.tracks;
      case DURATION:
        return this.duration;
      case ASIN:
        return this.aSIN;
      default:
        throw ArgumentError("Field $fieldID doesn't exist!");
    }
  }

  @override
  setFieldValue(int fieldID, Object value) {
    switch (fieldID) {
      case TRACKS:
        if (value == null) {
          unsetTracks();
        } else {
          this.tracks = value as List<t_v1_music.Track>;
        }
        break;

      case DURATION:
        if (value == null) {
          unsetDuration();
        } else {
          this.duration = value as double;
        }
        break;

      case ASIN:
        if (value == null) {
          unsetASIN();
        } else {
          this.aSIN = value as String;
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
      case TRACKS:
        return isSetTracks();
      case DURATION:
        return isSetDuration();
      case ASIN:
        return isSetASIN();
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
        case TRACKS:
          if (field.type == thrift.TType.LIST) {
            thrift.TList elem0 = iprot.readListBegin();
            this.tracks = List<t_v1_music.Track>();
            for(int elem2 = 0; elem2 < elem0.length; ++elem2) {
              t_v1_music.Track elem1 = t_v1_music.Track();
              elem1.read(iprot);
              this.tracks.add(elem1);
            }
            iprot.readListEnd();
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
        case ASIN:
          if (field.type == thrift.TType.STRING) {
            this.aSIN = iprot.readString();
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
    if (this.tracks != null) {
      oprot.writeFieldBegin(_TRACKS_FIELD_DESC);
      oprot.writeListBegin(thrift.TList(thrift.TType.STRUCT, this.tracks.length));
      for(var elem3 in this.tracks) {
        elem3.write(oprot);
      }
      oprot.writeListEnd();
      oprot.writeFieldEnd();
    }
    oprot.writeFieldBegin(_DURATION_FIELD_DESC);
    oprot.writeDouble(this.duration);
    oprot.writeFieldEnd();
    if (this.aSIN != null) {
      oprot.writeFieldBegin(_ASIN_FIELD_DESC);
      oprot.writeString(this.aSIN);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  @override
  String toString() {
    StringBuffer ret = StringBuffer('Album(');

    ret.write('tracks:');
    if (this.tracks == null) {
      ret.write('null');
    } else {
      ret.write(this.tracks);
    }

    ret.write(', ');
    ret.write('duration:');
    ret.write(this.duration);

    ret.write(', ');
    ret.write('aSIN:');
    if (this.aSIN == null) {
      ret.write('null');
    } else {
      ret.write(this.aSIN);
    }

    ret.write(')');

    return ret.toString();
  }

  @override
  bool operator ==(Object o) {
    if (o is Album) {
      return DeepCollectionEquality().equals(this.tracks, o.tracks) &&
        this.duration == o.duration &&
        this.aSIN == o.aSIN;
    }
    return false;
  }

  @override
  int get hashCode {
    var value = 17;
    value = (value * 31) ^ this.tracks.hashCode;
    value = (value * 31) ^ this.duration.hashCode;
    value = (value * 31) ^ this.aSIN.hashCode;
    return value;
  }

  Album clone({
    List<t_v1_music.Track> tracks,
    double duration,
    String aSIN,
  }) {
    return Album()
      ..tracks = tracks ?? this.tracks
      ..duration = duration ?? this.duration
      ..aSIN = aSIN ?? this.aSIN;
  }

  validate() {
  }
}
