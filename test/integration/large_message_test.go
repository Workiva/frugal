package integration

import (
	"flag"
	"log"
	"testing"

	"github.com/nats-io/nats"

	"git.apache.org/thrift.git/lib/go/thrift"
)

func TestLargeMessage(t *testing.T) {

	protocolFactories := map[string]thrift.TProtocolFactory{
		"TCompactProtocolFactory":       thrift.NewTCompactProtocolFactory(),
		"TSimpleJSONProtocolFactory":    thrift.NewTSimpleJSONProtocolFactory(),
		"TJSONProtocolFactory":          thrift.NewTJSONProtocolFactory(),
		"TBinaryProtocolFactoryDefault": thrift.NewTBinaryProtocolFactoryDefault(),
	}
	transportFactories := map[string]thrift.TTransportFactory{
		"TBufferedTransportFactory": thrift.NewTBufferedTransportFactory(8192),
		"TTransportFactory":         thrift.NewTTransportFactory(),
	}

	// If framed is to be tested, it is a separate option for TransportFactories
	// thrift.NewTFramedTransportFactory(transportFactory),

	tls := false // This will need to match the server configuration

	addr := flag.String("addr", nats.DefaultURL, "NATS address")
	// Probably need to set this to true for additional testing
	secure := flag.Bool("secure", tls, "Use tls secure transport")

	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{*addr}
	natsOptions.Secure = *secure
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}

	for pf, protocolFactory := range protocolFactories {
		for tf, transportFactory := range transportFactories {
			log.Printf("Testing with %v and %v.", pf, tf)
			name := pf + " " + tf
			LargeMessage(t, protocolFactory, transportFactory, conn, name)
		}
	}

}
