// Autogenerated by Frugal Compiler (2.23.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING



import 'dart:async';

import 'dart:typed_data' show Uint8List;
import 'package:logging/logging.dart' as logging;
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;

import 'package:vendor_namespace/vendor_namespace.dart' as t_vendor_namespace;


abstract class FVendoredBase {}

class FVendoredBaseClient implements FVendoredBase {
  static final logging.Logger _frugalLog = new logging.Logger('VendoredBase');
  List<frugal.Middleware> _combinedMiddleware;
  FVendoredBaseClient(frugal.FServiceProvider provider, [List<frugal.Middleware> middleware]) {
    _transport = provider.transport;
    _protocolFactory = provider.protocolFactory;
    _combinedMiddleware = middleware ?? [];
    _combinedMiddleware.addAll(provider.middleware);
  }

  frugal.FTransport _transport;
  frugal.FProtocolFactory _protocolFactory;

}

