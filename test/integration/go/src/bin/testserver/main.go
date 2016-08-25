package main

import (
	"flag"

	"github.com/Workiva/frugal/test/integration/go/common"
)

var host = flag.String("host", "localhost", "Host to connect")
var port = flag.Int64("port", 9090, "Port number to connect")
var transport = flag.String("transport", "stateless", "Transport: stateless, stateful, stateless-stateful, http")
var protocol = flag.String("protocol", "binary", "Protocol: binary, compact, json")

func main() {
	flag.Parse()
	common.StartServer(*host, *port, *transport, *protocol, common.PrintingHandler)
}
