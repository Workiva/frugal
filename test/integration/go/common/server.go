package common

import (
	"crypto/tls"
	"flag"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/test/integration/go/gen/frugaltest"
)

var (
	debugServerProtocol bool
	certPath            string
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
	ssl bool,
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

	// Not sure if this section will work, leaving it in for now.
	if debugServerProtocol {
		protocolFactory = thrift.NewTDebugProtocolFactory(protocolFactory, "server:")
	}

	var serverTransport thrift.TServerTransport
	if ssl {
		cfg := new(tls.Config)
		if cert, err := tls.LoadX509KeyPair(certPath+"/server.crt", certPath+"/server.key"); err != nil {
			return nil, err
		} else {
			cfg.Certificates = append(cfg.Certificates, cert)
		}
	} else {
		if domain_socket != "" {
			serverTransport, err = thrift.NewTServerSocket(domain_socket)
		} else {
			serverTransport, err = thrift.NewTServerSocket(hostPort)
		}
	}
	if err != nil {
		return nil, err
	}

	fTransportFactory := frugal.NewFMuxTransportFactory(2)
	// processor := frugaltest.NewFProcessor
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
