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

/// Transport layer for scope publishers.
abstract class FPublisherTransport {
  /// Query whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen;

  /// Open the transport for reading/writing.
  /// Throws [TTransportError] if the transport could not be opened.
  Future open();

  /// Close the transport.
  Future close();

  /// The maximum publish size permitted by the transport. If [publishSizeLimit]
  /// is a non-positive number, the transport is assumed to have no publish size
  /// limit.
  int get publishSizeLimit;

  /// Publish the given framed frugal payload over the transport.
  /// Throws [TTransportError] if publishing the payload failed.
  void publish(String topic, Uint8List payload);
}

/// Produces [FPublisherTransport] instances.
abstract class FPublisherTransportFactory {
  /// Return a new [FPublisherTransport] instance.
  FPublisherTransport getTransport();
}
