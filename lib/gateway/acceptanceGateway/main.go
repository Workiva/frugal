package main

import (
	"fmt"
	"log"
	"net/http"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/gateway"
	"github.com/Workiva/frugal/lib/gateway/gen-go/v1/music"
	"github.com/Workiva/frugal/lib/go"
)

var (
	storeEndpoint = "http://localhost:8080/frugal"
)

// Run an HTTP+JSON gateway
func main() {
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	httpTransport := frugal.NewFHTTPTransportBuilder(&http.Client{}, storeEndpoint).Build()
	defer httpTransport.Close()
	if err := httpTransport.Open(); err != nil {
		panic(err)
	}

	provider := frugal.NewFServiceProvider(httpTransport, fProtocolFactory)
	c := music.NewFStoreClient(provider)

	mux := gateway.NewRouter()

	var marshaler gateway.Marshaler
	err := music.RegisterStoreServiceHandler(marshaler, mux, c)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting the gateway server ...")
	log.Fatal(http.ListenAndServe(":9001", mux))
}
