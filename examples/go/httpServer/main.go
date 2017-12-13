package main

import (
	"fmt"
	"log"
	"net/http"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/rs/cors"

	"github.com/Workiva/frugal/examples/go/gen-go/twitter"
	"github.com/Workiva/frugal/lib/go"
)

// Run an HTTP server
func main() {
	// Set the protocol used for serialization.
	// The protocol stack must match between client and server
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	// Create a handler. Each incoming request at the processor is sent to
	// the handler. Responses from the handler are returned back to the
	// client
	handler := &TwitterHandler{}
	processor := twitter.NewFTwitterProcessor(handler)

	// Start the server using the configured processor, and protocol
	mux := http.NewServeMux()
	mux.HandleFunc("/frugal", frugal.NewFrugalHandlerFunc(processor, fProtocolFactory))
	corsOptions := cors.Options{
		AllowedHeaders: []string{"Content-Transfer-Encoding"},
	}
	httpHandler := cors.New(corsOptions).Handler(mux)

	fmt.Println("Starting the http server...")
	log.Fatal(http.ListenAndServe(":9090", httpHandler))
}

// TwitterHandler handles all incoming requests to the server.
// The handler must satisfy the interface the server exposes.
type TwitterHandler struct{}

// CreateTweet creates a new tweet
func (f *TwitterHandler) CreateTweet(ctx frugal.FContext, tweet *twitter.Tweet) (err error) {
	return nil
}

// SearchTweets searches for a bunch of tweets
func (f *TwitterHandler) SearchTweets(ctx frugal.FContext, query string) (r *twitter.TweetSearchResult_, err error) {
	return nil, nil
}

// DeleteTweet deletes the tweeit with id
func (f *TwitterHandler) DeleteTweet(ctx frugal.FContext, tweet_id int32) (err error) { return nil }

// UpdateTweet updates the tweeit with id
func (f *TwitterHandler) UpdateTweet(ctx frugal.FContext, tweet_id int32) (r *twitter.Tweet, err error) {
	return nil, nil
}
