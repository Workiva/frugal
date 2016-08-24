package frugal

import (
	"bytes"
	"encoding/binary"
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
	completeFrugalFrame = []byte{0, 0, 0, 85, 0, 0, 0, 0, 59, 0, 0, 0, 5, 95, 111, 112, 105,
		100, 0, 0, 0, 1, 48, 0, 0, 0, 4, 95, 99, 105, 100, 0, 0, 0, 5, 49, 50, 51,
		52, 53, 0, 0, 0, 3, 102, 111, 111, 0, 0, 0, 3, 98, 97, 114, 0, 0, 0, 3, 98,
		97, 122, 0, 0, 0, 3, 113, 117, 120, 0, 0, 0, 17, 116, 104, 105, 115, 32,
		105, 115, 32, 97, 32, 114, 101, 113, 117, 101, 115, 116}
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
	mft.AssertExpectations(t)
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
	mft.AssertExpectations(t)
}

// Ensures writeHeader properly encodes header bytes.
func TestWriteHeader(t *testing.T) {
	assert := assert.New(t)
	mft := &mockFTransport{}
	mft.On("Write", basicFrame).Return(len(basicFrame), nil)
	proto := &FProtocol{tProtocolFactory.GetProtocol(mft)}
	assert.Nil(proto.writeHeader(basicHeaders))
	mft.AssertExpectations(t)
}

// Ensures WriteRequestHeader properly encodes header bytes and
// ReadRequestHeader properly decodes them.
func TestWriteReadRequestHeader(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: &bytes.Buffer{}}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}
	ctx := NewFContext("123")
	ctx.AddRequestHeader("hello", "world")
	ctx.AddRequestHeader("foo", "bar")
	assert.Nil(proto.WriteRequestHeader(ctx))
	ctx, err := proto.ReadRequestHeader()
	assert.Nil(err)
	header, ok := ctx.RequestHeader("hello")
	assert.True(ok)
	assert.Equal("world", header)
	header, ok = ctx.RequestHeader("foo")
	assert.True(ok)
	assert.Equal("bar", header)
	assert.Equal("123", ctx.CorrelationID())
	assert.Equal(uint64(0), ctx.opID())
}

// Ensures WriteResponseHeader properly encodes header bytes and
// ReadResponseHeader properly decodes them.
func TestWriteReadResponseHeader(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: &bytes.Buffer{}}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}
	ctx := NewFContext("123")
	ctx.AddResponseHeader("hello", "world")
	ctx.AddResponseHeader("foo", "bar")
	assert.Nil(proto.WriteResponseHeader(ctx))
	ctx = NewFContext("123")
	err := proto.ReadResponseHeader(ctx)
	assert.Nil(err)
	header, ok := ctx.ResponseHeader("hello")
	assert.True(ok)
	assert.Equal("world", header)
	header, ok = ctx.ResponseHeader("foo")
	assert.True(ok)
	assert.Equal("bar", header)
	assert.Equal("123", ctx.CorrelationID())
	assert.Equal(uint64(0), ctx.opID())
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
	expectedErr := NewFProtocolExceptionWithType(thrift.INVALID_DATA, "frugal: invalid v0 frame size 0")
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

// Ensures addHeadersToFrame returns a new frame with the headers added.
func TestAddHeadersToFrame(t *testing.T) {
	assert := assert.New(t)
	oldComponents, err := unmarshalFrame(completeFrugalFrame)
	assert.Nil(err)
	assert.Equal(byte(protocolV0), oldComponents.protocolVersion)
	assert.Equal("bar", oldComponents.headers["foo"])
	assert.Equal("qux", oldComponents.headers["baz"])
	assert.Equal("0", oldComponents.headers["_opid"])
	assert.Equal("12345", oldComponents.headers["_cid"])
	assert.Equal([]byte("this is a request"), oldComponents.payload)

	headers := map[string]string{"bat": "man", "spider": "person"}
	newFrame, err := addHeadersToFrame(completeFrugalFrame, headers)
	assert.Nil(err)

	newComponents, err := unmarshalFrame(newFrame)
	assert.Nil(err)
	assert.Equal(byte(protocolV0), newComponents.protocolVersion)
	assert.Equal("bar", newComponents.headers["foo"])
	assert.Equal("qux", newComponents.headers["baz"])
	assert.Equal("0", newComponents.headers["_opid"])
	assert.Equal("12345", newComponents.headers["_cid"])
	assert.Equal("man", newComponents.headers["bat"])
	assert.Equal("person", newComponents.headers["spider"])
	assert.Equal([]byte("this is a request"), newComponents.payload)
}

// Ensures addHeadersToFrame returns an error if the frame size is less than 5.
func TestAddHeadersToFrameShortFrame(t *testing.T) {
	assert := assert.New(t)
	_, err := addHeadersToFrame(make([]byte, 3), make(map[string]string))
	assert.NotNil(err)
	assert.Equal(thrift.INVALID_DATA, err.(FProtocolException).TypeId())
}

// Ensures addHeadersToFrame returns an error if the frame has an unsupported
// protocol version.
func TestAddHeadersToFrameBadVersion(t *testing.T) {
	assert := assert.New(t)
	frame := make([]byte, 10)
	frame[4] = 0xFF
	_, err := addHeadersToFrame(frame, make(map[string]string))
	assert.NotNil(err)
	assert.Equal(thrift.BAD_VERSION, err.(FProtocolException).TypeId())
}

// Ensures addHeadersToFrame returns an error if the frame has an incorrect
// frame size.
func TestAddHeadersToFrameBadFrameSize(t *testing.T) {
	assert := assert.New(t)
	frame := make([]byte, 10)
	binary.BigEndian.PutUint32(frame[5:], 9000)
	frame[4] = protocolV0
	_, err := addHeadersToFrame(frame, make(map[string]string))
	assert.NotNil(err)
	assert.Equal(thrift.INVALID_DATA, err.(FProtocolException).TypeId())
}

// Ensures headers with non-ascii characters can be encodeded and decoded
// properly.
func TestMarshalUnmarshalHeadersUTF8(t *testing.T) {
	assert := assert.New(t)
	marshaler := v0Marshaler
	headers := map[string]string {
		"Đ¥ÑØ": "δάüΓ",
		"good\u00F1ight": "moo\u00F1",
	}
	encodedHeaders := marshaler.marshalHeaders(headers)
	decodedHeaders, err := marshaler.unmarshalHeadersFromFrame(encodedHeaders[1:])
	assert.Nil(err)
	assert.Equal(headers, decodedHeaders)
}

func BenchmarkAddHeadersToFrame(b *testing.B) {
	headers := map[string]string{"bat": "man", "spider": "man", "super": "man"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addHeadersToFrame(completeFrugalFrame, headers)
	}
}
