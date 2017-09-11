/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package frugal

import (
	"git.apache.org/thrift.git/lib/go/thrift"
)

// FSimpleServer is a simple FServer which starts a goroutine for each
// connection.
type FSimpleServer struct {
	quit            chan struct{}
	processor       FProcessor
	serverTransport thrift.TServerTransport
	protocolFactory *FProtocolFactory
}

// NewFSimpleServer creates a new FSimpleServer which is a simple FServer that
// starts a goroutine for each connection.
func NewFSimpleServer(
	processor FProcessor,
	serverTransport thrift.TServerTransport,
	protocolFactory *FProtocolFactory) *FSimpleServer {

	return &FSimpleServer{
		processor:       processor,
		serverTransport: serverTransport,
		protocolFactory: protocolFactory,
		quit:            make(chan struct{}, 1),
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
	framed := NewTFramedTransport(client)
	iprot := p.protocolFactory.GetProtocol(framed)
	oprot := p.protocolFactory.GetProtocol(framed)
	processor := p.processor

	logger().Debug("frugal: client connection accepted")

	for {
		err := processor.Process(iprot, oprot)
		if err, ok := err.(thrift.TTransportException); ok && err.TypeId() == TRANSPORT_EXCEPTION_END_OF_FILE {
			return nil
		} else if err != nil {
			logger().Printf("error processing request: %s", err)
			return err
		}
		if err, ok := err.(thrift.TApplicationException); ok && err.TypeId() == APPLICATION_EXCEPTION_UNKNOWN_METHOD {
			continue
		}
	}
}
