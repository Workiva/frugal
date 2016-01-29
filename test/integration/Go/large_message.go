package integration

import (
	"testing"

	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/frugal/lib/go"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
)

func LargeMessage(
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

	evt := event.NewEvent()
	ctx := frugal.NewFContext("Context")

	// Publish event
	evt.Message = WarAndPeace
	if err := publisher.PublishEventCreated(ctx, name, evt); err != nil {
		assert.Error(t, err, "*event.Event.Message (2) field write error: Message is too large")
	} else {
		t.Errorf("Sending message larger than 1MB succeeded on %v", name)
	}
}
