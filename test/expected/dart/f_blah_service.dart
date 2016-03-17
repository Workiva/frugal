// Autogenerated by Frugal Compiler (1.0.6)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

library valid.src.f_blah_scope;

import 'dart:async';

import 'dart:typed_data' show Uint8List;
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;

import 'package:valid/valid.dart' as t_valid;
import 'blah.dart' as t_blah_file;


abstract class FBlah {

  /// Use this to ping the server.
  Future ping(frugal.FContext ctx);

  /// Use this to tell the sever how you feel.
  Future<int> bleh(frugal.FContext ctx, t_valid.Thing one, t_valid.Stuff two);
}

class FBlahClient implements FBlah {

  FBlahClient(frugal.FTransport transport, frugal.FProtocolFactory protocolFactory) {
    _transport = transport;
    _transport.setRegistry(new frugal.FClientRegistry());
    _protocolFactory = protocolFactory;
    _oprot = _protocolFactory.getProtocol(_transport);
  }

  frugal.FTransport _transport;
  frugal.FProtocolFactory _protocolFactory;
  frugal.FProtocol _oprot;
  frugal.FProtocol get oprot => _oprot;

  /// Use this to ping the server.
  Future ping(frugal.FContext ctx) async {
    var controller = new StreamController();
    _transport.register(ctx, _recvPingHandler(ctx, controller));
    try {
      oprot.writeRequestHeader(ctx);
      oprot.writeMessageBegin(new thrift.TMessage("ping", thrift.TMessageType.CALL, 0));
      t_blah_file.ping_args args = new t_blah_file.ping_args();
      args.write(oprot);
      oprot.writeMessageEnd();
      await oprot.transport.flush();
      return await controller.stream.first.timeout(ctx.timeout);
    } finally {
      _transport.unregister(ctx);
    }
  }

  _recvPingHandler(frugal.FContext ctx, StreamController controller) {
    pingCallback(thrift.TTransport transport) {
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

        t_blah_file.ping_result result = new t_blah_file.ping_result();
        result.read(iprot);
        iprot.readMessageEnd();
        controller.add(null);
      } catch(e) {
        controller.addError(e);
        rethrow;
      }
    }
    return pingCallback;
  }

  /// Use this to tell the sever how you feel.
  Future<int> bleh(frugal.FContext ctx, t_valid.Thing one, t_valid.Stuff two) async {
    var controller = new StreamController();
    _transport.register(ctx, _recvBlehHandler(ctx, controller));
    try {
      oprot.writeRequestHeader(ctx);
      oprot.writeMessageBegin(new thrift.TMessage("bleh", thrift.TMessageType.CALL, 0));
      t_blah_file.bleh_args args = new t_blah_file.bleh_args();
      args.one = one;
      args.two = two;
      args.write(oprot);
      oprot.writeMessageEnd();
      await oprot.transport.flush();
      return await controller.stream.first.timeout(ctx.timeout);
    } finally {
      _transport.unregister(ctx);
    }
  }

  _recvBlehHandler(frugal.FContext ctx, StreamController controller) {
    blehCallback(thrift.TTransport transport) {
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

        t_blah_file.bleh_result result = new t_blah_file.bleh_result();
        result.read(iprot);
        iprot.readMessageEnd();
        if (result.isSetSuccess()) {
          controller.add(result.success);
          return;
        }

        if (result.oops != null) {
          controller.addError(result.oops);
          return;
        }
        throw new thrift.TApplicationError(
          thrift.TApplicationErrorType.MISSING_RESULT, "bleh failed: unknown result"
        );
      } catch(e) {
        controller.addError(e);
        rethrow;
      }
    }
    return blehCallback;
  }

}
