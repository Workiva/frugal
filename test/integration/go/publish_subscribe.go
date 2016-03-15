package integration

import (
	"strconv"
	"testing"

	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/frugal/lib/go"
	"github.com/nats-io/nats"
)

func PublishSubscribe(
	t *testing.T,
	protocolFactory *frugal.FProtocolFactory,
	transportFactory frugal.FTransportFactory,
	conn *nats.Conn,
	// Name is used to display the protocol in each test case
	name string,
) {
	factory := frugal.NewFNatsScopeTransportFactory(conn)
	provider := frugal.NewFScopeProvider(factory, protocolFactory)
	publisher := event.NewEventsPublisher(provider)
	if err := publisher.Open(); err != nil {
		panic(err)
	}
	defer publisher.Close()

	subscriber := event.NewEventsSubscriber(provider)

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

	<-done
}
