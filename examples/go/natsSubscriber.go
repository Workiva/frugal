package main

import (
	"fmt"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/go-nats"

	"github.com/Workiva/frugal/examples/go/gen-go/v1/music"
	"github.com/Workiva/frugal/lib/go"
)

// Run a NATS subscriber
func main() {
	// Set the protocol used for serialization.
	// The protocol stack must match between client and server
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	// Setup a NATS connection (using default options)
	natsOptions := nats.DefaultOptions
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}

	// Create a NATS scoped transport for the PubSub scope
	pfactory := frugal.NewFNatsPublisherTransportFactory(conn)
	sfactory := frugal.NewFNatsSubscriberTransportFactory(conn)
	provider := frugal.NewFScopeProvider(pfactory, sfactory, fProtocolFactory)
	subscriber := music.NewAlbumWinnersSubscriber(provider)

	// Subscribe to messages
	var wg sync.WaitGroup
	wg.Add(2)

	subscriber.SubscribeWinner(func(ctx frugal.FContext, m *music.Album) {
		fmt.Printf("received %+v : %+v\n", ctx, m)
		wg.Done()
	})
	subscriber.SubscribeContestStart(func(ctx frugal.FContext, albums []*music.Album) {
		fmt.Printf("received %+v : %+v\n", ctx, albums)
		wg.Done()
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Subscriber started...")
	wg.Wait()
}
