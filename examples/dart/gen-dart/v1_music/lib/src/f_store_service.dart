// Autogenerated by Frugal Compiler (3.16.27)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING



// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:logging/logging.dart' as logging;
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;
import 'package:w_common/disposable.dart' as disposable;

import 'package:v1_music/v1_music.dart' as t_v1_music;


/// Services are the API for client and server interaction.
/// Users can buy an album or enter a giveaway for a free album.
abstract class FStore {
  Future<t_v1_music.Album> buyAlbum(frugal.FContext ctx, String aSIN, String acct);

  /// Deprecated: use something else
  @deprecated
  Future<bool> enterAlbumGiveaway(frugal.FContext ctx, String email, String name);
}

FStoreClient fStoreClientFactory(frugal.FServiceProvider provider, {List<frugal.Middleware> middleware}) =>
    FStoreClient(provider, middleware);

/// Services are the API for client and server interaction.
/// Users can buy an album or enter a giveaway for a free album.
class FStoreClient extends disposable.Disposable implements FStore {
  static final logging.Logger _frugalLog = logging.Logger('Store');
  Map<String, frugal.FMethod> _methods;

  FStoreClient(frugal.FServiceProvider provider, [List<frugal.Middleware> middleware])
      : this._provider = provider {
    _transport = provider.transport;
    _protocolFactory = provider.protocolFactory;
    var combined = middleware ?? [];
    combined.addAll(provider.middleware);
    this._methods = {};
    this._methods['buyAlbum'] = frugal.FMethod(this._buyAlbum, 'Store', 'buyAlbum', combined);
    this._methods['enterAlbumGiveaway'] = frugal.FMethod(this._enterAlbumGiveaway, 'Store', 'enterAlbumGiveaway', combined);
  }

  frugal.FServiceProvider _provider;
  frugal.FTransport _transport;
  frugal.FProtocolFactory _protocolFactory;

  @override
  Future<Null> onDispose() async {
    if (_provider is disposable.Disposable && !_provider.isOrWillBeDisposed)  {
      return _provider.dispose();
    }
    return null;
  }

  @override
  Future<t_v1_music.Album> buyAlbum(frugal.FContext ctx, String aSIN, String acct) {
    return this._methods['buyAlbum']([ctx, aSIN, acct]).then((value) => value as t_v1_music.Album);
  }

  Future<t_v1_music.Album> _buyAlbum(frugal.FContext ctx, String aSIN, String acct) async {
    final args = buyAlbum_args();
    args.aSIN = aSIN;
    args.acct = acct;
    final message = frugal.prepareMessage(ctx, 'buyAlbum', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = buyAlbum_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    if (result.error != null) {
      throw result.error;
    }
    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'buyAlbum failed: unknown result'
    );
  }
  /// Deprecated: use something else
  @deprecated
  @override
  Future<bool> enterAlbumGiveaway(frugal.FContext ctx, String email, String name) {
    _frugalLog.warning("Call to deprecated function 'Store.enterAlbumGiveaway'");
    return this._methods['enterAlbumGiveaway']([ctx, email, name]).then((value) => value as bool);
  }

  Future<bool> _enterAlbumGiveaway(frugal.FContext ctx, String email, String name) async {
    final args = enterAlbumGiveaway_args();
    args.email = email;
    args.name = name;
    final message = frugal.prepareMessage(ctx, 'enterAlbumGiveaway', args, thrift.TMessageType.CALL, _protocolFactory, _transport.requestSizeLimit);
    var response = await _transport.request(ctx, message);

    final result = enterAlbumGiveaway_result();
    frugal.processReply(ctx, result, response, _protocolFactory);
    if (result.success != null) {
      return result.success;
    }

    throw thrift.TApplicationError(
      frugal.FrugalTApplicationErrorType.MISSING_RESULT, 'enterAlbumGiveaway failed: unknown result'
    );
  }
}

// ignore: camel_case_types
class buyAlbum_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('buyAlbum_args');
  static final thrift.TField _ASIN_FIELD_DESC = thrift.TField('ASIN', thrift.TType.STRING, 1);
  static final thrift.TField _ACCT_FIELD_DESC = thrift.TField('acct', thrift.TType.STRING, 2);

  String aSIN;
  String acct;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (this.aSIN != null) {
      oprot.writeFieldBegin(_ASIN_FIELD_DESC);
      oprot.writeString(this.aSIN);
      oprot.writeFieldEnd();
    }
    if (this.acct != null) {
      oprot.writeFieldBegin(_ACCT_FIELD_DESC);
      oprot.writeString(this.acct);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class buyAlbum_result extends frugal.FGeneratedArgsResultBase {
  t_v1_music.Album success;
  t_v1_music.PurchasingError error;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.STRUCT) {
            this.success = t_v1_music.Album();
            success.read(iprot);
          } else {
            thrift.TProtocolUtil.skip(iprot, field.type);
          }
          break;
        case 1:
          if (field.type == thrift.TType.STRUCT) {
            this.error = t_v1_music.PurchasingError();
            error.read(iprot);
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

  validate() {
  }
}
// ignore: camel_case_types
class enterAlbumGiveaway_args extends frugal.FGeneratedArgsResultBase {
  static final thrift.TStruct _STRUCT_DESC = thrift.TStruct('enterAlbumGiveaway_args');
  static final thrift.TField _EMAIL_FIELD_DESC = thrift.TField('email', thrift.TType.STRING, 1);
  static final thrift.TField _NAME_FIELD_DESC = thrift.TField('name', thrift.TType.STRING, 2);

  String email;
  String name;


  @override
  write(thrift.TProtocol oprot) {
    validate();

    oprot.writeStructBegin(_STRUCT_DESC);
    if (this.email != null) {
      oprot.writeFieldBegin(_EMAIL_FIELD_DESC);
      oprot.writeString(this.email);
      oprot.writeFieldEnd();
    }
    if (this.name != null) {
      oprot.writeFieldBegin(_NAME_FIELD_DESC);
      oprot.writeString(this.name);
      oprot.writeFieldEnd();
    }
    oprot.writeFieldStop();
    oprot.writeStructEnd();
  }

  validate() {
  }
}
// ignore: camel_case_types
class enterAlbumGiveaway_result extends frugal.FGeneratedArgsResultBase {
  bool success;


  @override
  read(thrift.TProtocol iprot) {
    iprot.readStructBegin();
    for (thrift.TField field = iprot.readFieldBegin();
        field.type != thrift.TType.STOP;
        field = iprot.readFieldBegin()) {
      switch (field.id) {
        case 0:
          if (field.type == thrift.TType.BOOL) {
            this.success = iprot.readBool();
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

  validate() {
  }
}
