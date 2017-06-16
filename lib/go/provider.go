package frugal

// FScopeProvider produces FScopeTransports and FProtocols for use by pub/sub
// scopes. It does this by wrapping an FScopeTransportFactory and
// FProtocolFactory. This also provides a shim for adding middleware to a
// publisher or subscriber.
type FScopeProvider struct {
	publisherTransportFactory  FPublisherTransportFactory
	subscriberTransportFactory FSubscriberTransportFactory
	protocolFactory            *FProtocolFactory
	middleware                 []ServiceMiddleware
}

// NewFScopeProvider creates a new FScopeProvider using the given factories.
func NewFScopeProvider(pub FPublisherTransportFactory, sub FSubscriberTransportFactory,
	prot *FProtocolFactory, middleware ...ServiceMiddleware) *FScopeProvider {
	return &FScopeProvider{
		publisherTransportFactory:  pub,
		subscriberTransportFactory: sub,
		protocolFactory:            prot,
		middleware:                 middleware,
	}
}

// NewPublisher returns a new FPublisherTransport and FProtocol used by
// scope publishers.
func (p *FScopeProvider) NewPublisher() (FPublisherTransport, *FProtocolFactory) {
	transport := p.publisherTransportFactory.GetTransport()
	return transport, p.protocolFactory
}

// NewSubscriber returns a new FSubscriberTransport and FProtocolFactory used by
// scope subscribers.
func (p *FScopeProvider) NewSubscriber() (FSubscriberTransport, *FProtocolFactory) {
	transport := p.subscriberTransportFactory.GetTransport()
	return transport, p.protocolFactory
}

// GetMiddleware returns the ServiceMiddleware stored on this FScopeProvider.
func (p *FScopeProvider) GetMiddleware() []ServiceMiddleware {
	middleware := make([]ServiceMiddleware, len(p.middleware))
	copy(middleware, p.middleware)
	return middleware
}

type FUnifiedScopeProvider struct {
	publisherTransportFactory  FUnifiedPublisherTransportFactory
	subscriberTransportFactory FUnifiedSubscriberTransportFactory
	protocolFactory            *FProtocolFactory
	middleware                 []ServiceMiddleware
}

// NewFScopeProvider creates a new FScopeProvider using the given factories.
func NewFUnifiedScopeProvider(pub FUnifiedPublisherTransportFactory, sub FUnifiedSubscriberTransportFactory,
prot *FProtocolFactory, middleware ...ServiceMiddleware) *FUnifiedScopeProvider {
	return &FUnifiedScopeProvider{
		publisherTransportFactory:  pub,
		subscriberTransportFactory: sub,
		protocolFactory:            prot,
		middleware:                 middleware,
	}
}

// NewPublisher returns a new FPublisherTransport and FProtocol used by
// scope publishers.
func (p *FUnifiedScopeProvider) NewPublisher() (FUnifiedPublisherTransport, *FProtocolFactory) {
	transport := p.publisherTransportFactory.GetTransport()
	return transport, p.protocolFactory
}

// NewSubscriber returns a new FSubscriberTransport and FProtocolFactory used by
// scope subscribers.
func (p *FUnifiedScopeProvider) NewSubscriber() (FUnifiedSubscriberTransport, *FProtocolFactory) {
	transport := p.subscriberTransportFactory.GetTransport()
	return transport, p.protocolFactory
}

// GetMiddleware returns the ServiceMiddleware stored on this FScopeProvider.
func (p *FUnifiedScopeProvider) GetMiddleware() []ServiceMiddleware {
	middleware := make([]ServiceMiddleware, len(p.middleware))
	copy(middleware, p.middleware)
	return middleware
}

type FDurableScopeProvider struct {
	publisherTransportFactory  FDurablePublisherTransportFactory
	subscriberTransportFactory FDurableSubscriberTransportFactory
	protocolFactory            *FProtocolFactory
	middleware                 []ServiceMiddleware
}

// NewFDurableScopeProvider creates a new FScopeProvider using the given factories.
func NewFDurableScopeProvider(pub FDurablePublisherTransportFactory, sub FDurableSubscriberTransportFactory,
prot *FProtocolFactory, middleware ...ServiceMiddleware) *FDurableScopeProvider {
	return &FDurableScopeProvider{
		publisherTransportFactory:  pub,
		subscriberTransportFactory: sub,
		protocolFactory:            prot,
		middleware:                 middleware,
	}
}

// NewPublisher returns a new FPublisherTransport and FProtocol used by
// scope publishers.
func (p *FDurableScopeProvider) NewPublisher() (FDurablePublisherTransport, *FProtocolFactory) {
	transport := p.publisherTransportFactory.GetTransport()
	return transport, p.protocolFactory
}

// NewSubscriber returns a new FSubscriberTransport and FProtocolFactory used by
// scope subscribers.
func (p *FDurableScopeProvider) NewSubscriber() (FDurableSubscriberTransport, *FProtocolFactory) {
	transport := p.subscriberTransportFactory.GetTransport()
	return transport, p.protocolFactory
}

// GetMiddleware returns the ServiceMiddleware stored on this FScopeProvider.
func (p *FDurableScopeProvider) GetMiddleware() []ServiceMiddleware {
	middleware := make([]ServiceMiddleware, len(p.middleware))
	copy(middleware, p.middleware)
	return middleware
}

// FServiceProvider produces FTransports and FProtocolFactories for use by RPC
// service clients. The main purpose of this is to provide a shim for adding
// middleware to a client.
type FServiceProvider struct {
	transport       FTransport
	protocolFactory *FProtocolFactory
	middleware      []ServiceMiddleware
}

// NewFServiceProvider creates a new FServiceProvider containing the given
// FTransport and FProtocolFactory.
func NewFServiceProvider(transport FTransport, protocolFactory *FProtocolFactory, middleware ...ServiceMiddleware) *FServiceProvider {
	return &FServiceProvider{
		transport:       transport,
		protocolFactory: protocolFactory,
		middleware:      middleware,
	}
}

// GetTransport returns the contained FTransport.
func (f *FServiceProvider) GetTransport() FTransport {
	return f.transport
}

// GetProtocolFactory returns the contained FProtocolFactory.
func (f *FServiceProvider) GetProtocolFactory() *FProtocolFactory {
	return f.protocolFactory
}

// GetMiddleware returns the ServiceMiddleware stored on this FServiceProvider.
func (f *FServiceProvider) GetMiddleware() []ServiceMiddleware {
	middleware := make([]ServiceMiddleware, len(f.middleware))
	copy(middleware, f.middleware)
	return middleware
}
