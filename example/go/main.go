package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"

	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/frugal/lib/go"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	var (
		server   = flag.Bool("server", false, "Run server")
		protocol = flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
		addr     = flag.String("addr", nats.DefaultURL, "NATS address")
		secure   = flag.Bool("secure", false, "Use tls secure transport")
	)
	flag.Parse()

	var protocolFactory thrift.TProtocolFactory
	switch *protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary", "":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		fmt.Fprint(os.Stderr, "Invalid protocol specified", protocol, "\n")
		Usage()
		os.Exit(1)
	}

	fprotocolFactory := frugal.NewFProtocolFactory(protocolFactory)
	ftransportFactory := frugal.NewFMuxTransportFactory(5)

	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{*addr}
	natsOptions.Secure = *secure
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}

	if !*server {
		if err := runPublisher(conn, fprotocolFactory); err != nil {
			fmt.Println("error running publisher:", err)
		}
		if err := runClient(conn, ftransportFactory, fprotocolFactory); err != nil {
			fmt.Println("error running client:", err)
		}
	} else {
		if err := runSubscriber(conn, fprotocolFactory); err != nil {
			fmt.Println("error running subscriber:", err)
		}
		if err := runServer(conn, ftransportFactory, fprotocolFactory); err != nil {
			fmt.Println("error running server:", err)
		}
	}
}

// Client handler
func handleClient(client *event.FFooClient) (err error) {
	if err := client.Ping(frugal.NewFContext("")); err != nil {
		fmt.Println("Ping error:", err)
	} else {
		fmt.Println("Ping()")
	}
	if err := client.BasePing(frugal.NewFContext("")); err != nil {
		fmt.Println("BasePing error:", err)
	} else {
		fmt.Println("BasePing()")
	}
	if err := client.OneWay(frugal.NewFContext(""), 99, event.Request{99: "request"}); err != nil {
		fmt.Println("OneWay error:", err)
	} else {
		fmt.Println("OneWay()")
	}
	event := &event.Event{Message: "hello, world!"}
	ctx := frugal.NewFContext("")
	result, err := client.Blah(ctx, 100, "awesomesauce", event)
	fmt.Printf("Blah = %d\n", result)
	fmt.Println(ctx.ResponseHeader("foo"))
	fmt.Printf("%+v\n", ctx)
	return err
}

// Client runner
func runClient(conn *nats.Conn, transportFactory frugal.FTransportFactory, protocolFactory *frugal.FProtocolFactory) error {
	transport := frugal.NewNatsServiceTTransport(conn, "foo", 5*time.Second, 2)
	ftransport := transportFactory.GetTransport(transport)
	defer ftransport.Close()
	if err := ftransport.Open(); err != nil {
		return err
	}
	return handleClient(event.NewFFooClient(ftransport, protocolFactory))
}

// Sever handler
type FooHandler struct {
}

func (f *FooHandler) Ping(ctx *frugal.FContext) error {
	fmt.Printf("Ping(%s)\n", ctx)
	return nil
}

func (f *FooHandler) Blah(ctx *frugal.FContext, num int32, str string, e *event.Event) (int64, error) {
	fmt.Printf("Blah(%s, %d, %s, %v)\n", ctx, num, str, e)
	ctx.AddResponseHeader("foo", "bar")
	return 42, nil
}

func (f *FooHandler) BasePing(ctx *frugal.FContext) error {
	fmt.Printf("BasePing(%s)\n", ctx)
	return nil
}

func (f *FooHandler) OneWay(ctx *frugal.FContext, id event.ID, req event.Request) error {
	fmt.Printf("OneWay(%s, %s, %s)\n", ctx, id, req)
	return nil
}

// Server runner
func runServer(conn *nats.Conn, transportFactory frugal.FTransportFactory,
	protocolFactory *frugal.FProtocolFactory) error {
	handler := &FooHandler{}
	processor := event.NewFFooProcessor(handler)
	server := frugal.NewFNatsServerFactory(conn, "foo", 5*time.Second, 2,
		frugal.NewFProcessorFactory(processor), transportFactory, protocolFactory)
	fmt.Println("Starting the simple nats server... on ", "foo")
	return server.Serve()
}

// Subscriber runner
func runSubscriber(conn *nats.Conn, protocolFactory *frugal.FProtocolFactory) error {
	factory := frugal.NewFNatsScopeTransportFactory(conn)
	provider := frugal.NewFScopeProvider(factory, protocolFactory)
	subscriber := event.NewEventsSubscriber(provider)
	_, err := subscriber.SubscribeEventCreated("barUser", func(ctx *frugal.FContext, e *event.Event) {
		fmt.Printf("received %+v : %+v\n", ctx, e)
	})
	if err != nil {
		return err
	}
	log.Println("Subscriber started...")
	return nil
}

// Publisher runner
func runPublisher(conn *nats.Conn, protocolFactory *frugal.FProtocolFactory) error {
	factory := frugal.NewFNatsScopeTransportFactory(conn)
	provider := frugal.NewFScopeProvider(factory, protocolFactory)
	publisher := event.NewEventsPublisher(provider)
	if err := publisher.Open(); err != nil {
		return err
	}
	defer publisher.Close()
	ctx := frugal.NewFContext("a-corr-id")
	event := &event.Event{Message: "hello, world!"}
	if err := publisher.PublishEventCreated(ctx, "barUser", event); err != nil {
		return err
	}
	fmt.Println("EventCreated()")
	return nil
}
