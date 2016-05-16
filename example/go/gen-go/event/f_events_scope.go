// Autogenerated by Frugal Compiler (1.3.1)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package event

import (
	"fmt"
	"log"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/go"
)

const delimiter = "."

// This docstring gets added to the generated code because it has
// the @ sign. Prefix specifies topic prefix tokens, which can be static or
// variable.
type EventsPublisher interface {
	Open() error
	Close() error
	PublishEventCreated(ctx *frugal.FContext, user string, req *Event) error
}

type eventsPublisher struct {
	transport frugal.FScopeTransport
	protocol  *frugal.FProtocol
	methods   map[string]*frugal.Method
}

func NewEventsPublisher(provider *frugal.FScopeProvider, middleware ...frugal.ServiceMiddleware) EventsPublisher {
	transport, protocol := provider.New()
	methods := make(map[string]*frugal.Method)
	publisher := &eventsPublisher{
		transport: transport,
		protocol:  protocol,
		methods:   methods,
	}
	methods["publishEventCreated"] = frugal.NewMethod(publisher, publisher.publishEventCreated, "publishEventCreated", middleware)
	return publisher
}

func (l *eventsPublisher) Open() error {
	return l.transport.Open()
}

func (l *eventsPublisher) Close() error {
	return l.transport.Close()
}

// This is a docstring.
func (l *eventsPublisher) PublishEventCreated(ctx *frugal.FContext, user string, req *Event) error {
	ret := l.methods["publishEventCreated"].Invoke([]interface{}{ctx, user, req})
	if ret[0] != nil {
		return ret[0].(error)
	}
	return nil
}

func (l *eventsPublisher) publishEventCreated(ctx *frugal.FContext, user string, req *Event) error {
	op := "EventCreated"
	prefix := fmt.Sprintf("foo.%s.", user)
	topic := fmt.Sprintf("%sEvents%s%s", prefix, delimiter, op)
	if err := l.transport.LockTopic(topic); err != nil {
		return err
	}
	defer l.transport.UnlockTopic()
	oprot := l.protocol
	if err := oprot.WriteRequestHeader(ctx); err != nil {
		return err
	}
	if err := oprot.WriteMessageBegin(op, thrift.CALL, 0); err != nil {
		return err
	}
	if err := req.Write(oprot); err != nil {
		return err
	}
	if err := oprot.WriteMessageEnd(); err != nil {
		return err
	}
	return oprot.Flush()
}

// This docstring gets added to the generated code because it has
// the @ sign. Prefix specifies topic prefix tokens, which can be static or
// variable.
type EventsSubscriber interface {
	SubscribeEventCreated(user string, handler func(*frugal.FContext, *Event)) (*frugal.FSubscription, error)
}

type eventsSubscriber struct {
	provider   *frugal.FScopeProvider
	middleware []frugal.ServiceMiddleware
}

func NewEventsSubscriber(provider *frugal.FScopeProvider, middleware ...frugal.ServiceMiddleware) EventsSubscriber {
	return &eventsSubscriber{provider: provider, middleware: middleware}
}

// This is a docstring.
func (l *eventsSubscriber) SubscribeEventCreated(user string, handler func(*frugal.FContext, *Event)) (*frugal.FSubscription, error) {
	op := "EventCreated"
	prefix := fmt.Sprintf("foo.%s.", user)
	topic := fmt.Sprintf("%sEvents%s%s", prefix, delimiter, op)
	transport, protocol := l.provider.New()
	if err := transport.Subscribe(topic); err != nil {
		return nil, err
	}

	method := frugal.NewMethod(l, handler, "SubscribeEventCreated", l.middleware)
	sub := frugal.NewFSubscription(topic, transport)
	go func() {
		for {
			ctx, received, err := l.recvEventCreated(op, protocol)
			if err != nil {
				if e, ok := err.(thrift.TTransportException); ok && e.TypeId() == thrift.END_OF_FILE {
					return
				}
				log.Printf("frugal: error receiving %s, discarding frame: %s\n", topic, err.Error())
				transport.DiscardFrame()
				continue
			}
			method.Invoke([]interface{}{ctx, received})
		}
	}()

	return sub, nil
}

func (l *eventsSubscriber) recvEventCreated(op string, iprot *frugal.FProtocol) (*frugal.FContext, *Event, error) {
	ctx, err := iprot.ReadRequestHeader()
	if err != nil {
		return nil, nil, err
	}
	name, _, _, err := iprot.ReadMessageBegin()
	if err != nil {
		return nil, nil, err
	}
	if name != op {
		iprot.Skip(thrift.STRUCT)
		iprot.ReadMessageEnd()
		x9 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+name)
		return nil, nil, x9
	}
	req := &Event{}
	if err := req.Read(iprot); err != nil {
		return nil, nil, err
	}

	iprot.ReadMessageEnd()
	return ctx, req, nil
}
