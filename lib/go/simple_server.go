package frugal

import (
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
)

// FSimpleServer is a simple, single-threaded FServer.
type FSimpleServer struct {
	quit             chan struct{}
	processorFactory FProcessorFactory
	serverTransport  thrift.TServerTransport
	transportFactory FTransportFactory
	protocolFactory  *FProtocolFactory
	highWatermark    time.Duration
	waterMu          sync.RWMutex
}

func NewFSimpleServerFactory5(
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
		highWatermark:    defaultWatermark,
	}
}

func (p *FSimpleServer) Listen() error {
	return p.serverTransport.Listen()
}

func (p *FSimpleServer) AcceptLoop() error {
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
					log.Error("frugal: error accepting client transport:", err)
				}
			}()
		}
	}
}

// Serve starts the server.
func (p *FSimpleServer) Serve() error {
	err := p.Listen()
	if err != nil {
		return err
	}
	p.AcceptLoop()
	return nil
}

// Stop the server.
func (p *FSimpleServer) Stop() error {
	close(p.quit)
	p.serverTransport.Interrupt()
	return nil
}

// SetHighWatermark sets the maximum amount of time a frame is allowed to await
// processing before triggering server overload logic. For now, this just
// consists of logging a warning. If not set, default is 5 seconds.
func (p *FSimpleServer) SetHighWatermark(watermark time.Duration) {
	p.waterMu.Lock()
	p.highWatermark = watermark
	p.waterMu.Unlock()
}

func (p *FSimpleServer) accept(client thrift.TTransport) error {
	processor := p.processorFactory.GetProcessor(client)
	transport := p.transportFactory.GetTransport(client)
	protocol := p.protocolFactory.GetProtocol(transport)
	transport.SetRegistry(NewServerRegistry(processor, p.protocolFactory, protocol))
	p.waterMu.RLock()
	transport.SetHighWatermark(p.highWatermark)
	p.waterMu.RUnlock()
	if err := transport.Open(); err != nil {
		return err
	}

	log.Debug("frugal: client connection accepted")
	return nil
}
