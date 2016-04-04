package frugal

import (
	"bufio"

	"git.apache.org/thrift.git/lib/go/thrift"
)

type TBufferedTransportFactory struct {
	size int
}

type TBufferedTransport struct {
	bufio.ReadWriter
	tp thrift.TTransport
}

func (p *TBufferedTransportFactory) GetTransport(trans thrift.TTransport) thrift.TTransport {
	return NewTBufferedTransport(trans, p.size)
}

func NewTBufferedTransportFactory(bufferSize int) *TBufferedTransportFactory {
	return &TBufferedTransportFactory{size: bufferSize}
}

func NewTBufferedTransport(trans thrift.TTransport, bufferSize int) *TBufferedTransport {
	return &TBufferedTransport{
		ReadWriter: bufio.ReadWriter{
			Reader: bufio.NewReaderSize(trans, bufferSize),
			Writer: bufio.NewWriterSize(trans, bufferSize),
		},
		tp: trans,
	}
}

func (p *TBufferedTransport) IsOpen() bool {
	return p.tp.IsOpen()
}

func (p *TBufferedTransport) Open() (err error) {
	return p.tp.Open()
}

func (p *TBufferedTransport) Close() (err error) {
	return p.tp.Close()
}

func (p *TBufferedTransport) Read(pb []byte) (nn int, err error) {
	nn, err = p.ReadWriter.Read(pb)
	if err != nil {
		p.ReadWriter.Reader.Reset(p.tp)
	}
	return
}

func (p *TBufferedTransport) Write(pb []byte) (nn int, err error) {
	nn, err = p.ReadWriter.Write(pb)
	if err != nil {
		p.ReadWriter.Writer.Reset(p.tp)
	}
	return
}

func (p *TBufferedTransport) Flush() error {
	if err := p.ReadWriter.Flush(); err != nil {
		p.ReadWriter.Writer.Reset(p.tp)
		return err
	}
	return p.tp.Flush()
}

func (p *TBufferedTransport) RemainingBytes() (num_bytes uint64) {
	return p.tp.RemainingBytes()
}
