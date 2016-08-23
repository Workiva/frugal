// Autogenerated by Frugal Compiler (1.14.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package music

import (
	"fmt"
	"log"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/go"
)

const delimiter = "."

type AlbumWinnersPublisher interface {
	Open() error
	Close() error
	PublishWinner(ctx *frugal.FContext, req *Album) error
}

type albumWinnersPublisher struct {
	transport frugal.FScopeTransport
	protocol  *frugal.FProtocol
	methods   map[string]*frugal.Method
}

func NewAlbumWinnersPublisher(provider *frugal.FScopeProvider, middleware ...frugal.ServiceMiddleware) AlbumWinnersPublisher {
	transport, protocol := provider.New()
	methods := make(map[string]*frugal.Method)
	publisher := &albumWinnersPublisher{
		transport: transport,
		protocol:  protocol,
		methods:   methods,
	}
	methods["publishWinner"] = frugal.NewMethod(publisher, publisher.publishWinner, "publishWinner", middleware)
	return publisher
}

func (l *albumWinnersPublisher) Open() error {
	return l.transport.Open()
}

func (l *albumWinnersPublisher) Close() error {
	return l.transport.Close()
}

func (l *albumWinnersPublisher) PublishWinner(ctx *frugal.FContext, req *Album) error {
	ret := l.methods["publishWinner"].Invoke([]interface{}{ctx, req})
	if ret[0] != nil {
		return ret[0].(error)
	}
	return nil
}

func (l *albumWinnersPublisher) publishWinner(ctx *frugal.FContext, req *Album) error {
	op := "Winner"
	prefix := "store."
	topic := fmt.Sprintf("%sAlbumWinners%s%s", prefix, delimiter, op)
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

type AlbumWinnersSubscriber interface {
	SubscribeWinner(handler func(*frugal.FContext, *Album)) (*frugal.FSubscription, error)
}

type albumWinnersSubscriber struct {
	provider   *frugal.FScopeProvider
	middleware []frugal.ServiceMiddleware
}

func NewAlbumWinnersSubscriber(provider *frugal.FScopeProvider, middleware ...frugal.ServiceMiddleware) AlbumWinnersSubscriber {
	return &albumWinnersSubscriber{provider: provider, middleware: middleware}
}

func (l *albumWinnersSubscriber) SubscribeWinner(handler func(*frugal.FContext, *Album)) (*frugal.FSubscription, error) {
	op := "Winner"
	prefix := "store."
	topic := fmt.Sprintf("%sAlbumWinners%s%s", prefix, delimiter, op)
	transport, protocol := l.provider.New()
	if err := transport.Subscribe(topic); err != nil {
		return nil, err
	}

	method := frugal.NewMethod(l, handler, "SubscribeWinner", l.middleware)
	sub := frugal.NewFSubscription(topic, transport)
	go func() {
		for {
			ctx, received, err := l.recvWinner(op, protocol)
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

func (l *albumWinnersSubscriber) recvWinner(op string, iprot *frugal.FProtocol) (*frugal.FContext, *Album, error) {
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
	req := &Album{}
	if err := req.Read(iprot); err != nil {
		return nil, nil, err
	}

	iprot.ReadMessageEnd()
	return ctx, req, nil
}
