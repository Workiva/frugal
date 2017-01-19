package main

import (
	"fmt"
	"net/http"

	"git.apache.org/thrift.git/lib/go/thrift"

	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/examples/go/gen-go/telephonegame"
)

// Run a NATS client
func main() {
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	// Create an HTTP transport listening
	httpTransport := frugal.NewFHTTPTransportBuilder(&http.Client{}, "http://localhost:9090/frugal").Build()
	defer httpTransport.Close()
	if err := httpTransport.Open(); err != nil {
		panic(err)
	}

	word := "Super Frugal Context"
	telephoneGameClient := telephonegame.NewFTelephoneGameClient(frugal.NewFServiceProvider(httpTransport, fProtocolFactory, frugal.NewContextMiddleware()))

	endingWord, err := telephoneGameClient.PassOnMessage(frugal.NewFContext("corr-id-1"), word)
	if err != nil {
		panic(err)
	}

	fmt.Printf("I said:\n%s\nand the player heard:\n%s\n\n", word, endingWord)
}
