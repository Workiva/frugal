// Autogenerated by Frugal Compiler (1.18.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

library v1_music.src.f_store_scope;

import 'dart:async';

import 'dart:typed_data' show Uint8List;
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;

import 'package:v1_music/v1_music.dart' as t_v1_music;
import 'f_store_structs.dart' as t_store_file;


/// Services are the API for client and server interaction.
/// Users can buy an album or enter a giveaway for a free album.
abstract class FStore {

  Future<t_v1_music.Album> buyAlbum(frugal.FContext ctx, String aSIN, String acct);

  Future<bool> enterAlbumGiveaway(frugal.FContext ctx, String email, String name);
}

/// Services are the API for client and server interaction.
/// Users can buy an album or enter a giveaway for a free album.
class FStoreClient implements FStore {
  Map<String, frugal.FMethod> _methods;

  FStoreClient(frugal.FTransport transport, frugal.FProtocolFactory protocolFactory, [List<frugal.Middleware> middleware]) {
    _transport = transport;
    _transport.setRegistry(new frugal.FClientRegistry());
    _protocolFactory = protocolFactory;
    _oprot = _protocolFactory.getProtocol(_transport);

    writeLock = new frugal.Lock();
    this._methods = {};
    this._methods['buyAlbum'] = new frugal.FMethod(this._buyAlbum, 'Store', 'buyAlbum', middleware);
    this._methods['enterAlbumGiveaway'] = new frugal.FMethod(this._enterAlbumGiveaway, 'Store', 'enterAlbumGiveaway', middleware);
  }

  frugal.FTransport _transport;
  frugal.FProtocolFactory _protocolFactory;
  frugal.FProtocol _oprot;
  frugal.FProtocol get oprot => _oprot;
  frugal.Lock writeLock;

  Future<t_v1_music.Album> buyAlbum(frugal.FContext ctx, String aSIN, String acct) {
    return this._methods['buyAlbum']([ctx, aSIN, acct]);
  }

  Future<t_v1_music.Album> _buyAlbum(frugal.FContext ctx, String aSIN, String acct) async {
    var controller = new StreamController();
    var closeSubscription = _transport.onClose.listen((_) {
      controller.addError(new thrift.TTransportError(
        thrift.TTransportErrorType.NOT_OPEN,
        "Transport closed before request completed."));
      });
    _transport.register(ctx, _recvBuyAlbumHandler(ctx, controller));
    await writeLock.lock();
    try {
      try {
        oprot.writeRequestHeader(ctx);
        oprot.writeMessageBegin(new thrift.TMessage("buyAlbum", thrift.TMessageType.CALL, 0));
        t_store_file.buyAlbum_args args = new t_store_file.buyAlbum_args();
        args.aSIN = aSIN;
        args.acct = acct;
        args.write(oprot);
        oprot.writeMessageEnd();
        await oprot.transport.flush();
      } finally {
        writeLock.unlock();
      }

      return await controller.stream.first.timeout(ctx.timeout);
    } finally {
      closeSubscription.cancel();
      _transport.unregister(ctx);
    }
  }

  _recvBuyAlbumHandler(frugal.FContext ctx, StreamController controller) {
    buyAlbumCallback(thrift.TTransport transport) {
      try {
        var iprot = _protocolFactory.getProtocol(transport);
        iprot.readResponseHeader(ctx);
        thrift.TMessage msg = iprot.readMessageBegin();
        if (msg.type == thrift.TMessageType.EXCEPTION) {
          thrift.TApplicationError error = thrift.TApplicationError.read(iprot);
          iprot.readMessageEnd();
          if (error.type == frugal.FTransport.RESPONSE_TOO_LARGE) {
            controller.addError(new frugal.FMessageSizeError.response());
            return;
          }
          throw error;
        }

        t_store_file.buyAlbum_result result = new t_store_file.buyAlbum_result();
        result.read(iprot);
        iprot.readMessageEnd();
        if (result.isSetSuccess()) {
          controller.add(result.success);
          return;
        }

        if (result.error != null) {
          controller.addError(result.error);
          return;
        }
        throw new thrift.TApplicationError(
          thrift.TApplicationErrorType.MISSING_RESULT, "buyAlbum failed: unknown result"
        );
      } catch(e) {
        controller.addError(e);
        rethrow;
      }
    }
    return buyAlbumCallback;
  }

  Future<bool> enterAlbumGiveaway(frugal.FContext ctx, String email, String name) {
    return this._methods['enterAlbumGiveaway']([ctx, email, name]);
  }

  Future<bool> _enterAlbumGiveaway(frugal.FContext ctx, String email, String name) async {
    var controller = new StreamController();
    var closeSubscription = _transport.onClose.listen((_) {
      controller.addError(new thrift.TTransportError(
        thrift.TTransportErrorType.NOT_OPEN,
        "Transport closed before request completed."));
      });
    _transport.register(ctx, _recvEnterAlbumGiveawayHandler(ctx, controller));
    await writeLock.lock();
    try {
      try {
        oprot.writeRequestHeader(ctx);
        oprot.writeMessageBegin(new thrift.TMessage("enterAlbumGiveaway", thrift.TMessageType.CALL, 0));
        t_store_file.enterAlbumGiveaway_args args = new t_store_file.enterAlbumGiveaway_args();
        args.email = email;
        args.name = name;
        args.write(oprot);
        oprot.writeMessageEnd();
        await oprot.transport.flush();
      } finally {
        writeLock.unlock();
      }

      return await controller.stream.first.timeout(ctx.timeout);
    } finally {
      closeSubscription.cancel();
      _transport.unregister(ctx);
    }
  }

  _recvEnterAlbumGiveawayHandler(frugal.FContext ctx, StreamController controller) {
    enterAlbumGiveawayCallback(thrift.TTransport transport) {
      try {
        var iprot = _protocolFactory.getProtocol(transport);
        iprot.readResponseHeader(ctx);
        thrift.TMessage msg = iprot.readMessageBegin();
        if (msg.type == thrift.TMessageType.EXCEPTION) {
          thrift.TApplicationError error = thrift.TApplicationError.read(iprot);
          iprot.readMessageEnd();
          if (error.type == frugal.FTransport.RESPONSE_TOO_LARGE) {
            controller.addError(new frugal.FMessageSizeError.response());
            return;
          }
          throw error;
        }

        t_store_file.enterAlbumGiveaway_result result = new t_store_file.enterAlbumGiveaway_result();
        result.read(iprot);
        iprot.readMessageEnd();
        if (result.isSetSuccess()) {
          controller.add(result.success);
          return;
        }

        throw new thrift.TApplicationError(
          thrift.TApplicationErrorType.MISSING_RESULT, "enterAlbumGiveaway failed: unknown result"
        );
      } catch(e) {
        controller.addError(e);
        rethrow;
      }
    }
    return enterAlbumGiveawayCallback;
  }

}
