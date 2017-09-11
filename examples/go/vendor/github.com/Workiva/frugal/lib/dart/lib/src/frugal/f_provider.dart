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

/// Produces [FPublisherTransport], [FSubscriberTransport], and [FProtocol]
/// instances for use by pub/sub scopes. It does this by wrapping an
/// [FPublisherTransportFactory], an [FSubscriberTransportFactory], and an
/// [FProtocolFactory]. This also provides a shim for adding middleware to a
/// publisher or subscriber.
class FScopeProvider {
  /// [FPublisherTransportFactory] used by the scope.
  final FPublisherTransportFactory publisherTransportFactory;

  /// [FSubscriberTransportFactory] used by the scope.
  final FSubscriberTransportFactory subscriberTransportFactory;

  /// [FProtocolFactory] used by the scope.
  final FProtocolFactory protocolFactory;

  /// Middleware applied to publishers and subscribers.
  final List<Middleware> _middleware;

  /// Creates a new [FScopeProvider].
  FScopeProvider(this.publisherTransportFactory,
      this.subscriberTransportFactory, this.protocolFactory,
      {List<Middleware> middleware: null})
      : _middleware = middleware ?? [];

  /// The middleware stored on this FScopeProvider.
  List<Middleware> get middleware => new List.from(this._middleware);
}

/// The service equivalent of [FScopeProvider]. It produces [FTransport] and
/// [FProtocol] instances for use by RPC service clients. The main purpose of
/// this is to provide a shim for adding middleware to a client.
class FServiceProvider {
  /// [FTransport] used by the service.
  final FTransport transport;

  /// [FProtocolFactory] used by the service.
  final FProtocolFactory protocolFactory;

  /// Middleware applied to clients.
  final List<Middleware> _middleware;

  /// Creates a new [FServiceProvider].
  FServiceProvider(this.transport, this.protocolFactory,
      {List<Middleware> middleware: null})
      : _middleware = middleware ?? [];

  /// The middleware stored on this FServiceProvider.
  List<Middleware> get middleware => new List.from(this._middleware);
}
