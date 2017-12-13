package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/examples/go/gen-go/twitter"
	"github.com/Workiva/frugal/lib/gateway"
	"github.com/Workiva/frugal/lib/go"
)

var (
	storeEndpoint = "http://localhost:9090/frugal"
)

// Frugal logging middleware
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

// Create a new Frugal client connected to the store service
func newTwitterClient() *twitter.FTwitterClient {
	// Set the protocol used for serialization.
	// The protocol stack must match between client and server
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	// Create an HTTP transport listening
	httpTransport := frugal.NewFHTTPTransportBuilder(&http.Client{}, storeEndpoint).Build()
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
	storeClient := twitter.NewFTwitterClient(provider, newLoggingMiddleware())

	return storeClient
}

func main() {
	mux := gateway.NewRouter()
	c := newTwitterClient()

	err := twitter.RegisterTwitterServiceHandler(mux, c)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting the gateway server ...")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
