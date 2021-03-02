package frugal

import (
	"github.com/Workiva/frugal/lib/go/shim"
	"github.com/apache/thrift/lib/go/thrift"
)

var _ FClient = (*FStandardClient)(nil)

// FClient ...
type FClient interface {
	Open() error  // holdover from publisher refactor, remove in frugal v4
	Close() error // holdover from publisher refactor, remvoe in frugal v4
	Call(ctx FContext, method string, args, result shim.TStruct0_13) error
	Oneway(ctx FContext, method string, args shim.TStruct0_13) error
	Publish(ctx FContext, op, topic string, message shim.TStruct0_13) error
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
func (client *FStandardClient) Call(ctx FContext, method string, args, result shim.TStruct0_13) error {
	payload, err := client.prepareMessage(ctx, method, args, thrift.CALL)
	if err != nil {
		return err
	}
	resultTransport, err := client.transport.Request(ctx, payload)
	if err != nil {
		return err
	}
	return client.processReply(ctx, method, result, resultTransport)
}

// Oneway sends a message to a service, without waiting for a response.
func (client *FStandardClient) Oneway(ctx FContext, method string, args shim.TStruct0_13) error {
	payload, err := client.prepareMessage(ctx, method, args, thrift.ONEWAY)
	if err != nil {
		return err
	}
	return client.transport.Oneway(ctx, payload)
}

// Publish sends a message to a topic.
func (client *FStandardClient) Publish(ctx FContext, op, topic string, message shim.TStruct0_13) error {
	payload, err := client.prepareMessage(ctx, op, message, thrift.CALL)
	if err != nil {
		return err
	}
	return client.publisher.Publish(topic, payload)
}

func (client FStandardClient) prepareMessage(ctx FContext, method string, args shim.TStruct0_13, kind thrift.TMessageType) ([]byte, error) {
	buffer := NewTMemoryOutputBuffer(client.limit)
	oprot := client.protocolFactory.GetProtocol(buffer)
	if err := oprot.WriteRequestHeader(ctx); err != nil {
		return nil, err
	}
	if err := oprot.WriteMessageBegin(method, kind, 0); err != nil {
		return nil, err
	}
	if err := args.Write(oprot); err != nil {
		return nil, err
	}
	if err := oprot.WriteMessageEnd(); err != nil {
		return nil, err
	}
	if err := oprot.Flush(toCTX(ctx)); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (client FStandardClient) processReply(ctx FContext, method string, result shim.TStruct0_13, resultTransport thrift.TTransport) error {
	iprot := client.protocolFactory.GetProtocol(resultTransport)
	if err := iprot.ReadResponseHeader(ctx); err != nil {
		return err
	}
	oMethod, mTypeID, _, err := iprot.ReadMessageBegin()
	if err != nil {
		return err
	}
	if oMethod != method {
		return thrift.NewTApplicationException(APPLICATION_EXCEPTION_WRONG_METHOD_NAME, method+" failed: wrong method name")
	}
	if mTypeID == thrift.EXCEPTION {
		error0 := thrift.NewTApplicationException(APPLICATION_EXCEPTION_UNKNOWN, "Unknown Exception")
		err = error0.Read(iprot)
		if err != nil {
			return err
		}
		if err = iprot.ReadMessageEnd(); err != nil {
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
	if err = result.Read(iprot); err != nil {
		return err
	}
	return iprot.ReadMessageEnd()
}
