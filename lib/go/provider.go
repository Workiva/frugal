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

// FScopeProvider produces FScopeTransports and FProtocols for use by pub/sub
// scopes. It does this by wrapping an FScopeTransportFactory and
// FProtocolFactory. This also provides a shim for adding middleware to a
// publisher or subscriber.
type FScopeProvider struct {
	publisherTransportFactory  FPublisherTransportFactory
	subscriberTransportFactory FSubscriberTransportFactory
	protocolFactory            *FProtocolFactory
	outerMiddleware            []ServiceMiddleware
	innerMiddleware            []ServiceMiddleware
}

// NewFScopeProvider creates a new FScopeProvider using the given factories.
func NewFScopeProvider(pub FPublisherTransportFactory, sub FSubscriberTransportFactory,
	prot *FProtocolFactory, middleware ...ServiceMiddleware) *FScopeProvider {
	return &FScopeProvider{
		publisherTransportFactory:  pub,
		subscriberTransportFactory: sub,
		protocolFactory:            prot,
		outerMiddleware:            middleware,
		innerMiddleware:            []ServiceMiddleware{},
	}
}

// NewFScopeProvider2
// TODO better name
func NewFScopeProvider2(pub FPublisherTransportFactory, sub FSubscriberTransportFactory,
	prot *FProtocolFactory, outerMiddleware []ServiceMiddleware, innerMiddleware []ServiceMiddleware) *FScopeProvider {
	return &FScopeProvider{
		publisherTransportFactory: pub,
		subscriberTransportFactory: sub,
		protocolFactory: prot,
		outerMiddleware: outerMiddleware,
		innerMiddleware: innerMiddleware,
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
// DEPRECATED: replaced by GetOuterMiddleware() for more specificity
func (p *FScopeProvider) GetMiddleware() []ServiceMiddleware {
	return p.GetMiddleware()
}

// GetOuterMiddleware
func (p *FScopeProvider) GetOuterMiddleware() []ServiceMiddleware {
	middleware := make([]ServiceMiddleware, len(p.outerMiddleware))
	copy(middleware, p.outerMiddleware)
	return middleware
}

// GetInnerMiddleware
func (p *FScopeProvider) GetInnerMiddleware() []ServiceMiddleware {
	middleware := make([]ServiceMiddleware, len(p.innerMiddleware))
	copy(middleware, p.innerMiddleware)
	return middleware
}

// FServiceProvider produces FTransports and FProtocolFactories for use by RPC
// service clients. The main purpose of this is to provide a shim for adding
// middleware to a client.
type FServiceProvider struct {
	transport       FTransport
	protocolFactory *FProtocolFactory
	outerMiddleware      []ServiceMiddleware
	innerMiddleware []ServiceMiddleware
}

// NewFServiceProvider creates a new FServiceProvider containing the given
// FTransport and FProtocolFactory.
func NewFServiceProvider(transport FTransport, protocolFactory *FProtocolFactory, middleware ...ServiceMiddleware) *FServiceProvider {
	return &FServiceProvider{
		transport:       transport,
		protocolFactory: protocolFactory,
		outerMiddleware: middleware,
		innerMiddleware: []ServiceMiddleware{},
	}
}

// NewFServiceProvider2
// TODO better name
func NewFServiceProvider2(transport FTransport, prot *FProtocolFactory,
	outerMiddleware []ServiceMiddleware, innerMiddleware []ServiceMiddleware) *FServiceProvider {
	return &FServiceProvider{
		transport: transport,
		protocolFactory: prot,
		outerMiddleware: outerMiddleware,
		innerMiddleware: innerMiddleware,
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
// DEPRECATED: replaced by GetOuterMiddleware() for more specificity
func (p *FServiceProvider) GetMiddleware() []ServiceMiddleware {
	return p.GetMiddleware()
}

// GetOuterMiddleware
func (p *FServiceProvider) GetOuterMiddleware() []ServiceMiddleware {
	middleware := make([]ServiceMiddleware, len(p.outerMiddleware))
	copy(middleware, p.outerMiddleware)
	return middleware
}

// GetInnerMiddleware
func (p *FServiceProvider) GetInnerMiddleware() []ServiceMiddleware {
	middleware := make([]ServiceMiddleware, len(p.innerMiddleware))
	copy(middleware, p.innerMiddleware)
	return middleware
}
