package frugal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const defaultMaxLength = 16384000

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

func NewTFramedTransportFactory(factory thrift.TTransportFactory) thrift.TTransportFactory {
	return &tFramedTransportFactory{factory: factory, maxLength: defaultMaxLength}
}

func NewTFramedTransportFactoryMaxLength(factory thrift.TTransportFactory, maxLength uint32) thrift.TTransportFactory {
	return &tFramedTransportFactory{factory: factory, maxLength: maxLength}
}

func (p *tFramedTransportFactory) GetTransport(base thrift.TTransport) thrift.TTransport {
	return NewTFramedTransportMaxLength(p.factory.GetTransport(base), p.maxLength)
}

func NewTFramedTransport(transport thrift.TTransport) *TFramedTransport {
	return &TFramedTransport{transport: transport, reader: bufio.NewReader(transport), maxLength: defaultMaxLength}
}

func NewTFramedTransportMaxLength(transport thrift.TTransport, maxLength uint32) *TFramedTransport {
	return &TFramedTransport{transport: transport, reader: bufio.NewReader(transport), maxLength: maxLength}
}

func (p *TFramedTransport) Open() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transport.Open()
}

func (p *TFramedTransport) IsOpen() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transport.IsOpen()
}

func (p *TFramedTransport) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transport.Close()
}

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

func (p *TFramedTransport) Write(buf []byte) (int, error) {
	n, err := p.buf.Write(buf)
	return n, thrift.NewTTransportExceptionFromError(err)
}

func (p *TFramedTransport) Flush() error {
	size := p.buf.Len()
	buf := p.writeBuffer[:4]
	binary.BigEndian.PutUint32(buf, uint32(size))
	_, err := p.transport.Write(buf)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	if size > 0 {
		if _, err := p.buf.WriteTo(p.transport); err != nil {
			if err == ErrTooLarge {
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
	size := binary.BigEndian.Uint32(buf)
	if size < 0 || size > p.maxLength {
		return 0, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: incorrect frame size (%d)", size))
	}
	return size, nil
}

func (p *TFramedTransport) RemainingBytes() (num_bytes uint64) {
	return uint64(p.frameSize)
}
