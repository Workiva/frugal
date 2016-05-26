package frugal

import (
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFScopeTransportFactory struct {
	mock.Mock
}

func (m *mockFScopeTransportFactory) GetTransport() FScopeTransport {
	return m.Called().Get(0).(FScopeTransport)
}

func TestScopeProviderNew(t *testing.T) {
	mockScopeTransportFactory := new(mockFScopeTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protoFactory := NewFProtocolFactory(mockTProtocolFactory)
	provider := NewFScopeProvider(mockScopeTransportFactory, protoFactory)
	scopeTransport := new(fNatsScopeTransport)
	mockScopeTransportFactory.On("GetTransport").Return(scopeTransport)
	proto := new(thrift.TBinaryProtocol)
	mockTProtocolFactory.On("GetProtocol", scopeTransport).Return(proto)

	transport, protocol := provider.New()
	assert.Equal(t, scopeTransport, transport)
	assert.Equal(t, proto, protocol.TProtocol)
	mockScopeTransportFactory.AssertExpectations(t)
	mockTProtocolFactory.AssertExpectations(t)
}
