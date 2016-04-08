package frugal

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

var (
	basicFrame   = []byte{0, 0, 0, 0, 14, 0, 0, 0, 3, 102, 111, 111, 0, 0, 0, 3, 98, 97, 114}
	basicHeaders = map[string]string{"foo": "bar"}

	frugalFrame = []byte{0, 0, 0, 0, 65, 0, 0, 0, 5, 104, 101, 108, 108, 111, 0, 0, 0, 5,
		119, 111, 114, 108, 100, 0, 0, 0, 5, 95, 111, 112, 105, 100, 0, 0, 0, 1, 48, 0, 0, 0,
		4, 95, 99, 105, 100, 0, 0, 0, 21, 105, 89, 65, 71, 67, 74, 72, 66, 87, 67, 75, 76, 74,
		66, 115, 106, 107, 100, 111, 104, 98}
	frugalHeaders = map[string]string{opID: "0", cid: "iYAGCJHBWCKLJBsjkdohb", "hello": "world"}

	tProtocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
)

// Ensures ReadRequestHeader returns a error when the headers do not contain
// an opID.
func TestReadRequestHeaderMissingOpID(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(basicFrame)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}

	expectedErr := NewFProtocolExceptionWithType(thrift.INVALID_DATA, "frugal: request missing op id")
	_, err := proto.ReadRequestHeader()
	assert.Equal(expectedErr, err)
}

// Ensures ReadRequestHeader correctly reads frugal request headers from the
// protocol.
func TestReadRequestHeader(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(frugalFrame)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}

	ctx, err := proto.ReadRequestHeader()
	assert.Nil(err)
	assert.Equal(frugalHeaders[cid], ctx.CorrelationID())
	assert.Equal(uint64(0), ctx.opID())
	val, ok := ctx.RequestHeader("hello")
	assert.True(ok)
	assert.Equal(frugalHeaders["hello"], val)
}

// Ensures ReadRequestHeader correctly reads frugal response headers from the
// protocol.
func TestReadResponseHeader(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(basicFrame)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}

	ctx := NewFContext("")
	proto.ReadResponseHeader(ctx)
	val, ok := ctx.ResponseHeader("foo")
	assert.True(ok)
	assert.Equal(val, "bar")
}

// Ensures writeHeader bubbles up transport Write error.
func TestWriteHeaderErroredWrite(t *testing.T) {
	assert := assert.New(t)
	mft := &mockFTransport{}
	writeErr := errors.New("write falied")
	mft.On("Write", basicFrame).Return(0, writeErr)
	proto := &FProtocol{tProtocolFactory.GetProtocol(mft)}
	expectedErr := thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, fmt.Sprintf("frugal: error writing protocol headers: %s", writeErr))
	assert.Equal(expectedErr, proto.writeHeader(basicHeaders))
}

// Ensures writeHeader returns an error if transport Write fails to write all
// the header bytes.
func TestWriteHeaderBadWrite(t *testing.T) {
	assert := assert.New(t)
	mft := &mockFTransport{}
	mft.On("Write", basicFrame).Return(0, nil)
	proto := &FProtocol{tProtocolFactory.GetProtocol(mft)}
	expectedErr := thrift.NewTTransportException(thrift.UNKNOWN_PROTOCOL_EXCEPTION, "frugal: failed to write complete protocol headers")
	assert.Equal(expectedErr, proto.writeHeader(basicHeaders))
}

// Ensures writeHeader properly encodes header bytes.
func TestWriteHeader(t *testing.T) {
	assert := assert.New(t)
	mft := &mockFTransport{}
	mft.On("Write", basicFrame).Return(len(basicFrame), nil)
	proto := &FProtocol{tProtocolFactory.GetProtocol(mft)}
	assert.Nil(proto.writeHeader(basicHeaders))
}

// Ensures readHeader returns an error if there are not enough frame bytes to
// read from the transport.
func TestReadHeaderTransportError(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer([]byte{0})}
	_, err := readHeader(transport)
	assert.Error(err)
}

// Ensures readHeader returns an error for an unsupported frugal frame
// encoding version.
func TestReadHeaderUnsupportedVersion(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer([]byte{0x01, 0, 0, 0, 0})}
	expectedErr := NewFProtocolExceptionWithType(thrift.BAD_VERSION, "frugal: unsupported protocol version 1")
	_, err := readHeader(transport)
	assert.Equal(expectedErr, err)
}

// Ensures readHeader returns an error for a frugal frame with an incorrectly
// encoded length.
func TestReadHeaderBadLength(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer([]byte{protocolV0, 0, 0, 0, 1})}
	_, err := readHeader(transport)
	assert.Error(err)
}

// Ensures readHeader correctly reads properly encoded frugal headers.
func TestReadHeader(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(basicFrame)}

	headers, err := readHeader(transport)
	assert.Nil(err)
	assert.Equal(basicHeaders, headers)
}

// Ensures getHeadersFromFrame returns an error for frames with invalid size.
func TestGetHeadersFromFrameInvalidSize(t *testing.T) {
	assert := assert.New(t)
	expectedErr := NewFProtocolExceptionWithType(thrift.INVALID_DATA, "frugal: invalid frame size 1")
	_, err := getHeadersFromFrame([]byte{0})
	assert.Equal(expectedErr, err)
}

// Ensures getHeadersFromeFrame returns an error for an unsupported frugal
// frame encoding version.
func TestGetHeadersFromFrameUnsupportedVersion(t *testing.T) {
	assert := assert.New(t)
	expectedErr := NewFProtocolExceptionWithType(thrift.BAD_VERSION, "frugal: unsupported protocol version 1")
	_, err := getHeadersFromFrame([]byte{0x01, 0, 0, 0, 0})
	assert.Equal(expectedErr, err)
}

// Ensures getHeadersFromFrame properly decodes frugal headers from frame.
func TestGetHeadersFromFrame(t *testing.T) {
	assert := assert.New(t)
	headers, err := getHeadersFromFrame(basicFrame)
	assert.Nil(err)
	assert.Equal(basicHeaders, headers)
}
