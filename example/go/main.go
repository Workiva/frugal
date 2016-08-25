package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
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
		port     = flag.String("port", "8090", "Port for http transport")
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
		if err := runClient(conn, fprotocolFactory, *port); err != nil {
			fmt.Println("error running client:", err)
		}
	} else {
		if err := runSubscriber(conn, fprotocolFactory); err != nil {
			fmt.Println("error running subscriber:", err)
		}
		if err := runServer(conn, fprotocolFactory, *port); err != nil {
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
func runClient(conn *nats.Conn, protocolFactory *frugal.FProtocolFactory, port string) error {
	natsT := frugal.NewFNatsTransport(conn, "foo", "bar")
	defer natsT.Close()
	if err := natsT.Open(); err != nil {
		return err
	}

	if err := handleClient(event.NewFFooClient(natsT, protocolFactory)); err != nil {
		return err
	}

	httpT := frugal.NewHttpFTransportBuilder(&http.Client{}, fmt.Sprintf("http://localhost:%s/frugal", port)).Build()
	defer httpT.Close()
	if err := httpT.Open(); err != nil {
		return err
	}

	return handleClient(event.NewFFooClient(httpT, protocolFactory))
}

// Server handler
type FooHandler struct {
}

func (f *FooHandler) Ping(ctx *frugal.FContext) error {
	fmt.Printf("Ping(%+v)\n", ctx)
	return nil
}

func (f *FooHandler) Blah(ctx *frugal.FContext, num int32, str string, e *event.Event) (int64, error) {
	fmt.Printf("Blah(%+v, %d, %s, %v)\n", ctx, num, str, e)
	ctx.AddResponseHeader("foo", "bar")
	return 42, nil
}

func (f *FooHandler) BasePing(ctx *frugal.FContext) error {
	fmt.Printf("BasePing(%+v)\n", ctx)
	return nil
}

func (f *FooHandler) OneWay(ctx *frugal.FContext, id event.ID, req event.Request) error {
	fmt.Printf("OneWay(%+v, %s, %s)\n", ctx, id, req)
	return nil
}

// Server runner
func runServer(conn *nats.Conn, protocolFactory *frugal.FProtocolFactory, port string) error {
	handler := &FooHandler{}
	processor := event.NewFFooProcessor(handler)

	http.HandleFunc("/frugal", frugal.NewFrugalHandlerFunc(processor, protocolFactory, protocolFactory))
	go func() {
		fmt.Printf("Starting the http server... on :%s/frugal\n", port)
		http.ListenAndServe(fmt.Sprintf(":%s", port), http.DefaultServeMux)
	}()

	server := frugal.NewFNatsServerBuilder(conn, processor, protocolFactory, "foo").Build()
	fmt.Println("Starting the nats server... on ", "foo")
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

func newLoggingMiddleware() frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			fmt.Printf("==== CALLING %s.%s ====\n", service.Type(), method.Name)
			ret := next(service, method, args)
			fmt.Printf("==== CALLED  %s.%s ====\n", service.Type(), method.Name)
			return ret
		}
	}
}

func newRetryMiddleware() frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			var ret frugal.Results
			for i := 0; i < 5; i++ {
				ret = next(service, method, args)
				if ret.Error() != nil {
					fmt.Printf("%s.%s failed (%s), retrying...\n", service.Type(), method.Name, ret.Error())
					time.Sleep(500 * time.Millisecond)
					continue
				}
				return ret
			}
			return ret
		}
	}
}
