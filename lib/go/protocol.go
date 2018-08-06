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
	"errors"
	"fmt"
	"io"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const protocolV0 = 0x00

var (
	writeMarshaler = v0Marshaler
	v0Marshaler    = &v0ProtocolMarshaler{}
)

type frameComponents struct {
	frameSize       uint32
	protocolVersion byte
	headers         map[string]string
	payload         []byte
}

// protocolMarshaler is responsible for serializing and deserializing the
// Frugal protocol for a specific version.
type protocolMarshaler interface {
	// marshalHeaders serializes the given headers map to a byte slice.
	marshalHeaders(headers map[string]string) []byte

	// unmarshalHeaders reads serialized headers from the reader into a map.
	unmarshalHeaders(reader io.Reader) (map[string]string, error)

	// unmarshalHeadersFromFrame reads serialized headers from the byte slice
	// into a map.
	unmarshalHeadersFromFrame(frame []byte) (map[string]string, error)

	// addHeadersToFrame returns a new frame containing the given headers. This
	// assumes the frame still has the frame size header at the beginning.
	addHeadersToFrame(frame []byte, headers map[string]string) ([]byte, error)

	// unmarshalFrame deserializes the byte slice into frame components.
	unmarshalFrame(frame []byte, components *frameComponents) error
}

// getMarshaler returns a protocolMarshaler for the given protocol version.
// An error is returned if the version is not supported.
func getMarshaler(version byte) (protocolMarshaler, error) {
	switch version {
	case protocolV0:
		return v0Marshaler, nil
	default:
		return nil, thrift.NewTProtocolExceptionWithType(
			thrift.BAD_VERSION, fmt.Errorf("frugal: unsupported protocol version %d", version))
	}
}

// FProtocolFactory creates new FProtocol instances. It takes a
// TProtocolFactory and a TTransport and returns an FProtocol which wraps a
// TProtocol produced by the TProtocolFactory. The TProtocol itself wraps the
// provided TTransport. This makes it easy to produce an FProtocol which uses
// any existing Thrift transports and protocols in a composable manner.
type FProtocolFactory struct {
	protoFactory thrift.TProtocolFactory
}

// NewFProtocolFactory creates a new FProtocolFactory with the given
// TProtocolFactory.
func NewFProtocolFactory(protoFactory thrift.TProtocolFactory) *FProtocolFactory {
	return &FProtocolFactory{protoFactory}
}

// GetProtocol returns a new FProtocol instance using the given TTransport.
func (f *FProtocolFactory) GetProtocol(tr thrift.TTransport) *FProtocol {
	return &FProtocol{f.protoFactory.GetProtocol(tr)}
}

// FProtocol is Frugal's equivalent of Thrift's TProtocol. It defines the
// serialization protocol used for messages, such as JSON, binary, etc.
// FProtocol actually extends TProtocol and adds support for serializing
// FContext. In practice, FProtocol simply wraps a TProtocol and uses Thrift's
// built-in serialization. FContext is encoded before the TProtocol
// serialization of the message using a simple binary protocol. See the
// protocol documentation for more details.
type FProtocol struct {
	thrift.TProtocol
}

// WriteRequestHeader writes the request headers set on the given Context
// into the protocol
func (f *FProtocol) WriteRequestHeader(ctx FContext) error {
	return f.writeHeader(ctx.RequestHeaders())
}

// ReadRequestHeader reads the request headers on the protocol into a
// returned Context
func (f *FProtocol) ReadRequestHeader() (FContext, error) {
	headers, err := readHeader(f.Transport())
	if err != nil {
		return nil, err
	}

	ctx := &FContextImpl{
		requestHeaders:  make(map[string]string),
		responseHeaders: make(map[string]string),
	}

	for name, value := range headers {
		if name == opIDHeader {
			continue
		}
		ctx.AddRequestHeader(name, value)
	}

	// Put op id in response headers
	opid, ok := headers[opIDHeader]
	if !ok {
		return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, errors.New("frugal: request missing op id"))
	}
	setResponseOpID(ctx, opid)

	// Put a new opid in the request headers so this context
	// can be used/propagated on the receiver
	ctx.AddRequestHeader(opIDHeader, getNextOpID())

	cid := ctx.CorrelationID()
	if cid != "" {
		ctx.AddResponseHeader(cidHeader, cid)
	}

	return ctx, nil
}

// WriteResponseHeader writes the response headers set on the given Context
// into the protocol
func (f *FProtocol) WriteResponseHeader(ctx FContext) error {
	return f.writeHeader(ctx.ResponseHeaders())
}

// ReadResponseHeader reads the response headers on the protocol into a
// provided Context
func (f *FProtocol) ReadResponseHeader(ctx FContext) error {
	headers, err := readHeader(f.Transport())
	if err != nil {
		return err
	}

	for name, value := range headers {
		// Don't want to overwrite the opid header we set for a
		// propagated response
		if name == opIDHeader {
			continue
		}
		ctx.AddResponseHeader(name, value)
	}

	return nil
}

// writeHeader serializes the headers and writes them to the underlying
// transport.
func (f *FProtocol) writeHeader(headers map[string]string) error {
	buff := writeMarshaler.marshalHeaders(headers)
	if n, err := f.Transport().Write(buff); err != nil {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
			fmt.Sprintf("frugal: error writing protocol headers in writeHeader: %s", err))
	} else if n != len(buff) {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
			"frugal: failed to write complete protocol headers")
	}

	return nil
}

// readHeader deserializes headers from the given Reader.
func readHeader(reader io.Reader) (map[string]string, error) {
	buff := make([]byte, 1)
	if _, err := io.ReadFull(reader, buff); err != nil {
		if e, ok := err.(thrift.TTransportException); ok && e.TypeId() == TRANSPORT_EXCEPTION_END_OF_FILE {
			return nil, err
		}
		return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
			fmt.Sprintf("frugal: error reading protocol headers in readHeader: %s", err))
	}

	marshaler, err := getMarshaler(buff[0])
	if err != nil {
		return nil, err
	}

	return marshaler.unmarshalHeaders(reader)
}

// getHeadersFromFrame deserializes headers from the frame into a map.
func getHeadersFromFrame(frame []byte) (map[string]string, error) {
	// Need at least 1 byte for the version.
	if len(frame) == 0 {
		return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, errors.New("frugal: invalid frame size 0"))
	}

	marshaler, err := getMarshaler(frame[0])
	if err != nil {
		return nil, err
	}

	return marshaler.unmarshalHeadersFromFrame(frame[1:])
}

// addHeadersToFrame returns a new frame containing the given headers. This
// assumes the frame still has the frame size header at the beginning.
func addHeadersToFrame(frame []byte, headers map[string]string) ([]byte, error) {
	// Need at least 5 bytes (4 for size and 1 for version).
	if len(frame) < 5 {
		return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
			fmt.Errorf("frugal: invalid frame size %d", len(frame)))
	}

	marshaler, err := getMarshaler(frame[4])
	if err != nil {
		return nil, err
	}

	return marshaler.addHeadersToFrame(frame, headers)
}

// unmarshalFrame deserializes the byte slice into frame components. This
// assumes the frame still has the frame size header at the beginning.
func unmarshalFrame(frame []byte) (*frameComponents, error) {
	// Need at least 5 bytes (4 for size and 1 for version).
	if len(frame) < 5 {
		return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
			fmt.Errorf("frugal: invalid frame size %d", len(frame)))
	}

	// Read frame size.
	frameSize := bigEndianUint32(frame)

	if uint32(len(frame[4:])) != frameSize {
		return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
			fmt.Errorf("frugal: frame size %d does not match actual size %d", frameSize, len(frame[4:])))
	}

	marshaler, err := getMarshaler(frame[4])
	if err != nil {
		return nil, err
	}

	components := &frameComponents{
		frameSize:       frameSize,
		protocolVersion: frame[4],
	}
	return components, marshaler.unmarshalFrame(frame[5:], components)
}

// v0ProtocolMarshaler implements the protocolMarshaler interface for v0 of the
// Frugal protocol.
type v0ProtocolMarshaler struct{}

// marshalHeaders serializes the given headers map to a byte slice.
func (v *v0ProtocolMarshaler) marshalHeaders(headers map[string]string) []byte {
	size := v.calculateHeaderSize(headers)

	// Header buff = [version (1 byte), size (4 bytes), headers (size bytes)]
	// Headers = [size (4 bytes) name (size bytes) size (4 bytes) value (size bytes)*]
	buff := make([]byte, size+5)

	// Write version
	buff[0] = protocolV0

	// Write size
	bigEndianPutUint32(buff[1:5], uint32(size))

	// Write headers
	i := 5
	for name, value := range headers {
		bigEndianPutUint32(buff[i:i+4], uint32(len(name)))
		i += 4
		i += copy(buff[i:], name)
		bigEndianPutUint32(buff[i:i+4], uint32(len(value)))
		i += 4
		i += copy(buff[i:], value)
	}

	return buff
}

// unmarshalHeaders reads headers from the reader into a map.
func (v *v0ProtocolMarshaler) unmarshalHeaders(reader io.Reader) (map[string]string, error) {
	buff := make([]byte, 4)
	if _, err := io.ReadFull(reader, buff); err != nil {
		if e, ok := err.(thrift.TTransportException); ok && e.TypeId() == TRANSPORT_EXCEPTION_END_OF_FILE {
			return nil, err
		}
		return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
			fmt.Sprintf("frugal: error reading protocol headers in unmarshalHeaders reading header size: %s", err))
	}
	size := int32(bigEndianUint32(buff))
	buff = make([]byte, size)
	if _, err := io.ReadFull(reader, buff); err != nil {
		if e, ok := err.(thrift.TTransportException); ok && e.TypeId() == TRANSPORT_EXCEPTION_END_OF_FILE {
			return nil, err
		}
		return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
			fmt.Sprintf("frugal: error reading protocol headers in unmarshalHeaders reading headers: %s", err))
	}

	return v.readPairs(buff, 0, size)
}

// unmarshalHeadersFromFrame reads serialized headers from the byte slice into
// a map.
func (v *v0ProtocolMarshaler) unmarshalHeadersFromFrame(frame []byte) (map[string]string, error) {
	// Need at least 4 bytes for headers size.
	if len(frame) < 4 {
		return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
			fmt.Errorf("frugal: invalid v0 frame size %d", len(frame)))
	}
	size := int32(bigEndianUint32(frame))
	if size > int32(len(frame[4:])) {
		return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
			fmt.Errorf("frugal: v0 frame size %d does not match actual size %d", size, len(frame[4:])))
	}
	return v.readPairs(frame, 4, size+4)
}

// addHeadersToFrame returns a new frame containing the given headers. This
// assumes the frame still has the frame size header at the beginning.
func (v *v0ProtocolMarshaler) addHeadersToFrame(frame []byte, headers map[string]string) ([]byte, error) {
	existing, err := v.unmarshalHeadersFromFrame(frame[5:])
	if err != nil {
		return nil, err
	}
	// QUESTION: If a header already exists, should we overwrite it or return
	// an error?
	for name, value := range headers {
		existing[name] = value
	}
	serializedHeaders := v.marshalHeaders(existing)
	oldHeadersSize := int32(bigEndianUint32(frame[5:]))
	frameSize := v.calculateHeaderSize(existing) + int32(len(frame)) - oldHeadersSize
	buff := make([]byte, frameSize)

	// Add frame size.
	bigEndianPutUint32(buff, uint32(frameSize-4))

	// Add headers (version is included).
	offset := copy(buff[4:], serializedHeaders)

	// Add payload.
	copy(buff[4+offset:], frame[9+oldHeadersSize:])

	return buff, nil
}

// unmarshalFrame deserializes the byte slice into frame components.
func (v *v0ProtocolMarshaler) unmarshalFrame(frame []byte, components *frameComponents) error {
	headers, err := v.unmarshalHeadersFromFrame(frame)
	if err != nil {
		return err
	}

	components.headers = headers
	components.payload = frame[v.calculateHeaderSize(headers)+8:]

	return nil
}

func (v *v0ProtocolMarshaler) readPairs(buff []byte, start, end int32) (map[string]string, error) {
	headers := make(map[string]string)
	i := start
	for i < end {
		// Read header name.
		nameSize := int32(bigEndianUint32(buff[i : i+4]))
		i += 4
		if i > end || i+nameSize > end {
			return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
				errors.New("frugal: invalid v0 protocol header name"))
		}
		name := string(buff[i : i+nameSize])
		i += nameSize

		// Read header value.
		valueSize := int32(bigEndianUint32(buff[i : i+4]))
		i += 4
		if i > end || i+valueSize > end {
			return nil, thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
				errors.New("frugal: invalid v0 protocol header value"))
		}
		value := string(buff[i : i+valueSize])
		i += valueSize

		headers[name] = value
	}
	return headers, nil
}

func (v *v0ProtocolMarshaler) calculateHeaderSize(headers map[string]string) int32 {
	size := int32(0)
	for name, value := range headers {
		size += int32(8 + len(name) + len(value))
	}
	return size
}
