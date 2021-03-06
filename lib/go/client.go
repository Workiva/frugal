package frugal

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"
)

var _ FClient = (*FStandardClient)(nil)

// FClient ...
type FClient interface {
	Open() error  // holdover from publisher refactor, remove in frugal v4
	Close() error // holdover from publisher refactor, remvoe in frugal v4
	Call(ctx FContext, method string, args, result thrift.TStruct) error
	Oneway(ctx FContext, method string, args thrift.TStruct) error
	Publish(ctx FContext, op, topic string, message thrift.TStruct) error
}

// FStandardClient implements FClient, and uses the standard message format for Frugal.
type FStandardClient struct {
	transport       FTransport
	publisher       FPublisherTransport
	protocolFactory *FProtocolFactory
	limit           uint
}

// NewFStandardClient implements FClient, and uses the standard message format for Frugal.
func NewFStandardClient(provider *FServiceProvider) *FStandardClient {
	client := &FStandardClient{
		transport:       provider.GetTransport(),
		protocolFactory: provider.GetProtocolFactory(),
	}
	client.limit = client.transport.GetRequestSizeLimit()
	return client
}

// NewFScopeClient ...
func NewFScopeClient(provider *FScopeProvider) *FStandardClient {
	transport, protocolFactory := provider.NewPublisher()
	client := &FStandardClient{
		publisher:       transport,
		protocolFactory: protocolFactory,
	}
	client.limit = client.publisher.GetPublishSizeLimit()
	return client
}

// Open ...
func (client *FStandardClient) Open() error {
	return client.publisher.Open()
}

// Close ...
func (client *FStandardClient) Close() error {
	return client.publisher.Close()
}

// Call invokes a service and waits for a response.
func (client *FStandardClient) Call(fctx FContext, method string, args, result thrift.TStruct) error {
	ctx, cancelFn := ToContext(fctx)
	defer cancelFn()
	payload, err := client.prepareMessage(ctx, fctx, method, args, thrift.CALL)
	if err != nil {
		return err
	}
	resultTransport, err := client.transport.Request(fctx, payload)
	if err != nil {
		return err
	}
	return client.processReply(ctx, fctx, method, result, resultTransport)
}

// Oneway sends a message to a service, without waiting for a response.
func (client *FStandardClient) Oneway(fctx FContext, method string, args thrift.TStruct) error {
	ctx, cancelFn := ToContext(fctx)
	defer cancelFn()
	payload, err := client.prepareMessage(ctx, fctx, method, args, thrift.ONEWAY)
	if err != nil {
		return err
	}
	return client.transport.Oneway(fctx, payload)
}

// Publish sends a message to a topic.
func (client *FStandardClient) Publish(fctx FContext, op, topic string, message thrift.TStruct) error {
	ctx, cancelFn := ToContext(fctx)
	defer cancelFn()
	payload, err := client.prepareMessage(ctx, fctx, op, message, thrift.CALL)
	if err != nil {
		return err
	}
	return client.publisher.Publish(topic, payload)
}

func (client FStandardClient) prepareMessage(ctx context.Context, fctx FContext, method string, args thrift.TStruct, kind thrift.TMessageType) ([]byte, error) {
	buffer := NewTMemoryOutputBuffer(client.limit)
	oprot := client.protocolFactory.GetProtocol(buffer)
	if err := oprot.WriteRequestHeader(fctx); err != nil {
		return nil, err
	}
	if err := oprot.WriteMessageBegin(ctx, method, kind, 0); err != nil {
		return nil, err
	}
	if err := args.Write(ctx, oprot); err != nil {
		return nil, err
	}
	if err := oprot.WriteMessageEnd(ctx); err != nil {
		return nil, err
	}
	if err := oprot.Flush(ctx); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (client FStandardClient) processReply(ctx context.Context, fctx FContext, method string, result thrift.TStruct, resultTransport thrift.TTransport) error {
	iprot := client.protocolFactory.GetProtocol(resultTransport)
	if err := iprot.ReadResponseHeader(fctx); err != nil {
		return err
	}
	oMethod, mTypeID, _, err := iprot.ReadMessageBegin(ctx)
	if err != nil {
		return err
	}
	if oMethod != method {
		return thrift.NewTApplicationException(APPLICATION_EXCEPTION_WRONG_METHOD_NAME, method+" failed: wrong method name")
	}
	if mTypeID == thrift.EXCEPTION {
		error0 := thrift.NewTApplicationException(APPLICATION_EXCEPTION_UNKNOWN, "Unknown Exception")
		err = error0.Read(ctx, iprot)
		if err != nil {
			return err
		}
		if err = iprot.ReadMessageEnd(ctx); err != nil {
			return err
		}
		if error0.TypeId() == APPLICATION_EXCEPTION_RESPONSE_TOO_LARGE {
			return thrift.NewTTransportException(TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE, error0.Error())
		}
		return error0
	}
	if mTypeID != thrift.REPLY {
		return thrift.NewTApplicationException(APPLICATION_EXCEPTION_INVALID_MESSAGE_TYPE, method+" failed: invalid message type")
	}
	if err = result.Read(ctx, iprot); err != nil {
		return err
	}
	return iprot.ReadMessageEnd(ctx)
}
