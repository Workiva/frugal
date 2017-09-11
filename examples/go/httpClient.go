package main

import (
	"fmt"
	"net/http"
	"reflect"

	"git.apache.org/thrift.git/lib/go/thrift"

	"github.com/Workiva/frugal/examples/go/gen-go/v1/music"
	"github.com/Workiva/frugal/lib/go"
)

// Run a NATS client
func main() {
	// Set the protocol used for serialization.
	// The protocol stack must match between client and server
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	// Create an HTTP transport listening
	httpTransport := frugal.NewFHTTPTransportBuilder(&http.Client{}, "http://localhost:9090/frugal").Build()
	defer httpTransport.Close()
	if err := httpTransport.Open(); err != nil {
		panic(err)
	}

	// Create a provider with the transport and protocol factory. The provider
	// can be used to create multiple Clients.
	provider := frugal.NewFServiceProvider(httpTransport, fProtocolFactory)

	// Create a client used to send messages with our desired protocol.  You
	// can also pass middleware in here if you only want it to intercept calls
	// for this specific client.
	storeClient := music.NewFStoreClient(provider, newLoggingMiddleware())

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
