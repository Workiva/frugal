package frugal

import (
	"fmt"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
)

// Ensures FStatelessNatsServer receives requests and sends back responses on
// the correct subject.
func TestFStatelessNatsServer(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	processor := &processor{t}
	protoFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	server := NewFNatsServerBuilder(conn, processor, protoFactory, "foo").WithQueueGroup("queue").Build()
	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)
	defer server.Stop()

	tr := NewFNatsTransport(conn, "foo", "bar")
	r := &mockFRegistry{}
	r.On("Execute", []byte{0, 0, 0, 3, 102, 111, 111}).Return(nil)
	tr.SetRegistry(r)

	assert.Nil(t, tr.Open())

	// Send a request.
	_, err = tr.Write([]byte("xxxxhello"))
	assert.Nil(t, err)
	assert.Nil(t, tr.Flush())
	time.Sleep(50 * time.Millisecond)
	r.AssertExpectations(t)
}

type processor struct {
	t *testing.T
}

func (p *processor) Process(in, out *FProtocol) error {
	assert.Equal(p.t, "xxxxhello", string(in.Transport().(*thrift.TMemoryBuffer).Bytes()))
	out.WriteString("foo")
	return nil
}
