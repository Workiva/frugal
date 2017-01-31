package main

import (
	"fmt"
	"log"
	"net/http"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/examples/go/gen-go/telephonegame"
	"math/rand"
	"github.com/Workiva/frugal/examples/go/SuperFContext"
	"time"
)

// Run an HTTP server
func main() {
	// Set the protocol used for serialization.
	// The protocol stack must match between client and server
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	// Create a handler. Each incoming request at the processor is sent to
	// the handler. Responses from the handler are returned back to the
	// client
	handler := &TelephoneHandler{}
	processor := telephonegame.NewFTelephoneGameProcessor(handler, middleware.AddTimeoutContextMiddleware(time.Duration(-1) * time.Hour))

	// Start the server using the configured processor, and protocol
	http.HandleFunc("/frugal", frugal.NewFrugalHandlerFunc(processor, fProtocolFactory))
	fmt.Println("Listening for the first player...")
	log.Fatal(http.ListenAndServe(":9090", http.DefaultServeMux))
}

type TelephoneHandler struct{}

func (f *TelephoneHandler) PassOnMessage(ctx frugal.FContext, word string) (string, error){
	newLetterIndex := rand.Intn(len(word))
	letterIndex := rand.Intn(len(word) - 1)
	newWord := word[:newLetterIndex-1] + string(word[letterIndex]) + word[newLetterIndex:]

	httpTransport := frugal.NewFHTTPTransportBuilder(&http.Client{}, "http://localhost:9091/frugal").Build()
	defer httpTransport.Close()
	if err := httpTransport.Open(); err != nil {
		panic(err)
	}

	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	telephoneGameClient := telephonegame.NewFTelephoneGameClient(frugal.NewFServiceProvider(httpTransport, fProtocolFactory))

	newWord, err := telephoneGameClient.PassOnMessage(ctx, newWord)
	if err != nil {
		panic(err)
	}

	return newWord, nil
}
