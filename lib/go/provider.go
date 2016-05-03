package frugal

// FScopeProvider produces FScopeTransports and FProtocols for use by pub/sub
// scopes. It does this by wrapping an FScopeTransportFactory and
// FProtocolFactory.
type FScopeProvider struct {
	scopeTransportFactory FScopeTransportFactory
	protocolFactory       *FProtocolFactory
}

// NewFScopeProvider creates a new FScopeProvider using the given factories.
func NewFScopeProvider(t FScopeTransportFactory, p *FProtocolFactory) *FScopeProvider {
	return &FScopeProvider{
		scopeTransportFactory: t,
		protocolFactory:       p,
	}
}

// New returns a new FScopeTransport and TProtocol used by pub/sub scopes.
func (p *FScopeProvider) New() (FScopeTransport, *FProtocol) {
	transport := p.scopeTransportFactory.GetTransport()
	protocol := p.protocolFactory.GetProtocol(transport)
	return transport, protocol
}
