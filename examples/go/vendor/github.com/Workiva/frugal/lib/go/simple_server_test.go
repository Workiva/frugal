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
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const simpleServerAddr = "localhost:5535"

// Ensures FSimpleServer accepts connections.
func TestSimpleServer(t *testing.T) {
	mockFProcessor := new(mockFProcessor)
	protoFactory := thrift.NewTJSONProtocolFactory()
	fTransportFactory := NewAdapterTransportFactory()
	serverTr, err := thrift.NewTServerSocket(simpleServerAddr)
	if err != nil {
		t.Fatal(err)
	}
	server := NewFSimpleServer(
		mockFProcessor,
		serverTr,
		NewFProtocolFactory(protoFactory),
	)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	mockFProcessor.On("Process", mock.AnythingOfType("*frugal.FProtocol"),
		mock.AnythingOfType("*frugal.FProtocol")).Return(nil)

	transport, err := thrift.NewTSocket(simpleServerAddr)
	if err != nil {
		t.Fatal(err)
	}
	fTransport := fTransportFactory.GetTransport(transport)
	defer fTransport.Close()
	if err := fTransport.Open(); err != nil {
		t.Fatal(err)
	}

	ctx := NewFContext("")
	ctx.SetTimeout(5 * time.Millisecond)
	_, err = fTransport.Request(ctx, make([]byte, 10))
	assert.Equal(t, TRANSPORT_EXCEPTION_TIMED_OUT, err.(thrift.TTransportException).TypeId())

	assert.Nil(t, server.Stop())

	mockFProcessor.AssertExpectations(t)
	mockFProcessor.AssertExpectations(t)
}
