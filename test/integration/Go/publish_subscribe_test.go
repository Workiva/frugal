package integration

import (
	"testing"

	"github.com/nats-io/nats"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/go"
)

func TestPublishSubscribe(t *testing.T) {
	CheckShort(t)

	protocolFactories := map[string]thrift.TProtocolFactory{
		"TCompactProtocolFactory": thrift.NewTCompactProtocolFactory(),
		// "TSimpleJSONProtocolFactory":    thrift.NewTSimpleJSONProtocolFactory(),
		"TJSONProtocolFactory":          thrift.NewTJSONProtocolFactory(),
		"TBinaryProtocolFactoryDefault": thrift.NewTBinaryProtocolFactoryDefault(),
	}
	ftransportFactory := frugal.NewFMuxTransportFactory(5)

	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{nats.DefaultURL}
	natsOptions.Secure = false // TODO: Test with TLS enabled
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for name, protocolFactory := range protocolFactories {
		fprotocolFactory := frugal.NewFProtocolFactory(protocolFactory)
		PublishSubscribe(t, fprotocolFactory, ftransportFactory, conn, name)
	}
}
