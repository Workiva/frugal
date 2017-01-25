package frugal

import (
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const simpleServerAddr = "localhost:5535"

// Ensures FSimpleServer accepts connections.
func TestSimpleServer(t *testing.T) {
	mockFProcessor := new(mockFProcessor)
	protoFactory := thrift.NewTJSONProtocolFactory()
	fTransportFactory := NewAdapterTransportFactory()
	serverTr, err := thrift.NewTServerSocket(simpleServerAddr)
	if err != nil {
		t.Fatal(err)
	}
	server := NewFSimpleServer(
		mockFProcessor,
		serverTr,
		NewFProtocolFactory(protoFactory),
	)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

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

	ctx := NewFContext("")
	ctx.SetTimeout(5 * time.Millisecond)
	_, err = fTransport.Request(ctx, make([]byte, 10))
	assert.Equal(t, thrift.TIMED_OUT, err.(thrift.TTransportException).TypeId())

	assert.Nil(t, server.Stop())

	mockFProcessor.AssertExpectations(t)
	mockFProcessor.AssertExpectations(t)
}
