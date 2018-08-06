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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const defaultMaxLength = 16384000

// TFramedTransport is an implementation of thrift.TTransport which frames
// messages with their size.
type TFramedTransport struct {
	transport   thrift.TTransport
	buf         bytes.Buffer
	reader      *bufio.Reader
	frameSize   uint32 // Current remaining size of the frame. If 0, read next frame header.
	maxLength   uint32
	readBuffer  [4]byte
	writeBuffer [4]byte
	mu          sync.Mutex
}

type tFramedTransportFactory struct {
	factory   thrift.TTransportFactory
	maxLength uint32
}

// NewTFramedTransportFactory creates a new TTransportFactory that produces
// TFramedTransports.
func NewTFramedTransportFactory(factory thrift.TTransportFactory) thrift.TTransportFactory {
	return &tFramedTransportFactory{factory: factory, maxLength: defaultMaxLength}
}

// NewTFramedTransportFactoryMaxLength creates a new TTransportFactory that
// produces TFramedTransports with the given max length.
func NewTFramedTransportFactoryMaxLength(factory thrift.TTransportFactory, maxLength uint32) thrift.TTransportFactory {
	return &tFramedTransportFactory{factory: factory, maxLength: maxLength}
}

// GetTransport creates a new TFramedTransport wrapping the given TTransport.
func (p *tFramedTransportFactory) GetTransport(base thrift.TTransport) thrift.TTransport {
	return NewTFramedTransportMaxLength(p.factory.GetTransport(base), p.maxLength)
}

// NewTFramedTransport creates a new TFramedTransport wrapping the given
// TTransport.
func NewTFramedTransport(transport thrift.TTransport) *TFramedTransport {
	return &TFramedTransport{transport: transport, reader: bufio.NewReader(transport), maxLength: defaultMaxLength}
}

// NewTFramedTransportMaxLength creates a new TFramedTransport wrapping the
// given TTransport using the given max length.
func NewTFramedTransportMaxLength(transport thrift.TTransport, maxLength uint32) *TFramedTransport {
	return &TFramedTransport{transport: transport, reader: bufio.NewReader(transport), maxLength: maxLength}
}

// Open the transport.
func (p *TFramedTransport) Open() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transport.Open()
}

// IsOpen checks if the transport is open.
func (p *TFramedTransport) IsOpen() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transport.IsOpen()
}

// Close the transport.
func (p *TFramedTransport) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transport.Close()
}

// Read from the transport.
func (p *TFramedTransport) Read(buf []byte) (l int, err error) {
	if p.frameSize == 0 {
		p.frameSize, err = p.readFrameHeader()
		if err != nil {
			return
		}
	}
	if p.frameSize < uint32(len(buf)) {
		frameSize := p.frameSize
		tmp := make([]byte, p.frameSize)
		l, err = p.Read(tmp)
		copy(buf, tmp)
		if err == nil {
			err = thrift.NewTTransportExceptionFromError(
				fmt.Errorf("frugal: not enough frame (size %d) to read %d bytes", frameSize, len(buf)))
			return
		}
	}
	got, err := p.reader.Read(buf)
	p.frameSize = p.frameSize - uint32(got)
	return got, thrift.NewTTransportExceptionFromError(err)
}

// Write to the transport.
func (p *TFramedTransport) Write(buf []byte) (int, error) {
	n, err := p.buf.Write(buf)
	return n, thrift.NewTTransportExceptionFromError(err)
}

// Flush the transport.
func (p *TFramedTransport) Flush() error {
	size := p.buf.Len()
	buf := p.writeBuffer[:4]
	bigEndianPutUint32(buf, uint32(size))
	_, err := p.transport.Write(buf)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	if size > 0 {
		if _, err := p.buf.WriteTo(p.transport); err != nil {
			if IsErrTooLarge(err) {
				p.buf.Reset()
			}
			return thrift.NewTTransportExceptionFromError(err)
		}
	}
	err = p.transport.Flush()
	return thrift.NewTTransportExceptionFromError(err)
}

func (p *TFramedTransport) readFrameHeader() (uint32, error) {
	buf := p.readBuffer[:4]
	if _, err := io.ReadFull(p.reader, buf); err != nil {
		return 0, err
	}
	size := bigEndianUint32(buf)
	if size < 0 || size > p.maxLength {
		return 0, thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
			fmt.Sprintf("frugal: incorrect frame size (%d)", size))
	}
	return size, nil
}

// RemainingBytes returns the current frame size.
func (p *TFramedTransport) RemainingBytes() uint64 {
	return uint64(p.frameSize)
}
