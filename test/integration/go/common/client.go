package common

import (
	"flag"
	"fmt"
	"log"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/test/integration/go/gen/frugaltest"
)

var debugClientProtocol bool

func init() {
	flag.BoolVar(&debugClientProtocol, "debug_client_protocol", false, "turn client protocol trace on")
}

func StartClient(
	host string,
	port int64,
	transport string,
	protocol string,
	pubSub chan bool,
	sent chan bool) (client *frugaltest.FFrugalTestClient, err error) {

	hostPort := fmt.Sprintf("%s:%d", host, port)

	var protocolFactory thrift.TProtocolFactory
	switch protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		return nil, fmt.Errorf("Invalid protocol specified %s", protocol)
	}
	if debugClientProtocol {
		protocolFactory = thrift.NewTDebugProtocolFactory(protocolFactory, "client:")
	}

	var trans thrift.TTransport
	trans, err = thrift.NewTSocket(hostPort)
	if err != nil {
		return nil, err
	}
	switch transport {
	case "http":
		trans, err = thrift.NewTHttpClient(fmt.Sprintf("http://%s/service", hostPort))
		if err != nil {
			return nil, err
		}
	case "framed":
		trans = thrift.NewTFramedTransport(trans)
	case "buffered":
		trans = thrift.NewTBufferedTransport(trans, 8192)
	case "":
	default:
		return nil, fmt.Errorf("Invalid transport specified %s", transport)
	}

	fTransportFactory := frugal.NewAdapterTransportFactory()
	fTransport := fTransportFactory.GetTransport(trans)

	if err := fTransport.Open(); err != nil {
		log.Fatal(err)
	}

	/*
		Pub/Sub Test
		Publish a message, verify that a subscriber receives the message and publishes a response.
		Verifies that scopes are correctly generated.
	*/
	go func() {
		<-pubSub

		conn, err := getNatsConn()
		if err != nil {
			panic(err)
		}

		factory := frugal.NewFNatsScopeTransportFactory(conn)
		provider := frugal.NewFScopeProvider(factory, frugal.NewFProtocolFactory(protocolFactory))
		publisher := event.NewEventsPublisher(provider)
		if err := publisher.Open(); err != nil {
			panic(err)
		}
		defer publisher.Close()

		// Start Subscription, pass timeout
		resp := make(chan bool)
		subscriber := event.NewEventsSubscriber(provider)
		// TODO: Document SubscribeEventCreated "user" cannot contain spaces
		_, err = subscriber.SubscribeEventCreated(fmt.Sprintf("%d-response", port), func(ctx *frugal.FContext, e *event.Event) {
			fmt.Printf("Response received %v\n", e)
			close(resp)
		})
		ctx := frugal.NewFContext("Call")
		event := &event.Event{Message: "Sending call"}
		fmt.Println("Publishing...")
		if err := publisher.PublishEventCreated(ctx, fmt.Sprintf("%d-call", port), event); err != nil {
			panic(err)
		}

		select {
		case <-resp:
			fmt.Println("Pub/Sub response received from server")
		case <-time.After(2 * time.Second):
			log.Fatal("Pub/Sub response timed out!")
		}
		close(sent)
	}()

	fProtocolFactory := frugal.NewFProtocolFactory(protocolFactory)

	client = frugaltest.NewFFrugalTestClient(fTransport, fProtocolFactory)
	return
}
