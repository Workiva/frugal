package main

import (
	"flag"
	"github.com/Workiva/frugal/test/integration/Go/common"
	"log"
)

var host = flag.String("host", "localhost", "Host to connect")
var port = flag.Int64("port", 9090, "Port number to connect")
var domain_socket = flag.String("domain-socket", "", "Domain Socket (e.g. /tmp/ThriftTest.thrift), instead of host and port")
var transport = flag.String("transport", "buffered", "Transport: buffered, framed, http, zlib")
var protocol = flag.String("protocol", "binary", "Protocol: binary, compact, json")
var ssl = flag.Bool("ssl", false, "Encrypted Transport using SSL")
var certPath = flag.String("certPath", "keys", "Directory that contains SSL certificates")

func main() {
	flag.Parse()
	server, err := integration.StartServer(*host, *port, *domain_socket, *transport, *protocol, *ssl, *certPath, common.PrintingHandler)
	if err != nil {
		log.Fatalln("Unable to start server: ", err)
	}
	server.Server()
}
