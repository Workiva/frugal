package frugal

// FSubscription is a subscription to a pub/sub topic created by a scope. The
// topic subscription is actually handled by an FScopeTransport, which the
// FSubscription wraps. Each FSubscription should have its own FScopeTransport.
// The FSubscription is used to unsubscribe from the topic.
type FSubscription struct {
	topic     string
	transport FScopeTransport
	errorC    chan error
}

// NewFSubscription creates a new FSubscription to the given topic which should
// be subscribed on the given FScopeTransport. This is to be used by generated
// code and should not be called directly.
func NewFSubscription(topic string, transport FScopeTransport) *FSubscription {
	return &FSubscription{
		topic:     topic,
		transport: transport,
		errorC:    make(chan error, 1),
	}
}

// Unsubscribe from the topic.
func (s *FSubscription) Unsubscribe() error {
	return s.transport.Close()
}

// Topic returns the subscription topic name.
func (s *FSubscription) Topic() string {
	return s.topic
}

// Error returns a channel which is signaled when something went wrong with the
// subscription. If an error is returned on this channel, the Subscription has
// been closed.
//
// DEPRECATED. TODO 2.0.0: Remove in a future release.
func (s *FSubscription) Error() <-chan error {
	return s.errorC
}
