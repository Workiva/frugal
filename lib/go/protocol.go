package frugal

import (
	"encoding/binary"
	"fmt"
	"io"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const protocolV0 = 0x00

// FProtocolFactory is a factory for FProtocol.
type FProtocolFactory struct {
	protoFactory thrift.TProtocolFactory
}

func NewFProtocolFactory(protoFactory thrift.TProtocolFactory) *FProtocolFactory {
	return &FProtocolFactory{protoFactory}
}

func (f *FProtocolFactory) GetProtocol(tr thrift.TTransport) *FProtocol {
	return &FProtocol{f.protoFactory.GetProtocol(tr)}
}

// FProtocol is an extension of thrift TProtocol with the addition of headers
type FProtocol struct {
	thrift.TProtocol
}

// WriteRequestHeader writes the request headers set on the given Context
// into the protocol
func (f *FProtocol) WriteRequestHeader(ctx *FContext) error {
	return f.writeHeader(ctx.RequestHeaders())
}

// ReadRequestHeader reads the request headers on the protocol into a
// returned Context
func (f *FProtocol) ReadRequestHeader() (*FContext, error) {
	headers, err := readHeader(f.Transport())
	if err != nil {
		return nil, err
	}

	ctx := &FContext{
		requestHeaders:  make(map[string]string),
		responseHeaders: make(map[string]string),
	}

	for name, value := range headers {
		ctx.addRequestHeader(name, value)
	}

	// Put op id in response headers
	opid, ok := headers[opID]
	if !ok {
		return nil, NewFProtocolExceptionWithType(thrift.INVALID_DATA, "frugal: request missing op id")
	}
	ctx.setResponseOpID(opid)

	return ctx, nil
}

// WriteResponseHeader writes the response headers set on the given Context
// into the protocol
func (f *FProtocol) WriteResponseHeader(ctx *FContext) error {
	return f.writeHeader(ctx.ResponseHeaders())
}

// ReadResponseHeader reads the response headers on the protocol into a
// provided Context
func (f *FProtocol) ReadResponseHeader(ctx *FContext) error {
	headers, err := readHeader(f.Transport())
	if err != nil {
		return err
	}

	for name, value := range headers {
		ctx.addResponseHeader(name, value)
	}

	return nil
}

func (f *FProtocol) writeHeader(headers map[string]string) error {
	size := int32(0)
	for name, value := range headers {
		size += int32(8 + len(name) + len(value))
	}

	// Header buff = [version (1 byte), size (4 bytes), headers (size bytes)]
	// Headers = [size (4 bytes) name (size bytes) size (4 bytes) value (size bytes)*]
	buff := make([]byte, size+5)

	// Write version
	buff[0] = protocolV0

	// Write size
	binary.BigEndian.PutUint32(buff[1:5], uint32(size))

	// Write headers
	i := 5
	for name, value := range headers {
		binary.BigEndian.PutUint32(buff[i:i+4], uint32(len(name)))
		i += 4
		for k := 0; k < len(name); k++ {
			buff[i] = name[k]
			i++
		}
		binary.BigEndian.PutUint32(buff[i:i+4], uint32(len(value)))
		i += 4
		for k := 0; k < len(value); k++ {
			buff[i] = value[k]
			i++
		}
	}

	if n, err := f.Transport().Write(buff); err != nil {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, fmt.Sprintf("frugal: error writing protocol headers: %s", err.Error()))
	} else if n != len(buff) {
		return thrift.NewTTransportException(thrift.UNKNOWN_PROTOCOL_EXCEPTION, "frugal: failed to write complete protocol headers")
	}

	return nil
}

func readHeader(reader io.Reader) (map[string]string, error) {
	buff := make([]byte, 5)
	if _, err := io.ReadFull(reader, buff); err != nil {
		if e, ok := err.(thrift.TTransportException); ok && e.TypeId() == thrift.END_OF_FILE {
			return nil, err
		}
		return nil, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, fmt.Sprintf("frugal: error reading protocol headers: %s", err.Error()))
	}

	// Support more versions when available.
	if buff[0] != protocolV0 {
		return nil, NewFProtocolExceptionWithType(thrift.BAD_VERSION, fmt.Sprintf("frugal: unsupported protocol version %d", buff[0]))
	}

	size := int32(binary.BigEndian.Uint32(buff[1:]))
	buff = make([]byte, size)
	if _, err := io.ReadFull(reader, buff); err != nil {
		if e, ok := err.(thrift.TTransportException); ok && e.TypeId() == thrift.END_OF_FILE {
			return nil, err
		}
		return nil, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, fmt.Sprintf("frugal: error reading protocol headers: %s", err.Error()))
	}

	return readPairs(buff, 0, size)
}

func getHeadersFromFrame(frame []byte) (map[string]string, error) {
	if len(frame) < 5 {
		return nil, NewFProtocolExceptionWithType(thrift.INVALID_DATA, fmt.Sprintf("frugal: invalid frame size %d", len(frame)))
	}

	// Support more versions when available.
	if frame[0] != protocolV0 {
		return nil, NewFProtocolExceptionWithType(thrift.BAD_VERSION, fmt.Sprint("frugal: unsupported protocol version %d", frame[0]))
	}

	size := int32(binary.BigEndian.Uint32(frame[1:5]))
	return readPairs(frame, 5, size+5)
}

func readPairs(buff []byte, start, end int32) (map[string]string, error) {
	headers := make(map[string]string)
	i := start
	for i < end {
		// Read header name.
		nameSize := int32(binary.BigEndian.Uint32(buff[i : i+4]))
		i += 4
		if i > end || i+nameSize > end {
			return nil, NewFProtocolExceptionWithType(thrift.INVALID_DATA, "frugal: invalid protocol header name")
		}
		name := string(buff[i : i+nameSize])
		i += nameSize

		// Read header value.
		valueSize := int32(binary.BigEndian.Uint32(buff[i : i+4]))
		i += 4
		if i > end || i+valueSize > end {
			return nil, NewFProtocolExceptionWithType(thrift.INVALID_DATA, "frugal: invalid protocol header value")
		}
		value := string(buff[i : i+valueSize])
		i += valueSize

		headers[name] = value
	}

	return headers, nil
}
