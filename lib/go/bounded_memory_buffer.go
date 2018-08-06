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
	bigEndianPutUint32(data, uint32(len(data)-4))
	return data
}

// HasWriteData determines if there's any data in the buffer to send.
func (f *TMemoryOutputBuffer) HasWriteData() bool {
	return f.Len() > 4
}
