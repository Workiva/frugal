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

part of frugal.src.frugal;

/// Comparable to Thrift's [TTransport] in that it represents the transport
/// layer for frugal clients. However, frugal is callback based and sends only
/// framed data. Therefore, instead of exposing read, write, and flush, the
/// transport has a simple [oneway] and [request] methods that send framed
/// frugal requests.
abstract class FTransport {
  MonitorRunner _monitor;
  StreamController _closeController = new StreamController.broadcast();

  /// Limits the size of requests to the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int requestSizeLimit;

  /// Create an [FTransport] with the optional [requestSizeLimit].
  FTransport({this.requestSizeLimit});

  /// Listen to close events on the transport.
  Stream<Object> get onClose => _closeController.stream;

  /// Set an [FTransportMonitor] on the transport.
  set monitor(FTransportMonitor monitor) {
    _monitor = new MonitorRunner(monitor, this);
  }

  /// Queries whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen;

  /// Opens the transport for reading/writing.
  /// Throws [TTransportError] if the transport could not be opened.
  Future open();

  /// Closes the transport.
  Future close([Error error]) => _signalClose(error);

  /// Send the given framed frugal payload over the transport and return a
  /// future containing the response. Throws [TTransportError] if problems
  /// are encountered with the request.
  Future<TTransport> request(FContext ctx, Uint8List payload);

  /// Send the given framed frugal payload over the transport and don't
  /// expect a response.
  Future<Null> oneway(FContext ctx, Uint8List payload);

  Future _signalClose(cause) async {
    _closeController.add(cause);
    await _monitor?.onClose(cause);
  }

  /// Checks if a transport is open and the payload is within the request size
  /// limit.
  void _preflightRequestCheck(Uint8List payload) {
    if (!isOpen) {
      throw new TTransportError(
          FrugalTTransportErrorType.NOT_OPEN, 'transport not open');
    }

    if (requestSizeLimit != null &&
        requestSizeLimit > 0 &&
        payload.length > requestSizeLimit) {
      throw new TTransportError(FrugalTTransportErrorType.REQUEST_TOO_LARGE);
    }
  }
}
