package integration

import (
	"strconv"
	"testing"

	"github.com/nats-io/nats"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/frugal/lib/go"
)

func TestPublishSubscribe(t *testing.T) {
	CheckShort(t)

	protocolFactories := map[string]thrift.TProtocolFactory{
		"TCompactProtocolFactory":       thrift.NewTCompactProtocolFactory(),
		"TJSONProtocolFactory":          thrift.NewTJSONProtocolFactory(),
		"TBinaryProtocolFactoryDefault": thrift.NewTBinaryProtocolFactoryDefault(),
	}
	ftransportFactory := frugal.NewFMuxTransportFactory(5)

	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{nats.DefaultURL}
	natsOptions.Secure = false // TODO: Test with TLS enabled
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for name, protocolFactory := range protocolFactories {
		fprotocolFactory := frugal.NewFProtocolFactory(protocolFactory)
		testPublishSubscribe(t, fprotocolFactory, ftransportFactory, conn, name)
	}
}

func testPublishSubscribe(
	t *testing.T,
	protocolFactory *frugal.FProtocolFactory,
	transportFactory frugal.FTransportFactory,
	conn *nats.Conn,
	// Name is used to display the protocol in each test case
	name string,
) {
	factory := frugal.NewFNatsScopeTransportFactory(conn)
	provider := frugal.NewFScopeProvider(factory, protocolFactory)
	publisherMiddlewareCalled := make(chan bool, 1)
	publisherMiddleware := newMiddleware(publisherMiddlewareCalled)
	publisher := event.NewEventsPublisher(provider, publisherMiddleware)
	if err := publisher.Open(); err != nil {
		panic(err)
	}
	defer publisher.Close()

	subscriberMiddlewareCalled := make(chan bool, 1)
	subscriberMiddleware := newMiddleware(subscriberMiddlewareCalled)
	subscriber := event.NewEventsSubscriber(provider, subscriberMiddleware)

	started := make(chan bool)
	wait := make(chan bool)
	done := make(chan struct{})

	expected := new(expectedMessages)
	expected.messageList = make(map[event.Event]bool)

	for i := 1; i < 6; i++ {
		expected.messageList[event.Event{Message: "message" + strconv.Itoa(i)}] = false
	}

	go messageHandler(t, subscriber, started, wait, done, expected, name)
	<-started

	// Publish events
	for i := 1; i < 6; i++ {
		ctx := frugal.NewFContext(strconv.Itoa(i))
		evt := &event.Event{Message: "message" + strconv.Itoa(i)}
		if err := publisher.PublishEventCreated(ctx, name, evt); err != nil {
			panic(err)
		}
	}

	select {
	case <-publisherMiddlewareCalled:
	default:
		t.Fatal("Publisher middleware not invoked")
	}

	<-done

	select {
	case <-subscriberMiddlewareCalled:
	default:
		t.Fatal("Subscriber middleware not invoked")
	}
}
