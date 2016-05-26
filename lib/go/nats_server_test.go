package frugal

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

type mockFTransportFactory struct {
	mock.Mock
}

func (m *mockFTransportFactory) GetTransport(tr thrift.TTransport) FTransport {
	return m.Called(tr).Get(0).(FTransport)
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

// Ensures FNatsServer accepts connections.
func TestNatsServerHappyPath(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServer(conn, "foo", 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)
	mockTransport := new(mockFTransport)
	mockTransport.On("SetRegistry", mock.Anything).Return(nil)
	mockTransport.On("SetHighWatermark", defaultWatermark).Return(nil)
	mockTransport.On("Open").Return(nil)
	mockTransport.On("Closed").Return(toRecvChan(make(chan error)))
	mockTransport.On("Close").Return(nil)
	mockTransportFactory.On("GetTransport", mock.AnythingOfType("*frugal.natsServiceTTransport")).Return(mockTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mockTransport).Return(proto)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	client := NewNatsServiceTTransport(conn, "foo", time.Second, 2)
	assert.Nil(t, client.Open())
	defer client.Close()

	assert.Nil(t, server.Stop())
	time.Sleep(5 * time.Millisecond)

	mockTransport.AssertExpectations(t)
	mockTransportFactory.AssertExpectations(t)
	mockTProtocolFactory.AssertExpectations(t)
}

// Ensures FNatsServer accepts connections using the specified high watermark.
func TestNatsServerHappyPathHighWatermark(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServer(conn, "foo", 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)
	server.SetHighWatermark(time.Second)
	mockTransport := new(mockFTransport)
	mockTransport.On("SetRegistry", mock.Anything).Return(nil)
	mockTransport.On("SetHighWatermark", time.Second).Return(nil)
	mockTransport.On("Open").Return(nil)
	mockTransport.On("Closed").Return(toRecvChan(make(chan error)))
	mockTransport.On("Close").Return(nil)
	mockTransportFactory.On("GetTransport", mock.AnythingOfType("*frugal.natsServiceTTransport")).Return(mockTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mockTransport).Return(proto)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	client := NewNatsServiceTTransport(conn, "foo", time.Second, 2)
	assert.Nil(t, client.Open())
	defer client.Close()

	assert.Nil(t, server.Stop())
	time.Sleep(5 * time.Millisecond)

	mockTransport.AssertExpectations(t)
	mockTransportFactory.AssertExpectations(t)
	mockTProtocolFactory.AssertExpectations(t)
}

// Ensures FNatsServer accepts connections on multiple subjects when supplied.
func TestNatsServerHappyPathMultiSubjects(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServerWithSubjects(conn, []string{"foo", "bar"}, 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)
	mockTransport := new(mockFTransport)
	mockTransport.On("SetRegistry", mock.Anything).Return(nil)
	mockTransport.On("SetHighWatermark", defaultWatermark).Return(nil)
	mockTransport.On("Open").Return(nil)
	mockTransport.On("Closed").Return(toRecvChan(make(chan error)))
	mockTransport.On("Close").Return(nil)
	mockTransportFactory.On("GetTransport", mock.AnythingOfType("*frugal.natsServiceTTransport")).Return(mockTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mock.AnythingOfType("*frugal.mockFTransport")).Return(proto)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	client := NewNatsServiceTTransport(conn, "foo", time.Second, 2)
	assert.Nil(t, client.Open())
	defer client.Close()
	client = NewNatsServiceTTransport(conn, "bar", time.Second, 2)
	assert.Nil(t, client.Open())
	defer client.Close()

	assert.Nil(t, server.Stop())
	time.Sleep(5 * time.Millisecond)

	mockTransport.AssertExpectations(t)
	mockTransportFactory.AssertExpectations(t)
	mockTProtocolFactory.AssertExpectations(t)
}

// Ensures Serve returns an error if the subscribe fails.
func TestNatsServerServeSubscribeError(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServerWithSubjects(conn, []string{"foo", "bar"}, 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)

	conn.Close()
	assert.Error(t, server.Serve())
}

// Ensures FNatsServer discards connect messages with no reply subject.
func TestNatsServerDiscardInvalidConnectNoReply(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServer(conn, "foo", 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	conn.Publish("foo", []byte("{}"))
	conn.Flush()
	time.Sleep(5 * time.Millisecond)

	assert.Nil(t, server.Stop())
}

// Ensures FNatsServer discards invalid connect messages.
func TestNatsServerDiscardInvalidConnect(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServer(conn, "foo", 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	conn.PublishRequest("foo", "reply", []byte("invalid connect message"))
	conn.Flush()
	time.Sleep(5 * time.Millisecond)

	assert.Nil(t, server.Stop())
}

// Ensures FNatsServer discards connect messages with an unsupported version.
func TestNatsServerDiscardInvalidConnectBadVersion(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFNatsServer(conn, "foo", 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	conn.PublishRequest("foo", "reply", []byte(`{"version": 42}`))
	conn.Flush()
	time.Sleep(5 * time.Millisecond)

	assert.Nil(t, server.Stop())
}

// Ensure FNatsServer discards failed connections.
func TestNatsServerAcceptError(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	mockTransport := new(mockFTransport)
	mockTransport.On("SetRegistry", mock.Anything).Return(nil)
	mockTransport.On("SetHighWatermark", defaultWatermark).Return(nil)
	mockTransport.On("Open").Return(errors.New("error"))
	mockTransportFactory.On("GetTransport", mock.AnythingOfType("*frugal.natsServiceTTransport")).Return(mockTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mockTransport).Return(proto)
	server := NewFNatsServer(conn, "foo", 5*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	conn.PublishRequest("foo", "reply", []byte(`{"version": 0}`))
	conn.Flush()
	time.Sleep(5 * time.Millisecond)

	assert.Nil(t, server.Stop())
}

// Ensure FNatsServer removes expired connections.
func TestNatsServerExpiredConnections(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	mockProcessor := new(mockFProcessor)
	mockTransportFactory := new(mockFTransportFactory)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	mockTransport := new(mockFTransport)
	mockTransport.On("SetRegistry", mock.Anything).Return(nil)
	mockTransport.On("SetHighWatermark", defaultWatermark).Return(nil)
	mockTransport.On("Open").Return(nil)
	mockTransport.On("Closed").Return(toRecvChan(make(chan error)))
	mockTransport.On("Close").Return(nil)
	mockTransportFactory.On("GetTransport", mock.AnythingOfType("*frugal.natsServiceTTransport")).Return(mockTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mockTransport).Return(proto)
	server := NewFNatsServer(conn, "foo", 1*time.Millisecond, mockProcessor,
		mockTransportFactory, protocolFactory)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	conn.PublishRequest("foo", "reply", []byte(`{"version": 0}`))
	conn.Flush()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, len(server.(*FNatsServer).clients))

	assert.Nil(t, server.Stop())
}

func toRecvChan(c chan error) <-chan error {
	return c
}
