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

package frugal

// FSubscription is a subscription to a pub/sub topic created by a scope. The
// topic subscription is actually handled by an FScopeTransport, which the
// FSubscription wraps. Each FSubscription should have its own FScopeTransport.
// The FSubscription is used to unsubscribe from the topic.
type FSubscription struct {
	topic     string
	transport FSubscriberTransport
}

// remover allows unsubscribing and removing durably stored information
// on the message broker.
type remover interface {
	// Remove unsubscribes and removes durably stored information on the broker,
	// if applicable.
	Remove() error
}

// NewFSubscription creates a new FSubscription to the given topic which should
// be subscribed on the given FScopeTransport. This is to be used by generated
// code and should not be called directly.
func NewFSubscription(topic string, transport FSubscriberTransport) *FSubscription {
	return &FSubscription{
		topic:     topic,
		transport: transport,
	}
}

// Unsubscribe from the topic.
func (s *FSubscription) Unsubscribe() error {
	return s.transport.Unsubscribe()
}

// Remove unsubscribes and removes durably stored information on the broker,
// if applicable.
func (s *FSubscription) Remove() error {
	// If the subscriber transport has a remove method, use it
	// otherwise call unsubscribe
	// TODO 3.0 get rid of this
	if suspender, ok := s.transport.(remover); ok {
		return suspender.Remove()
	}
	return s.transport.Unsubscribe()
}

// Topic returns the subscription topic name.
func (s *FSubscription) Topic() string {
	return s.topic
}
