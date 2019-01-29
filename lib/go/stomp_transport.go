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
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/go-stomp/stomp"
	"sync"
)

type ReconnectHandler func() (error, *stomp.Conn)

// FStompPublisherTransportFactory creates fStompPublisherTransports.
type FStompPublisherTransportFactory struct {
	conn           *stomp.Conn
	maxPublishSize int
	topicPrefix    string
	reconHandler   ReconnectHandler
}

// NewFStompPublisherTransportFactory creates an FStompPublisherTransportFactory using the
// provided stomp connection.
func NewFStompPublisherTransportFactory(conn *stomp.Conn, maxPublishSize int, topicPrefix string, recon ReconnectHandler) *FStompPublisherTransportFactory {
	return &FStompPublisherTransportFactory{conn: conn, maxPublishSize: maxPublishSize, topicPrefix: topicPrefix, reconHandler: recon}
}

// GetTransport creates a new stomp FPublisherTransport.
func (m *FStompPublisherTransportFactory) GetTransport() FPublisherTransport {
	return NewStompFPublisherTransport(m.conn, m.maxPublishSize, m.topicPrefix, m.reconHandler)
}

// fStompPublisherTransport implements FPublisherTransport.
type fStompPublisherTransport struct {
	conn           *stomp.Conn
	maxPublishSize int
	topicPrefix    string
	reconHandler   ReconnectHandler
}

// NewStompFPublisherTransport creates a new FPublisherTransport which is used for
// publishing using stomp protocol with scopes.
func NewStompFPublisherTransport(conn *stomp.Conn, maxPublishSize int, topicPrefix string, recon ReconnectHandler) FPublisherTransport {
	return &fStompPublisherTransport{conn: conn, maxPublishSize: maxPublishSize, topicPrefix: topicPrefix, reconHandler: recon}
}

// Open initializes the transport.
func (m *fStompPublisherTransport) Open() error {
	if m.conn == nil {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_NOT_OPEN, "frugal: mq transport not open")
	}
	return nil
}

// IsOpen returns true if the transport is open, false otherwise.
func (m *fStompPublisherTransport) IsOpen() bool {
	return m.conn != nil
}

// Close closes the transport.
func (m *fStompPublisherTransport) Close() error {
	return nil
}

// GetPublishSizeLimit returns the maximum allowable size of a payload
// to be published. 0 is returned to indicate an unbounded allowable size.
func (m *fStompPublisherTransport) GetPublishSizeLimit() uint {
	return uint(m.maxPublishSize)
}

// Publish sends the given payload with the transport.
func (m *fStompPublisherTransport) Publish(topic string, data []byte) error {
	if !m.IsOpen() {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_NOT_OPEN, "frugal: stomp transport not open")
	}

	if len(data) > m.maxPublishSize {
		return thrift.NewTTransportException(
			TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE,
			fmt.Sprintf("Message exceeds %d bytes, was %d bytes", m.maxPublishSize, len(data)))
	}

	destination := m.formatStompPublishTopic(topic)
	if err := m.conn.Send(destination, "application/octet-stream", data, stomp.SendOpt.Header("persistent", "true")); err != nil {
		if err.Error() == "connection already closed" {
			e, conn := m.reconHandler()
			if e != nil {
				return thrift.NewTTransportExceptionFromError(err)
			}
			m.conn = conn
			e = m.Publish(topic, data)
			if e != nil {
				return thrift.NewTTransportExceptionFromError(err)
			}
			return nil
		}
		return thrift.NewTTransportExceptionFromError(err)
	}
	return nil
}

func (m *fStompPublisherTransport) formatStompPublishTopic(topic string) string {
	return fmt.Sprintf("/topic/%s%s%s", m.topicPrefix, frugalPrefix, topic)
}

// FStompSubscribeTransportFactory creates fStompSubscriberTransports.
type FStompSubscribeTransportFactory struct {
	conn           *stomp.Conn
	consumerPrefix string
	useQueue       bool
	reconHandler   ReconnectHandler
}

// NewFStompSubscriberTransportFactory creates FStompSubscribeTransportFactory with the given stomp
// connection and consumer name.
func NewFStompSubscriberTransportFactory(conn *stomp.Conn, consumerPrefix string, useQueue bool, recon ReconnectHandler) *FStompSubscribeTransportFactory {
	return &FStompSubscribeTransportFactory{conn: conn, consumerPrefix: consumerPrefix, useQueue: useQueue, reconHandler: recon}
}

// GetTransport creates a new fStompSubscriberTransport.
func (m *FStompSubscribeTransportFactory) GetTransport() FSubscriberTransport {
	return NewStompFSubscriberTransport(m.conn, m.consumerPrefix, m.useQueue, m.reconHandler)
}

// fStompSubscriberTransport implements FSubscriberTransport.
type fStompSubscriberTransport struct {
	conn           *stomp.Conn
	consumerPrefix string
	topic          string
	useQueue       bool
	sub            *stomp.Subscription
	openMu         sync.RWMutex
	isSubscribed   bool
	callback       FAsyncCallback
	stopC          chan bool
	reconHandler   ReconnectHandler
}

// NewStompFSubscriberTransport creates a new FSubscriberTransport which is used for
// pub/sub.
func NewStompFSubscriberTransport(conn *stomp.Conn, consumerPrefix string, useQueue bool, recon ReconnectHandler) FSubscriberTransport {
	return &fStompSubscriberTransport{conn: conn, consumerPrefix: consumerPrefix, useQueue: useQueue, reconHandler: recon}
}

// Subscribe sets the subscribe topic and opens the transport.
func (m *fStompSubscriberTransport) Subscribe(topic string, callback FAsyncCallback) error {
	m.openMu.Lock()
	defer m.openMu.Unlock()

	if m.conn == nil {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_NOT_OPEN, "frugal: stomp transport not open")
	}

	if m.isSubscribed {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_ALREADY_OPEN, "frugal: stomp transport already has a subscription")
	}

	if topic == "" {
		return thrift.NewTTransportException(TRANSPORT_EXCEPTION_UNKNOWN, "frugal: stomp transport cannot subscribe to empty topic")
	}

	m.topic = topic
	var destination string
	if m.useQueue {
		destination = fmt.Sprintf("/queue/%s%s%s", m.consumerPrefix, frugalPrefix, topic)
	} else {
		destination = fmt.Sprintf("/topic/%s%s%s", m.consumerPrefix, frugalPrefix, topic)
	}

	sub, err := m.conn.Subscribe(destination, stomp.AckClientIndividual)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	m.stopC = make(chan bool, 1)
	m.sub = sub
	m.isSubscribed = true
	m.callback = callback
	go m.processMessages()
	return nil
}

// IsSubscribed returns true if the transport is subscribed to a topic, false
// otherwise.
func (m *fStompSubscriberTransport) IsSubscribed() bool {
	m.openMu.RLock()
	defer m.openMu.RUnlock()
	return m.conn != nil && m.isSubscribed
}

// Unsubscribe unsubscribes from the destination.
func (m *fStompSubscriberTransport) Unsubscribe() error {
	m.openMu.Lock()
	defer m.openMu.Unlock()
	if !m.isSubscribed {
		logger().Info("frugal: unable to unsubscribe, subscription already unsubscribed")
		return nil
	}

	close(m.stopC)
	if err := m.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}

	m.isSubscribed = false
	m.callback = nil
	return nil
}

func (m *fStompSubscriberTransport) reconnect() error {
	m.openMu.Lock()
	err, conn := m.reconHandler()

	if err != nil {
		return err
	}
	logger().Infof("reconnected to broker successfully")

	m.conn = conn
	m.isSubscribed = false
	m.openMu.Unlock()

	if err = m.Subscribe(m.topic, m.callback); err != nil {
		logger().Errorf("failed to resubscribe")
		return err
	}
	logger().Infof("resubscribed to topic successfully")
	return nil
}

// Processes messages from subscription channel with the given FAsyncCallback.
func (m *fStompSubscriberTransport) processMessages() {
	stopC := m.stopC
	for {
		select {
		case <-stopC:
			return
		case message, ok := <-m.sub.C:
			if !ok {
				logger().Warnf("frugal: message channel closed unexpectedly, will try to reconnect")
				err := m.reconnect()
				if err != nil {
					logger().Errorf("frugal: reconnect to broker failed, error is %s", err)
				}
				return
			}

			if len(message.Body) < 4 {
				logger().Warnf("frugal: %s", message.Body)
				logger().Warnf("frugal: discarding invalid scope message frame")
				continue
			}

			transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(message.Body[4:])}
			if err := m.callback(transport); err != nil {
				logger().Warn("frugal: error executing callback: ", err)
				continue
			}

			go m.ackMessage(message)
		}
	}
}

// Acknowledges the stomp message.
func (m *fStompSubscriberTransport) ackMessage(message *stomp.Message) {
	if err := m.conn.Ack(message); err != nil {
		logger().Errorf("frugal: error acking mq message: ", err.Error())
	}
}
