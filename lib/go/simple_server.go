package frugal

import (
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// FSimpleServer is a simple FServer which starts a goroutine for each
// connection.
type FSimpleServer struct {
	quit             chan struct{}
	processorFactory FProcessorFactory
	serverTransport  thrift.TServerTransport
	transportFactory FTransportFactory
	protocolFactory  *FProtocolFactory
	waterMu          sync.RWMutex
}

// NewFSimpleServerFactory4 creates a new FSimpleServer which is a simple
// FServer that starts a goroutine for each connection.
func NewFSimpleServerFactory4(
	processorFactory FProcessorFactory,
	serverTransport thrift.TServerTransport,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) *FSimpleServer {

	return &FSimpleServer{
		processorFactory: processorFactory,
		serverTransport:  serverTransport,
		transportFactory: transportFactory,
		protocolFactory:  protocolFactory,
		quit:             make(chan struct{}, 1),
	}
}

// Listen should not be called directly.
func (p *FSimpleServer) listen() error {
	return p.serverTransport.Listen()
}

// AcceptLoop should not be called directly.
func (p *FSimpleServer) acceptLoop() error {
	for {
		client, err := p.serverTransport.Accept()
		if err != nil {
			select {
			case <-p.quit:
				return nil
			default:
			}
			return err
		}
		if client != nil {
			go func() {
				if err := p.accept(client); err != nil {
					logger().Error("frugal: error accepting client transport:", err)
				}
			}()
		}
	}
}

// Serve starts the server.
func (p *FSimpleServer) Serve() error {
	if err := p.listen(); err != nil {
		return err
	}
	p.acceptLoop()
	return nil
}

// Stop the server.
func (p *FSimpleServer) Stop() error {
	close(p.quit)
	p.serverTransport.Interrupt()
	return nil
}

func (p *FSimpleServer) accept(client thrift.TTransport) error {
	processor := p.processorFactory.GetProcessor(client)
	transport := p.transportFactory.GetTransport(client)
	protocol := p.protocolFactory.GetProtocol(transport)
	transport.SetRegistry(NewServerRegistry(processor, p.protocolFactory, protocol))

	if err := transport.Open(); err != nil {
		return err
	}

	logger().Debug("frugal: client connection accepted")
	return nil
}
