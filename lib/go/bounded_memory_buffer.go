package frugal

import (
	"encoding/binary"
	"fmt"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// TMemoryOutputBuffer implements TTransport using a bounded memory buffer.
// Writes which cause the buffer to exceed its size return ErrTooLarge.
// The TMemoryOutputBuffer handles framing data.
type TMemoryOutputBuffer struct {
	limit uint
	*thrift.TMemoryBuffer
}

var emptyFrameSize = []byte{0, 0, 0, 0}

// NewTMemoryOutputBuffer returns a new TFramedMemoryBuffer with the given
// size limit. If the provided limit is non-positive, the buffer is allowed
// to grow unbounded.
func NewTMemoryOutputBuffer(size uint) *TMemoryOutputBuffer {
	buffer := &TMemoryOutputBuffer{size, thrift.NewTMemoryBuffer()}
	buffer.Write(emptyFrameSize)
	return buffer
}

// Write the data to the buffer. Returns ErrTooLarge if the write would cause
// the buffer to exceed its limit.
func (f *TMemoryOutputBuffer) Write(buf []byte) (int, error) {
	if f.limit > 0 && uint(len(buf)+f.Len()) > f.limit {
		f.Reset()
		return 0, thrift.NewTTransportException(
			TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE,
			fmt.Sprintf("Buffer size reached (%d)", f.limit))
	}
	return f.TMemoryBuffer.Write(buf)
}

// Reset clears the buffer
func (f *TMemoryOutputBuffer) Reset() {
	f.TMemoryBuffer.Reset()
	f.Write(emptyFrameSize)
}

// Bytes retrieves the framed contents of the buffer.
func (f *TMemoryOutputBuffer) Bytes() []byte {
	data := f.TMemoryBuffer.Bytes()
	binary.BigEndian.PutUint32(data, uint32(len(data)-4))
	return data
}

// HasWriteData determines if there's any data in the buffer to send.
func (f *TMemoryOutputBuffer) HasWriteData() bool {
	return f.Len() > 4
}
