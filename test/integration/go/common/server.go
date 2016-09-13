package common

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/test/integration/go/gen/frugaltest"
)

func StartServer(
	host string,
	port int64,
	transport string,
	protocol string,
	handler frugaltest.FFrugalTest,
	serverMiddlewareCalled chan bool,
	pubSubResponseSent chan bool) {

	var protocolFactory thrift.TProtocolFactory
	switch protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		panic(fmt.Errorf("Invalid protocol specified %s", protocol))
	}

	conn := getNatsConn()
	var err error

	/*
		Subscriber for Pub/Sub tests
		Subscribe to events, publish response upon receipt
	*/
	go func() {
		factory := frugal.NewFNatsScopeTransportFactory(conn)
		provider := frugal.NewFScopeProvider(factory, frugal.NewFProtocolFactory(protocolFactory))
		subscriber := event.NewEventsSubscriber(provider)

		// TODO: Document SubscribeEventCreated "user" cannot contain spaces
		_, err = subscriber.SubscribeEventCreated(fmt.Sprintf("%d-call", port), func(ctx *frugal.FContext, e *event.Event) {
			// Send a message back to the client
			fmt.Printf("received %+v : %+v\n", ctx, e)
			publisher := event.NewEventsPublisher(provider)
			if err := publisher.Open(); err != nil {
				panic(err)
			}
			defer publisher.Close()
			ctx = frugal.NewFContext("Response")
			event := &event.Event{Message: "received call"}
			if err := publisher.PublishEventCreated(ctx, fmt.Sprintf("%d-response", port), event); err != nil {
				panic(err)
			}
			// Explicitly flushing the publish to ensure it is sent before the main thread exits
			conn.Flush()
			pubSubResponseSent <- true
		})
		if err != nil {
			panic(err)
		}
	}()

	hostPort := fmt.Sprintf("%s:%d", host, port)

	/*
		Server for RPC tests
		Server handler is defined in printing_handler.go
	*/
	processor := frugaltest.NewFFrugalTestProcessor(handler, serverLoggingMiddleware(serverMiddlewareCalled))
	var server frugal.FServer
	switch transport {
	case "stateless":
		builder := frugal.NewFNatsServerBuilder(
			conn,
			processor,
			frugal.NewFProtocolFactory(protocolFactory),
			fmt.Sprintf("%d", port))
		server = builder.Build()
		// Start http server
		// Healthcheck used in the cross language runner to check for server availability
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
		go http.ListenAndServe(hostPort, nil)
	case "http":
		http.HandleFunc("/",
			frugal.NewFrugalHandlerFunc(processor,
				frugal.NewFProtocolFactory(protocolFactory),
				frugal.NewFProtocolFactory(protocolFactory)))
		server = &httpServer{hostPort: hostPort}
	case "stateful", "stateless-stateful": // @Deprecated TODO: Remove in 2.0
		fTransportFactory := frugal.NewFMuxTransportFactory(2)
		server = frugal.NewFNatsServer(
			conn,
			fmt.Sprintf("%d", port),
			time.Minute,
			processor,
			fTransportFactory,
			frugal.NewFProtocolFactory(protocolFactory),
		)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
		go http.ListenAndServe(hostPort, nil)
	}
	fmt.Println("Starting %v server...", transport)
	if err := server.Serve(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

type httpServer struct {
	hostPort string
}

func (h *httpServer) Serve() error {
	return http.ListenAndServe(h.hostPort, http.DefaultServeMux)
}

func (h *httpServer) Stop() error {
	return nil
}

func (h *httpServer) SetHighWatermark(_ time.Duration) {
}
