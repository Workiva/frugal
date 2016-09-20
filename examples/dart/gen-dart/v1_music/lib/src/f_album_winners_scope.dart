// Autogenerated by Frugal Compiler (1.18.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

library v1_music.src.f_albumwinners_scope;

import 'dart:async';

import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;

import 'package:v1_music/v1_music.dart' as t_v1_music;


const String delimiter = '.';

/// Scopes are a Frugal extension to the IDL for declaring PubSub
/// semantics. Subscribers to this scope will be notified if they win a contest.
/// Scopes must have a prefix.
class AlbumWinnersPublisher {
  frugal.FScopeTransport fTransport;
  frugal.FProtocol fProtocol;
  Map<String, frugal.FMethod> _methods;
  frugal.Lock _writeLock;

  AlbumWinnersPublisher(frugal.FScopeProvider provider, [List<frugal.Middleware> middleware]) {
    fTransport = provider.fTransportFactory.getTransport();
    fProtocol = provider.fProtocolFactory.getProtocol(fTransport);
    _writeLock = new frugal.Lock();
    this._methods = {};
    this._methods['Winner'] = new frugal.FMethod(this._publishWinner, 'AlbumWinners', 'publishWinner', middleware);
  }

  Future open() {
    return fTransport.open();
  }

  Future close() {
    return fTransport.close();
  }

  Future publishWinner(frugal.FContext ctx, t_v1_music.Album req) {
    return this._methods['Winner']([ctx, req]);
  }

  Future _publishWinner(frugal.FContext ctx, t_v1_music.Album req) async {
    await _writeLock.lock();
    try {
      var op = "Winner";
      var prefix = "v1.music.";
      var topic = "${prefix}AlbumWinners${delimiter}${op}";
      fTransport.setTopic(topic);
      var oprot = fProtocol;
      var msg = new thrift.TMessage(op, thrift.TMessageType.CALL, 0);
      oprot.writeRequestHeader(ctx);
      oprot.writeMessageBegin(msg);
      req.write(oprot);
      oprot.writeMessageEnd();
      await oprot.transport.flush();
    } finally {
      _writeLock.unlock();
    }
  }
}


/// Scopes are a Frugal extension to the IDL for declaring PubSub
/// semantics. Subscribers to this scope will be notified if they win a contest.
/// Scopes must have a prefix.
class AlbumWinnersSubscriber {
  final frugal.FScopeProvider provider;
  final List<frugal.Middleware> _middleware;

  AlbumWinnersSubscriber(this.provider, [this._middleware]) {}

  Future<frugal.FSubscription> subscribeWinner(dynamic onAlbum(frugal.FContext ctx, t_v1_music.Album req)) async {
    var op = "Winner";
    var prefix = "v1.music.";
    var topic = "${prefix}AlbumWinners${delimiter}${op}";
    var transport = provider.fTransportFactory.getTransport();
    await transport.subscribe(topic, _recvWinner(op, provider.fProtocolFactory, onAlbum));
    return new frugal.FSubscription(topic, transport);
  }

  _recvWinner(String op, frugal.FProtocolFactory protocolFactory, dynamic onAlbum(frugal.FContext ctx, t_v1_music.Album req)) {
    frugal.FMethod method = new frugal.FMethod(onAlbum, 'AlbumWinners', 'subscribeAlbum', this._middleware);
    callbackWinner(thrift.TTransport transport) {
      var iprot = protocolFactory.getProtocol(transport);
      var ctx = iprot.readRequestHeader();
      var tMsg = iprot.readMessageBegin();
      if (tMsg.name != op) {
        thrift.TProtocolUtil.skip(iprot, thrift.TType.STRUCT);
        iprot.readMessageEnd();
        throw new thrift.TApplicationError(
        thrift.TApplicationErrorType.UNKNOWN_METHOD, tMsg.name);
      }
      var req = new t_v1_music.Album();
      req.read(iprot);
      iprot.readMessageEnd();
      method([ctx, req]);
    }
    return callbackWinner;
  }
}

