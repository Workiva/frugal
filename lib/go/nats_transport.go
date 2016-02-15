package frugal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

const (
	// NATS limits messages to 1MB.
	natsMaxMessageSize = 1024 * 1024
	disconnect         = "DISCONNECT"
	natsV0             = 0
)

// natsServiceTTransport implements thrift.TTransport.
type natsServiceTTransport struct {
	conn                *nats.Conn
	listenTo            string
	writeTo             string
	reader              *io.PipeReader
	writer              *io.PipeWriter
	writeBuffer         *bytes.Buffer
	sub                 *nats.Subscription
	heartbeatSub        *nats.Subscription
	heartbeatListen     string
	heartbeatReply      string
	heartbeatInterval   time.Duration
	recvHeartbeat       chan struct{}
	closed              chan struct{}
	isOpen              bool
	mutex               sync.RWMutex
	connectSubject      string
	connectTimeout      time.Duration
	maxMissedHeartbeats uint
}

// NewNatsServiceTTransport returns a new thrift TTransport which uses
// the NATS messaging system as the underlying transport. It performs a
// handshake with a server listening on the given NATS subject upon open.
// This TTransport can only be used with FNatsServer. Message frames are
// limited to 1MB in size.
func NewNatsServiceTTransport(conn *nats.Conn, subject string,
	timeout time.Duration, maxMissedHeartbeats uint) thrift.TTransport {

	return &natsServiceTTransport{
		conn:                conn,
		connectSubject:      subject,
		connectTimeout:      timeout,
		maxMissedHeartbeats: maxMissedHeartbeats,
	}
}

// newNatsServiceTTransportServer returns a new thrift TTransport which uses
// the NATS messaging system as the underlying transport. This TTransport can
// only be used with FNatsServer.
func newNatsServiceTTransportServer(conn *nats.Conn, listenTo, writeTo string) thrift.TTransport {
	return &natsServiceTTransport{
		conn:     conn,
		listenTo: listenTo,
		writeTo:  writeTo,
	}
}

// Open handshakes with the server (if this is a client transport) initializes
// the write buffer and reader/writer pipe, subscribes to the specified
// subject, and starts heartbeating.
func (n *natsServiceTTransport) Open() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: NATS not connected, has status %d", n.conn.Status()))
	}

	if n.isOpen {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS transport already open")
	}

	// Handshake if this is a client.
	if n.connectSubject != "" {
		if err := n.handshake(); err != nil {
			return thrift.NewTTransportExceptionFromError(err)
		}
	}

	if n.listenTo == "" || n.writeTo == "" {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"frugal: listenTo and writeTo cannot be empty")
	}

	n.closed = make(chan struct{})
	n.writeBuffer = bytes.NewBuffer(make([]byte, 0, natsMaxMessageSize))

	n.reader, n.writer = io.Pipe()

	sub, err := n.conn.Subscribe(n.listenTo, func(msg *nats.Msg) {
		if msg.Reply == disconnect {
			// Remote client is disconnecting.
			n.Close()
			return
		}
		n.writer.Write(msg.Data)
	})
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	n.sub = sub

	// Handle heartbeats.
	if n.heartbeatInterval > 0 {
		hbSub, err := n.conn.Subscribe(n.heartbeatListen, func(msg *nats.Msg) {
			select {
			case n.recvHeartbeat <- struct{}{}:
			default:
			}
			n.conn.Publish(n.heartbeatReply, nil)
		})
		if err != nil {
			n.Close()
			return thrift.NewTTransportExceptionFromError(err)
		}
		n.heartbeatSub = hbSub
		go func() {
			missed := uint(0)
			for {
				select {
				case <-time.After(n.heartbeatInterval):
					missed++
					if missed >= n.maxMissedHeartbeats {
						log.Println("frugal: server heartbeat expired")
						n.Close()
						return
					}
				case <-n.recvHeartbeat:
					missed = 0
				case <-n.closed:
					return
				}
			}
		}()
	}
	n.isOpen = true
	return nil
}

type natsConnectionHandshake struct {
	Version uint8 `json:"version"`
}

func (n *natsServiceTTransport) handshake() error {
	hs := &natsConnectionHandshake{Version: natsV0}
	hsBytes, err := json.Marshal(hs)
	if err != nil {
		return err
	}
	msg, err := n.handshakeRequest(hsBytes)
	if err != nil {
		return err
	}

	if msg.Reply == "" {
		return errors.New("frugal: no reply subject on connect")
	}

	// Connect message consists of "[heartbeat subject] [heartbeat reply subject] [expected interval ms]"
	subjects := strings.Split(string(msg.Data), " ")
	if len(subjects) != 3 {
		return errors.New("frugal: invalid connect message")
	}
	var (
		heartbeatListen = subjects[0]
		heartbeatReply  = subjects[1]
		deadline, err2  = strconv.ParseInt(subjects[2], 10, 64)
	)
	if err2 != nil {
		return err2
	}
	var interval time.Duration
	if deadline > 0 {
		interval = time.Millisecond * time.Duration(deadline)
	}

	n.heartbeatListen = heartbeatListen
	n.heartbeatReply = heartbeatReply
	n.heartbeatInterval = interval
	n.recvHeartbeat = make(chan struct{}, 1)
	n.listenTo = msg.Subject
	n.writeTo = msg.Reply
	return nil
}

func (n *natsServiceTTransport) handshakeRequest(hsBytes []byte) (m *nats.Msg, err error) {
	inbox := newFrugalInbox()
	var s *nats.Subscription
	s, err = n.conn.SubscribeSync(inbox)
	if err != nil {
		return
	}
	s.AutoUnsubscribe(1)
	err = n.conn.PublishRequest(n.connectSubject, inbox, hsBytes)
	if err == nil {
		m, err = s.NextMsg(n.connectTimeout)
	}
	s.Unsubscribe()
	return
}

func (n *natsServiceTTransport) IsOpen() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.conn.Status() == nats.CONNECTED && n.isOpen
}

func (n *natsServiceTTransport) getClosedConditionError(prefix string) error {
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.NOT_OPEN,
			fmt.Sprintf("%s NATS client not connected (has status code %d)", prefix, n.conn.Status()))
	}
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s NATS service TTransport not open", prefix))
}

// Close unsubscribes, signals the remote peer, and stops heartbeating.
func (n *natsServiceTTransport) Close() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if !n.isOpen {
		return nil
	}

	// Signal remote peer for a graceful disconnect.
	n.conn.PublishRequest(n.writeTo, disconnect, nil)
	if err := n.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	if n.heartbeatSub != nil {
		if err := n.heartbeatSub.Unsubscribe(); err != nil {
			return thrift.NewTTransportExceptionFromError(err)
		}
	}

	// Flush the NATS connection to avoid an edge case where the program exits
	// after closing the transport. This is because NATS asynchronously flushes
	// in the background, so explicitly flushing prevents us from losing
	// anything buffered when we exit.
	n.conn.Flush()

	n.sub = nil
	n.heartbeatSub = nil
	close(n.closed)
	n.isOpen = false
	return thrift.NewTTransportExceptionFromError(n.writer.Close())
}

func (n *natsServiceTTransport) Read(p []byte) (int, error) {
	if !n.IsOpen() {
		return 0, n.getClosedConditionError("read:")
	}
	num, err := n.reader.Read(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

// Write the bytes to a buffer. Returns ErrTooLarge if the buffer exceeds 1MB.
func (n *natsServiceTTransport) Write(p []byte) (int, error) {
	if !n.IsOpen() {
		return 0, n.getClosedConditionError("write:")
	}
	if len(p)+n.writeBuffer.Len() > natsMaxMessageSize {
		n.writeBuffer.Reset() // Clear any existing bytes.
		return 0, ErrTooLarge
	}
	num, err := n.writeBuffer.Write(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

// Flush sends the buffered bytes over NATS. Returns ErrTooLarge if the number
// of bytes exceed 1MB.
func (n *natsServiceTTransport) Flush() error {
	if !n.IsOpen() {
		return n.getClosedConditionError("flush:")
	}
	defer n.writeBuffer.Reset()
	data := n.writeBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}
	if len(data) > natsMaxMessageSize {
		return ErrTooLarge
	}
	err := n.conn.Publish(n.writeTo, data)
	return thrift.NewTTransportExceptionFromError(err)
}

func (n *natsServiceTTransport) RemainingBytes() uint64 {
	return ^uint64(0) // We don't know unless framed is used.
}
