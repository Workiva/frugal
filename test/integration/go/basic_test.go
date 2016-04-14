package integration

import (
	"sync"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/example/go/gen-go/base"
	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/frugal/lib/go"
)

const addr = "localhost:4535"

func newMiddleware(called *bool) frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service, method string, args frugal.Arguments) frugal.Results {
			*called = true
			return next(service, method, args)
		}
	}
}

func TestBasic(t *testing.T) {
	protoFactories := []thrift.TProtocolFactory{
		thrift.NewTCompactProtocolFactory(),
		thrift.NewTJSONProtocolFactory(),
		thrift.NewTBinaryProtocolFactoryDefault(),
	}
	fTransportFactory := frugal.NewFMuxTransportFactory(2)

	for _, protoFactory := range protoFactories {
		testBasic(t, protoFactory, fTransportFactory)
	}
}

func testBasic(t *testing.T, protoFactory thrift.TProtocolFactory, fTransportFactory frugal.FTransportFactory) {
	// Setup server.
	serverMiddlewareCalled := false
	processor := event.NewFFooProcessor(&FooHandler{}, newMiddleware(&serverMiddlewareCalled))
	serverTr, err := thrift.NewTServerSocket(addr)
	if err != nil {
		t.Fatal(err)
	}
	server := frugal.NewFSimpleServerFactory5(
		frugal.NewFProcessorFactory(processor),
		serverTr,
		frugal.NewFMuxTransportFactory(2),
		frugal.NewFProtocolFactory(protoFactory),
	)

	// Start server.
	ch := make(chan struct{})
	go func() {
		ch <- struct{}{}
		if err := server.Serve(); err != nil {
			t.Fatal("Failed to start server:", err.Error())
		}
	}()
	<-ch

	// Setup client.
	transport, err := thrift.NewTSocket(addr)
	if err != nil {
		t.Fatal(err)
	}
	fTransport := fTransportFactory.GetTransport(transport)
	defer fTransport.Close()
	if err := fTransport.Open(); err != nil {
		t.Fatal(err)
	}
	clientMiddlewareCalled := false
	client := event.NewFFooClient(fTransport, frugal.NewFProtocolFactory(protoFactory), newMiddleware(&clientMiddlewareCalled))

	runClient(t, client)

	if !serverMiddlewareCalled {
		t.Fatal("Server middleware not invoked")
	}
	if !clientMiddlewareCalled {
		t.Fatal("Client middleware not invoked")
	}

	if err := server.Stop(); err != nil {
		t.Fatal("Failed to stop server:", err.Error())
	}
}

func runClient(t *testing.T, client *event.FFooClient) {
	ctx := frugal.NewFContext("")
	ctx.SetTimeout(2 * time.Millisecond)

	// First Ping should timeout.
	if err := client.Ping(ctx); err != frugal.ErrTimeout {
		t.Fatal("Expected Ping timeout:", err)
	}

	for i := 0; i < 5; i++ {
		if err := client.Ping(frugal.NewFContext("")); err != nil {
			t.Fatal("Ping error:", err.Error())
		}
	}

	if err := client.BasePing(frugal.NewFContext("")); err != nil {
		t.Fatal("BasePing error:", err.Error())
	}

	// First Blah call should throw AwesomeException.
	_, err := client.Blah(frugal.NewFContext(""), 1, "hello", event.NewEvent())
	if err == nil {
		t.Fatal("Expected Blah AwesomeException error")
	}
	if awe, ok := err.(*event.AwesomeException); !ok {
		t.Fatal("Expected Blah to return AwesomeException, got:", err)
	} else {
		if awe.ID != 42 {
			t.Fatal("Expected ID 42, got:", awe.ID)
		}
		if awe.Reason != "error" {
			t.Fatal("Expected reason 'error', got:", awe.Reason)
		}
	}

	// Second Blah call should throw APIException.
	_, err = client.Blah(frugal.NewFContext(""), 2, "hello", event.NewEvent())
	if err == nil {
		t.Fatal("Expected Blah APIException error")
	}
	if _, ok := err.(*base.APIException); !ok {
		t.Fatal("Expected Blah to return APIException, got:", err)
	}

	r, err := client.Blah(frugal.NewFContext(""), 3, "hello", event.NewEvent())
	if err != nil {
		t.Fatal("Blah error:", err.Error())
	}
	if r != 42 {
		t.Fatal("Expected Blah to return 42, got:", r)
	}

	if err := client.OneWay(frugal.NewFContext(""), 100, event.Request{1: "foo"}); err != nil {
		t.Fatal("OneWay error:", err.Error())
	}
}

type FooHandler struct {
	pingCount int
	mu        sync.Mutex
}

func (f *FooHandler) Ping(ctx *frugal.FContext) error {
	f.mu.Lock()
	f.pingCount++
	c := f.pingCount
	f.mu.Unlock()
	if c == 1 {
		time.Sleep(5 * time.Second)
	}
	return nil
}

func (f *FooHandler) Blah(ctx *frugal.FContext, num int32, str string, e *event.Event) (int64, error) {
	if num == 1 {
		awe := event.NewAwesomeException()
		awe.ID = 42
		awe.Reason = "error"
		return 0, awe
	} else if num == 2 {
		return 0, base.NewAPIException()
	}
	return 42, nil
}

func (f *FooHandler) BasePing(ctx *frugal.FContext) error {
	return nil
}

func (f *FooHandler) OneWay(ctx *frugal.FContext, id event.ID, req event.Request) error {
	return nil
}
