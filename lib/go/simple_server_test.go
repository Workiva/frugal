package frugal

import (
	"sync"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const simpleServerAddr = "localhost:5535"

type mockFProcessorFactory struct {
	mock.Mock
	sync.Mutex
}

func (m *mockFProcessorFactory) GetProcessor(tr thrift.TTransport) FProcessor {
	m.Lock()
	defer m.Unlock()
	return m.Called(tr).Get(0).(FProcessor)
}

func (m *mockFProcessorFactory) AssertExpectations(t *testing.T) {
	m.Lock()
	defer m.Unlock()
	m.Mock.AssertExpectations(t)
}

// Ensures FSimpleServer accepts connections.
func TestSimpleServer(t *testing.T) {
	mockFProcessorFactory := new(mockFProcessorFactory)
	protoFactory := thrift.NewTJSONProtocolFactory()
	fTransportFactory := NewAdapterTransportFactory()
	serverTr, err := thrift.NewTServerSocket(simpleServerAddr)
	if err != nil {
		t.Fatal(err)
	}
	server := NewFSimpleServerFactory4(
		mockFProcessorFactory,
		serverTr,
		NewAdapterTransportFactory(),
		NewFProtocolFactory(protoFactory),
	)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	mockFProcessor := new(mockFProcessor)
	mockFProcessorFactory.Lock() // IDK why this is needed to prevent races...
	mockFProcessorFactory.On("GetProcessor", mock.AnythingOfType("*thrift.TSocket")).Return(mockFProcessor)
	mockFProcessorFactory.Unlock()
	mockFProcessor.On("Process", mock.AnythingOfType("*frugal.FProtocol"),
		mock.AnythingOfType("*frugal.FProtocol")).Return(nil)

	transport, err := thrift.NewTSocket(simpleServerAddr)
	if err != nil {
		t.Fatal(err)
	}
	fTransport := fTransportFactory.GetTransport(transport)
	defer fTransport.Close()
	if err := fTransport.Open(); err != nil {
		t.Fatal(err)
	}

	_, err = fTransport.Write(make([]byte, 10))
	assert.Nil(t, err)
	assert.Nil(t, fTransport.Flush())
	time.Sleep(5 * time.Millisecond)

	assert.Nil(t, server.Stop())

	mockFProcessorFactory.AssertExpectations(t)
	mockFProcessor.AssertExpectations(t)
}
