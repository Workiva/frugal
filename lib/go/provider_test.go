package frugal

import (
	"sync"
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

type mockTProtocolFactory struct {
	mock.Mock
	sync.Mutex
}

func (m *mockTProtocolFactory) GetProtocol(tr thrift.TTransport) thrift.TProtocol {
	m.Lock()
	defer m.Unlock()
	return m.Called(tr).Get(0).(thrift.TProtocol)
}

func (m *mockTProtocolFactory) AssertExpectations(t *testing.T) {
	m.Lock()
	defer m.Unlock()
	m.Mock.AssertExpectations(t)
}

type mockFProcessor struct {
	mock.Mock
	sync.Mutex
}

func (m *mockFProcessor) Process(in, out *FProtocol) error {
	m.Lock()
	defer m.Unlock()
	return m.Called(in, out).Error(0)
}

func (m *mockFProcessor) AssertExpectations(t *testing.T) {
	m.Lock()
	defer m.Unlock()
	m.Mock.AssertExpectations(t)
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
