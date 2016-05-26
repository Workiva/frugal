package common

import (
	"flag"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/test/integration/go/gen/frugaltest"
)

var (
	debugServerProtocol bool
)

func init() {
	flag.BoolVar(&debugServerProtocol, "debug_server_protocol", false, "turn server protocol trace on")
}

func StartServer(
	host string,
	port int64,
	domain_socket string,
	transport string,
	protocol string,
	certPath string,
	handler frugaltest.FFrugalTest) (srv *frugal.FSimpleServer, err error) {

	hostPort := fmt.Sprintf("%s:%d", host, port)

	var protocolFactory thrift.TProtocolFactory
	switch protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		return nil, fmt.Errorf("Invalid protocol specified %s", protocol)
	}

	if debugServerProtocol {
		protocolFactory = thrift.NewTDebugProtocolFactory(protocolFactory, "server:")
	}

	var serverTransport thrift.TServerTransport
	if domain_socket != "" {
		serverTransport, err = thrift.NewTServerSocket(domain_socket)
	} else {
		serverTransport, err = thrift.NewTServerSocket(hostPort)
	}
	if err != nil {
		return nil, err
	}

	fTransportFactory := frugal.NewFMuxTransportFactory(2)
	processor := frugaltest.NewFFrugalTestProcessor(handler)
	server := frugal.NewFSimpleServerFactory5(
		frugal.NewFProcessorFactory(processor),
		serverTransport,
		fTransportFactory,
		frugal.NewFProtocolFactory(protocolFactory))

	if err = server.Listen(); err != nil {
		return
	}
	go server.AcceptLoop()
	return server, nil
}
