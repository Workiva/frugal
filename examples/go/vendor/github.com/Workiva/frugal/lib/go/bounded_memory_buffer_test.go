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
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// Ensures Write writes to the buffer until its size limit is reached, after
// which ErrTooLarge is returned and the buffer is reset.
func TestTFramedMemoryBufferWrite(t *testing.T) {
	buff := NewTMemoryOutputBuffer(100)
	assert.Equal(t, 4, buff.Len())
	n, err := buff.Write(make([]byte, 50))
	assert.Nil(t, err)
	assert.Equal(t, 50, n)
	n, err = buff.Write(make([]byte, 40))
	assert.Nil(t, err)
	assert.Equal(t, 40, n)
	assert.Equal(t, 94, buff.Len())
	_, err = buff.Write(make([]byte, 20))
	assert.True(t, IsErrTooLarge(err))
	assert.Equal(t, TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE, err.(thrift.TTransportException).TypeId())
	assert.Equal(t, 4, buff.Len())
}
