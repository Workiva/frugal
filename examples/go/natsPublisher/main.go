package main

import (
	"fmt"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"

	"github.com/Workiva/frugal/examples/go/gen-go/v1/music"
	"github.com/Workiva/frugal/lib/go"
)

// Run a NATS publisher
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
	defer conn.Close()

	// Create a NATS scoped transport for the PubSub scope
	factory := frugal.NewFNatsScopeTransportFactory(conn)
	provider := frugal.NewFScopeProvider(factory, fProtocolFactory)
	publisher := music.NewAlbumWinnersPublisher(provider)

	// Open the publisher to receive traffic
	if err := publisher.Open(); err != nil {
		panic(err)
	}
	defer publisher.Close()

	// Publish an event
	ctx := frugal.NewFContext("a-corr-id")
	album := &music.Album{
		ASIN:     "c54d385a-5024-4f3f-86ef-6314546a7e7f",
		Duration: 1200,
		Tracks: []*music.Track{&music.Track{
			Title:     "Comme des enfants",
			Artist:    "Coeur de pirate",
			Publisher: "Grosse Boîte",
			Composer:  "Béatrice Martin",
			Duration:  169,
			Pro:       music.PerfRightsOrg_ASCAP,
		}},
	}
	if err := publisher.PublishWinner(ctx, album); err != nil {
		panic(err)
	}

	fmt.Println("WinnerPublished ...")
}
