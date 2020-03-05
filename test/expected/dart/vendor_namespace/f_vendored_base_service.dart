// Autogenerated by Frugal Compiler (3.8.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING



// ignore_for_file: unused_import
// ignore_for_file: unused_field
import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:collection/collection.dart';
import 'package:logging/logging.dart' as logging;
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;

import 'package:vendor_namespace/vendor_namespace.dart' as t_vendor_namespace;


abstract class FVendoredBase {}

class FVendoredBaseClient implements FVendoredBase {
  static final logging.Logger _frugalLog = logging.Logger('VendoredBase');
  Map<String, frugal.FMethod> _methods;

  FVendoredBaseClient(frugal.FServiceProvider provider, [List<frugal.Middleware> middleware]) {
    _transport = provider.transport;
    _protocolFactory = provider.protocolFactory;
    var combined = middleware ?? [];
    combined.addAll(provider.middleware);
    this._methods = {};
  }

  frugal.FTransport _transport;
  frugal.FProtocolFactory _protocolFactory;

}

