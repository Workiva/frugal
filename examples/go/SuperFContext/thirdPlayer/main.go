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
)

// Run an HTTP server
func main() {
	// Set the protocol used for serialization.
	// The protocol stack must match between client and server
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	// Create a handler. Each incoming request at the processor is sent to
	// the handler. Responses from the handler are returned back to the
	// client
	handler := &LastPlayerTelephoneHandler{}
	processor := telephonegame.NewFTelephoneGameProcessor(handler, middleware.CheckDeadlineContextMiddleware())

	// Start the server using the configured processor, and protocol
	http.HandleFunc("/frugal", frugal.NewFrugalHandlerFunc(processor, fProtocolFactory))
	fmt.Println("Listening for the second player...")
	log.Fatal(http.ListenAndServe(":9091", http.DefaultServeMux))
}

// StoreHandler handles all incoming requests to the server.
// The handler must satisfy the interface the server exposes.
type LastPlayerTelephoneHandler struct{}

func (f *LastPlayerTelephoneHandler) PassOnMessage(ctx frugal.FContext, word string) (string, error){
	newLetterIndex := rand.Intn(len(word))
	letterIndex := rand.Intn(len(word) - 1)
	newWord := word[:newLetterIndex-1] + string(word[letterIndex]) + word[newLetterIndex:]
	return newWord, nil
}
