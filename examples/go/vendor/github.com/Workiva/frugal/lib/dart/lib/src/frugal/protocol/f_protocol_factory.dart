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

/// Creates new [FProtocol] instances. It takes a [TProtocolFactory] and a
/// [TTransport] and returns an [FProtocol] which wraps a [TProtocol] produced
/// by the [TProtocolFactory]. The [TProtocol] itself wraps the provided
/// [TTransport]. This makes it easy to produce an [FProtocol] which uses any
/// existing Thrift transports and protocols in a composable manner.
class FProtocolFactory {
  TProtocolFactory _tProtocolFactory;

  /// Create an [FProtocolFactory] wrapping the given [TProtocolFactory].
  FProtocolFactory(this._tProtocolFactory);

  /// Construct a new [FProtocol] instance from the given [TTransport].
  FProtocol getProtocol(TTransport transport) {
    return new FProtocol(_tProtocolFactory.getProtocol(transport));
  }
}
