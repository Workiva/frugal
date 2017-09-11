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
	"bytes"
	"fmt"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/go-nats"
)

const (
	natsMaxMessageSize = 1024 * 1024
	frugalPrefix       = "frugal."
)

// NewFNatsTransport returns a new FTransport which uses the NATS messaging
// system as the underlying transport. This FTransport is stateless in that
// there is no connection maintained between the client and server. A request
// is simply published to a subject and responses are received on another
// subject. This requires requests and responses to fit within a single NATS
// message.
func NewFNatsTransport(conn *nats.Conn, subject, inbox string) FTransport {
	if inbox == "" {
		inbox = nats.NewInbox()
	}
	return &fNatsTransport{
		// FTransports manually frame messages.
		// Leave enough room for frame size.
		fBaseTransport: newFBaseTransport(natsMaxMessageSize - 4),
		conn:           conn,
		subject:        subject,
		inbox:          inbox,
	}
}

// fNatsTransport implements FTransport. This is a "stateless" transport in the
// sense that there is no connection with a server. A request is simply
// published to a subject and responses are received on another subject.
// This assumes requests/responses fit within a single NATS message.
type fNatsTransport struct {
	*fBaseTransport
	conn    *nats.Conn
	subject string
	inbox   string
	sub     *nats.Subscription
}

// Open subscribes to the configured inbox subject.
func (f *fNatsTransport) Open() error {
	if f.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN,
			fmt.Sprintf("frugal: NATS not connected, has status %d", f.conn.Status()))
	}
	if f.sub != nil {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_ALREADY_OPEN,
			"frugal: NATS transport already open")
	}

	handler := f.handler

	sub, err := f.conn.Subscribe(f.inbox, handler)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = sub

	f.fBaseTransport.Open()
	return nil
}

// handler receives a NATS message and executes the frame
func (f *fNatsTransport) handler(msg *nats.Msg) {
	if err := f.fBaseTransport.ExecuteFrame(msg.Data); err != nil {
		logger().Warn("Could not execute frame", err)
	}
}

// Returns true if the transport is open
func (f *fNatsTransport) IsOpen() bool {
	return f.sub != nil && f.conn.Status() == nats.CONNECTED
}

// Close unsubscribes from the inbox subject.
func (f *fNatsTransport) Close() error {
	if f.sub == nil {
		return nil
	}
	if err := f.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = nil

	f.fBaseTransport.Close(nil)
	return nil
}

func (f *fNatsTransport) checkMessageSize(data []byte) error {
	if len(data) > natsMaxMessageSize {
		return thrift.NewTTransportException(
			TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE,
			fmt.Sprintf("Message exceeds %d bytes, was %d bytes", natsMaxMessageSize, len(data)))
	}
	return nil
}

// Oneway transmits the given data and doesn't wait for a response.
// Implementations of oneway should be threadsafe and respect the timeout
// present on the context.
func (f *fNatsTransport) Oneway(ctx FContext, data []byte) error {
	if !f.IsOpen() {
		return f.getClosedConditionError("request:")
	}

	if len(data) == 4 {
		return nil
	}

	if err := f.checkMessageSize(data); err != nil {
		return err
	}

	return f.conn.PublishRequest(f.subject, f.inbox, data)
}

// Request transmits the given data and waits for a response.
// Implementations of request should be threadsafe and respect the timeout
// present the on context. The data is expected to already be framed.
func (f *fNatsTransport) Request(ctx FContext, data []byte) (thrift.TTransport, error) {
	resultC := make(chan []byte, 1)

	if !f.IsOpen() {
		return nil, f.getClosedConditionError("request:")
	}

	if len(data) == 4 {
		return nil, nil
	}

	if err := f.registry.Register(ctx, resultC); err != nil {
		return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN, err.Error())
	}
	defer f.registry.Unregister(ctx)

	if err := f.checkMessageSize(data); err != nil {
		return nil, err
	}

	if err := f.conn.PublishRequest(f.subject, f.inbox, data); err != nil {
		return nil, err
	}

	select {
	case result := <-resultC:
		return &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(result)}, nil
	case <-time.After(ctx.Timeout()):
		return nil, thrift.NewTTransportException(TRANSPORT_EXCEPTION_TIMED_OUT, "frugal: nats request timed out")
	}
}

// GetRequestSizeLimit returns the maximum number of bytes that can be
// transmitted. Returns a non-positive number to indicate an unbounded
// allowable size.
func (f *fNatsTransport) GetRequestSizeLimit() uint {
	return uint(natsMaxMessageSize)
}

// This is a no-op for fNatsTransport
func (f *fNatsTransport) SetMonitor(monitor FTransportMonitor) {
}

func (f *fNatsTransport) getClosedConditionError(prefix string) error {
	if f.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_NOT_OPEN,
			fmt.Sprintf("%s stateless NATS client not connected (has status code %d)", prefix, f.conn.Status()))
	}
	return thrift.NewTTransportException(TRANSPORT_EXCEPTION_NOT_OPEN,
		fmt.Sprintf("%s stateless NATS service TTransport not open", prefix))
}
