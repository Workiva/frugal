package main

import (
	"fmt"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"

	"github.com/Workiva/frugal/example/go/gen-go/music"
	"github.com/Workiva/frugal/lib/go"
)

// Run a NATS client
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

	// Create a NATS transport listening on the music-service topic.
	// Configured with own inbox
	natsT := frugal.NewFNatsTransport(conn, "music-service", "service-inbox")
	defer natsT.Close()
	if err := natsT.Open(); err != nil {
		panic(err)
	}

	// Create a client using NATS to send messages with our desired
	// protocol
	storeClient := music.NewFStoreClient(natsT, fProtocolFactory)

	// Configure the context used for sending requests
	ctx := frugal.NewFContext("a-corr-id")

	// Request to buy an album
	album, err := storeClient.BuyAlbum(ctx, "ASIN-1290AIUBOA89", "ACCOUNT-12345")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Bought an album %s\n", album)

	// Enter the contest
	storeClient.EnterAlbumGiveaway(ctx, "kevin@workiva.com", "Kevin")
}
