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
	"bytes"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// Ensures Execute returns an error if a bad frugal frame is passed.
func TestClientRegistryBadFrame(t *testing.T) {
	assert := assert.New(t)
	registry := newFRegistry()
	assert.Error(registry.Execute([]byte{0}))
}

// Ensures Execute returns an error if the frame fruagl headers are missing an
// opID.
func TestClientRegistryMissingOpID(t *testing.T) {
	assert := assert.New(t)
	registry := newFRegistry()
	assert.Error(registry.Execute(basicFrame))
}

// Ensures context Register, Execute, and Unregister work as intended with
// a valid frugal frame.
func TestClientRegistry(t *testing.T) {
	assert := assert.New(t)
	resultC := make(chan []byte, 1)
	registry := newFRegistry()
	ctx := NewFContext("")
	opid, err := getOpID(ctx)
	assert.Nil(err)
	assert.True(opid > 0)

	// Register the context for the first time
	assert.Nil(registry.Register(ctx, resultC))
	// Encode a frame with this context
	transport := &thrift.TMemoryBuffer{Buffer: new(bytes.Buffer)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}
	assert.Nil(proto.writeHeader(ctx.RequestHeaders()))
	// Pass the frame to execute
	frame := transport.Bytes()
	assert.Nil(registry.Execute(frame))
	assert.Equal(1, len(resultC))

	// Re-assign the same context
	assert.Error(registry.Register(ctx, resultC))

	// Unregister the context
	registry.Unregister(ctx)
	opid, err = getOpID(ctx)
	assert.Nil(err)
	_, ok := registry.(*fRegistryImpl).channels[opid]
	assert.False(ok)
	// But make sure execute sill returns nil when executing a frame with the
	// same opID (it will just drop the frame)
	assert.Nil(registry.Execute(frame))
	assert.Equal(1, len(resultC))

	// Now, register the same context again and ensure the opID is increased.
	assert.Nil(registry.Register(ctx, resultC))
	_, err = getOpID(ctx)
	assert.Nil(err)
}

type mockProcessor struct {
	iprot *FProtocol
	oprot *FProtocol
}

func (p *mockProcessor) Process(in, out *FProtocol) error {
	p.iprot = in
	p.oprot = out
	return nil
}

func (p *mockProcessor) AddMiddleware(middleware ServiceMiddleware) {}
