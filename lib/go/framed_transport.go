package frugal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const DEFAULT_MAX_LENGTH = 16384000

type TFramedTransport struct {
	transport   thrift.TTransport
	buf         bytes.Buffer
	reader      *bufio.Reader
	frameSize   uint32 //Current remaining size of the frame. if ==0 read next frame header
	maxLength   uint32
	readBuffer  [4]byte
	writeBuffer [4]byte
}

type tFramedTransportFactory struct {
	factory   thrift.TTransportFactory
	maxLength uint32
}

func NewTFramedTransportFactory(factory thrift.TTransportFactory) thrift.TTransportFactory {
	return &tFramedTransportFactory{factory: factory, maxLength: DEFAULT_MAX_LENGTH}
}

func NewTFramedTransportFactoryMaxLength(factory thrift.TTransportFactory, maxLength uint32) thrift.TTransportFactory {
	return &tFramedTransportFactory{factory: factory, maxLength: maxLength}
}

func (p *tFramedTransportFactory) GetTransport(base thrift.TTransport) thrift.TTransport {
	return NewTFramedTransportMaxLength(p.factory.GetTransport(base), p.maxLength)
}

func NewTFramedTransport(transport thrift.TTransport) *TFramedTransport {
	return &TFramedTransport{transport: transport, reader: bufio.NewReader(transport), maxLength: DEFAULT_MAX_LENGTH}
}

func NewTFramedTransportMaxLength(transport thrift.TTransport, maxLength uint32) *TFramedTransport {
	return &TFramedTransport{transport: transport, reader: bufio.NewReader(transport), maxLength: maxLength}
}

func (p *TFramedTransport) Open() error {
	return p.transport.Open()
}

func (p *TFramedTransport) IsOpen() bool {
	return p.transport.IsOpen()
}

func (p *TFramedTransport) Close() error {
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
			err = thrift.NewTTransportExceptionFromError(fmt.Errorf("Not enough frame size %d to read %d bytes", frameSize, len(buf)))
			return
		}
	}
	got, err := p.reader.Read(buf)
	p.frameSize = p.frameSize - uint32(got)
	//sanity check
	if p.frameSize < 0 {
		return 0, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, "Negative frame size")
	}
	return got, thrift.NewTTransportExceptionFromError(err)
}

func (p *TFramedTransport) ReadByte() (c byte, err error) {
	if p.frameSize == 0 {
		p.frameSize, err = p.readFrameHeader()
		if err != nil {
			return
		}
	}
	if p.frameSize < 1 {
		return 0, thrift.NewTTransportExceptionFromError(fmt.Errorf("Not enough frame size %d to read %d bytes", p.frameSize, 1))
	}
	c, err = p.reader.ReadByte()
	if err == nil {
		p.frameSize--
	}
	return
}

func (p *TFramedTransport) Write(buf []byte) (int, error) {
	n, err := p.buf.Write(buf)
	return n, thrift.NewTTransportExceptionFromError(err)
}

func (p *TFramedTransport) WriteByte(c byte) error {
	return p.buf.WriteByte(c)
}

func (p *TFramedTransport) WriteString(s string) (n int, err error) {
	return p.buf.WriteString(s)
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
		return 0, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, fmt.Sprintf("Incorrect frame size (%d)", size))
	}
	return size, nil
}

func (p *TFramedTransport) RemainingBytes() (num_bytes uint64) {
	return uint64(p.frameSize)
}
