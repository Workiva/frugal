import 'dart:typed_data';

import 'package:frugal/src/frugal.dart';
import 'package:thrift/thrift.dart';

/// Pack the arguments for transmission
Uint8List prepareMessage(
  FContext ctx,
  String method,
  TBase args,
  int kind,
  FProtocolFactory pfactory,
  int limit,
) {
  final memoryBuffer = TMemoryOutputBuffer(limit);
  final oprot = pfactory.getProtocol(memoryBuffer);
  oprot.writeRequestHeader(ctx);
  oprot.writeMessageBegin(TMessage(method, kind, 0));
  args.write(oprot);
  oprot.writeMessageEnd();
  return memoryBuffer.writeBytes;
}

/// Unpack the response into the [result]
void processReply(
  FContext ctx,
  TBase result,
  TTransport response,
  FProtocolFactory pfactory,
) {
  final iprot = pfactory.getProtocol(response);
  iprot.readResponseHeader(ctx);
  final msg = iprot.readMessageBegin();
  if (msg.type == TMessageType.EXCEPTION) {
    final error = TApplicationError.read(iprot);
    iprot.readMessageEnd();
    if (error.type == FrugalTTransportErrorType.REQUEST_TOO_LARGE) {
      throw TTransportError(FrugalTTransportErrorType.RESPONSE_TOO_LARGE, error.message);
    }
    throw error;
  }

  result.read(iprot);
  iprot.readMessageEnd();
}
