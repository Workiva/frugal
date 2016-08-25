package main

import (
	"fmt"
	"reflect"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"

	"github.com/Workiva/frugal/examples/go/gen-go/v1/music"
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
	storeClient := music.NewFStoreClient(natsT, fProtocolFactory, newLoggingMiddleware())

	// Request to buy an album
	album, err := storeClient.BuyAlbum(frugal.NewFContext("corr-id-1"), "ASIN-1290AIUBOA89", "ACCOUNT-12345")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Bought an album %s\n", album)

	// Enter the contest
	storeClient.EnterAlbumGiveaway(frugal.NewFContext("corr-id-2"), "kevin@workiva.com", "Kevin")
}

func newLoggingMiddleware() frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			fmt.Printf("==== CALLING %s.%s ====\n", service.Type(), method.Name)
			ret := next(service, method, args)
			fmt.Printf("==== CALLED  %s.%s ====\n", service.Type(), method.Name)
			return ret
		}
	}
}
