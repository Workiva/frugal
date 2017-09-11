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

/// An internal callback which is constructed by generated code and invoked when
/// a pub/sub message is received. An FAsyncCallback is passed an
/// in-memory [TTransport] which wraps the complete message. The callback
/// returns an error or throws an exception if an unrecoverable error occurs and
/// the transport needs to be shutdown.
typedef void FAsyncCallback(TTransport transport);

/// Transport layer for scope subscribers.
abstract class FSubscriberTransport {
  /// Queries whether the transport is subscribed to a topic.
  /// Returns [true] if the transport is subscribed to a topic.
  bool get isSubscribed;

  /// Sets the subscribe topic and opens the transport.
  Future<Null> subscribe(String topic, FAsyncCallback callback);

  /// Unsets the subscribe topic and closes the transport.
  Future<Null> unsubscribe();

  /// Unsubscribe and remove durable information on the server,
  /// if applicable.
  Future remove() => unsubscribe();
}

/// Produces [FSubscriberTransport] instances.
abstract class FSubscriberTransportFactory {
  /// Return a new [FSubscriberTransport] instance.
  FSubscriberTransport getTransport();
}
