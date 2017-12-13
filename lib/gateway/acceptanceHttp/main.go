package main

import (
	"fmt"
	"log"
	"net/http"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/rs/cors"

	"github.com/Workiva/frugal/lib/gateway/gen-go/v1/music"
	"github.com/Workiva/frugal/lib/go"
)

// Run an HTTP server
func main() {
	fProtocolFactory := frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())

	handler := &StoreHandler{}
	processor := music.NewFStoreProcessor(handler)

	mux := http.NewServeMux()
	mux.HandleFunc("/frugal", frugal.NewFrugalHandlerFunc(processor, fProtocolFactory))
	corsOptions := cors.Options{
		AllowedHeaders: []string{"Content-Transfer-Encoding"},
	}
	httpHandler := cors.New(corsOptions).Handler(mux)

	fmt.Println("Starting the http server...")
	log.Fatal(http.ListenAndServe(":8080", httpHandler))
}

// StoreHandler handles all incoming requests to the server.
// The handler must satisfy the interface the server exposes.
type StoreHandler struct{}

// LookupAlbum always returns the same album
func (f *StoreHandler) LookupAlbum(frugal.FContext, *music.LookupAlbumRequest) (*music.Album, error) {
	album := &music.Album{
		Artist: "Coeur de Pirates",
		ASIN:   "c54d385a-5024-4f3f-86ef-6314546a7e7f",
		Title:  "Comme des enfants",
	}

	return album, nil
}

// LookupAlbum2 always returns the same album
func (f *StoreHandler) LookupAlbum2(frugal.FContext, *music.LookupAlbumRequest) (*music.Album, error) {
	album := &music.Album{
		Artist: "Coeur de Pirates",
		ASIN:   "c54d385a-5024-4f3f-86ef-6314546a7e7f",
		Title:  "Comme des enfants",
	}

	return album, nil
}

// LookupAlbum3 always returns the same album
func (f *StoreHandler) LookupAlbum3(frugal.FContext, *music.LookupAlbumRequest) (*music.Album, error) {
	album := &music.Album{
		Artist: "Coeur de Pirates",
		ASIN:   "c54d385a-5024-4f3f-86ef-6314546a7e7f",
		Title:  "Comme des enfants",
	}

	return album, nil
}

// BuyAlbum always buys the same album
func (f *StoreHandler) BuyAlbum(frugal.FContext, *music.BuyAlbumRequest) (*music.Album, error) {
	album := &music.Album{
		Artist: "Coeur de Pirates",
		ASIN:   "c54d385a-5024-4f3f-86ef-6314546a7e7f",
		Title:  "Comme des enfants",
	}

	return album, nil
}
