package main

import (
	"flag"
	"log"
	"time"

	"github.com/Workiva/frugal/test/integration/go/common"
)

var host = flag.String("host", "localhost", "Host to connect")
var port = flag.Int64("port", 9090, "Port number to connect")
var transport = flag.String("transport", "stateless", "Transport: stateless, stateful, stateless-stateful, http")
var protocol = flag.String("protocol", "binary", "Protocol: binary, compact, json")

func main() {
	flag.Parse()

	serverMiddlewareCalled := make(chan bool, 1)
	pubSubResponseSent := make(chan bool, 1)
	go common.StartServer(
		*host,
		*port,
		*transport,
		*protocol,
		common.PrintingHandler,
		serverMiddlewareCalled,
		pubSubResponseSent)

	// This matches the Java client timeout, which is the highest client timeout in the cross language tests
	timeout := time.After(time.Second * 20)

	select {
	case <-pubSubResponseSent:
		log.Println("Pub/Sub response sent")
	case <-timeout:
		log.Fatal("Pub/Sub response not sent within 20 seconds")
	}

	select {
	case <-serverMiddlewareCalled:
		log.Println("Server middleware called successfully")
	case <-timeout:
		log.Fatalf("Server middleware not called within 20 seconds")
	}

	// The cross runner takes care of killing the server. Tests will fail if the server dies before the cross runner
	// terminates it
	blockForever()
}

func blockForever() {
	select{ }
}
