/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import 'dart:async';
import 'dart:typed_data';

import 'package:frugal/src/frugal/transport/f_async_transport.dart';
import 'package:frugal/src/frugal/transport/t_framed_transport.dart';
import 'package:logging/logging.dart';
import 'package:thrift/thrift.dart' hide TFramedTransport;


/// Wraps a [TSocketTransport] to produce an [FTransport] which uses the given
/// socket for send/callback operations in a way that is compatible with Frugal.
/// Messages received on the [TSocket] (i.e. Frugal frames) are routed to the
/// [FAsyncTransport]'s handleResponse method.
class FAdapterTransport extends FAsyncTransport {
  final Logger _adapterTransportLog = new Logger('FAdapterTransport');
  TFramedTransport _framedTransport;

  StreamSubscription<FrameWrapper> _onFrameSub;

  /// Create an [FAdapterTransport] with the given [TSocketTransport].
  FAdapterTransport(TSocketTransport transport)
      : _framedTransport = new TFramedTransport(transport.socket),
        super() {
    // If there is an error on the socket, close the transport pessimistically.
    // This error is already logged upstream in TSocketTransport.
    listenToStream(transport.socket.onError, (e) => close(e));
    // Forward state changes on to the transport monitor.
    // Note: Just forwarding OPEN on for the time-being.
    listenToStream(transport.socket.onState, (state) {
      if (state == TSocketState.OPEN) monitor?.signalOpen();
    });

    manageDisposable(_framedTransport);
  }

  @override
  bool get isOpen => _framedTransport.isOpen;

  @override
  Future open() async {
    await _framedTransport.open();

    _onFrameSub = _framedTransport.onFrame.listen(_handleFrame);
  }

  void _handleFrame(FrameWrapper frame) {
    try {
      handleResponse(frame.frameBytes);
    } catch (e) {
      // Fatal error. Close the transport.
      _adapterTransportLog.severe(
          "FAsyncCallback had a fatal error ${e.toString()}." +
              "Closing transport.");
      close(e);
    }
  }

  @override
  Future close([Error error]) async {
    await _framedTransport?.close();
    await _onFrameSub?.cancel();
    await super.close(error);
  }

  @override
  Future<Null> flush(Uint8List payload) {
    _framedTransport.socket.send(payload);
    return new Future.value();
  }

  @override
  Future<Null> onDispose() async {
    await close();
    await super.onDispose();
  }
}
