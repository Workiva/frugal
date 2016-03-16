package common

import (
	"crypto/tls"
	"flag"
	"fmt"
	"gen/frugaltest"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/fruga/test/integration/go/gen/frugaltest"
	"github.com/Workiva/frugal/lib/go"
)

var debugClientProtocol bool

func init() {
	flag.BoolVar(&debugClientProtocol, "debug_client_protocol", false, "turn client protocol trace on")
}

func StartClient(
	host string,
	port int64,
	domain_socket string,
	transport string,
	protocol string,
	ssl bool) (client *frugaltest.FFrugalTestClient, err error) { // Not sure about this frugaltest

	hostPort := fmt.Sprintf("%s:%d", host, port)

	var protocolFactory = thrift.TProtocolFactory
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
		protocolFactory = thrift.NewTDebugProtocolFactory(protocolFactory, "client:")
	}

	// var trans frugal.FTransport
	var trans thrift.TTransport
	if ssl {
		trans, err = thrift.NewTSSLSocket(hostPort, &tls.Config{InsecureSkipVerify: true})
	} else {
		if domain_socket != "" {
			trans, err = thrift.NewTSocket(domain_socket)
		} else {
			trans, err = thrift.NewTSocket(hostPort)
		}
	}
	if err != nil {
		return nil, err
	}

	if err = trans.Open(); err != nil {
		return nil, err
	}
	client = frugaltest.NewFrugalTestClient(trans, protocolFactory)
}
