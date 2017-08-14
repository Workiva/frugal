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

/// A subscription to a pub/sub topic created by a scope. The topic subscription
/// is actually handled by an [FSubscriberTransport], which the [FSubscription]
/// wraps. Each [FSubscription] should have its own [FSubscriberTransport]. The
/// [FSubscription] is used to unsubscribe from the topic.
class FSubscription {
  /// Scope topic for the subscription.
  final String topic;
  FSubscriberTransport _transport;

  /// Create a new [FSubscription] with the given topic and transport.
  FSubscription(this.topic, this._transport);

  /// Unsubscribe from the topic.
  Future unsubscribe() => _transport.unsubscribe();

  /// Unsubscribes and removes durably stored information on the broker,
  /// if applicable.
  Future remove() => _transport.remove();
}
