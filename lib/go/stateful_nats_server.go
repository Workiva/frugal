// TODO: Remove this with 2.0
package frugal

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
)

const (
	queue                      = "rpc"
	defaultMaxMissedHeartbeats = 3
)

type client struct {
	tr            thrift.TTransport
	stopHeartbeat chan struct{}
	heartbeat     string
}

func newInbox(prefix string) string {
	tokens := strings.Split(prefix, ".")
	tokens[len(tokens)-1] = nats.NewInbox() // Always at least 1 token
	inbox := ""
	pre := ""
	for _, token := range tokens {
		inbox += pre + token
		pre = "."
	}
	return inbox
}

// FNatsServer implements FServer by using NATS as the underlying transport.
// Clients must connect with the transport created by NewNatsServiceTTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. Use the "stateless" Nats server instead with
// FNatsServerBuilder.
type FNatsServer struct {
	conn                *nats.Conn
	subjects            []string
	heartbeatSubject    string
	heartbeatInterval   time.Duration
	maxMissedHeartbeats uint
	clients             map[string]*client
	mu                  sync.Mutex
	quit                chan struct{}
	processorFactory    FProcessorFactory
	transportFactory    FTransportFactory
	protocolFactory     *FProtocolFactory
	highWatermark       time.Duration
	waterMu             sync.RWMutex
}

// NewFNatsServer creates a new FNatsServer which listens for requests on the
// given subject. Clients must connect with the transport created by
// NewNatsServiceTTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServer(
	conn *nats.Conn,
	subject string,
	heartbeatInterval time.Duration,
	processor FProcessor,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return NewFNatsServerFactory(
		conn,
		subject,
		heartbeatInterval,
		defaultMaxMissedHeartbeats,
		NewFProcessorFactory(processor),
		transportFactory,
		protocolFactory,
	)
}

// NewFNatsServerWithSubjects creates a new FNatsServer which listens for
// requests on the given subjects. Clients must connect with the transport
// created by NewNatsServiceTTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServerWithSubjects(
	conn *nats.Conn,
	subjects []string,
	heartbeatInterval time.Duration,
	processor FProcessor,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return NewFNatsServerFactoryWithSubjects(
		conn,
		subjects,
		heartbeatInterval,
		defaultMaxMissedHeartbeats,
		NewFProcessorFactory(processor),
		transportFactory,
		protocolFactory,
	)
}

// NewFNatsServerFactory creates a new FNatsServer which listens for requests
// on the given subject. Clients must connect with the transport created by
// NewNatsServiceTTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServerFactory(
	conn *nats.Conn,
	subject string,
	heartbeatInterval time.Duration,
	maxMissedHeartbeats uint,
	processorFactory FProcessorFactory,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return NewFNatsServerFactoryWithSubjects(
		conn,
		[]string{subject},
		heartbeatInterval,
		maxMissedHeartbeats,
		processorFactory,
		transportFactory,
		protocolFactory,
	)
}

// NewFNatsServerFactoryWithSubjects creates a new FNatsServer which listens
// for requests on the given subjects. Clients must connect with the transport
// created by NewNatsServiceTTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServerFactoryWithSubjects(
	conn *nats.Conn,
	subjects []string,
	heartbeatInterval time.Duration,
	maxMissedHeartbeats uint,
	processorFactory FProcessorFactory,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return &FNatsServer{
		conn:                conn,
		subjects:            subjects,
		heartbeatSubject:    nats.NewInbox(),
		heartbeatInterval:   heartbeatInterval,
		maxMissedHeartbeats: maxMissedHeartbeats,
		clients:             make(map[string]*client),
		processorFactory:    processorFactory,
		transportFactory:    transportFactory,
		protocolFactory:     protocolFactory,
		quit:                make(chan struct{}, 1),
		highWatermark:       defaultWatermark,
	}
}

// Serve starts the server.
func (n *FNatsServer) Serve() error {
	subscriptions := make([]*nats.Subscription, len(n.subjects))
	for i, subject := range n.subjects {
		sub, err := n.conn.QueueSubscribe(subject, queue, n.handleConnection)
		if err != nil {
			return err
		}
		subscriptions[i] = sub
	}

	n.conn.Flush()
	if n.isHeartbeating() {
		go n.startHeartbeat()
	}

	log.Info("frugal: server running...")
	<-n.quit
	log.Info("frugal: server stopping...")

	for _, sub := range subscriptions {
		sub.Unsubscribe()
	}

	return nil
}

// Stop the server.
func (n *FNatsServer) Stop() error {
	close(n.quit)
	return nil
}

// SetHighWatermark sets the maximum amount of time a frame is allowed to await
// processing before triggering server overload logic. For now, this just
// consists of logging a warning. If not set, default is 5 seconds.
func (n *FNatsServer) SetHighWatermark(watermark time.Duration) {
	n.waterMu.Lock()
	n.highWatermark = watermark
	n.waterMu.Unlock()
}

// handleConnection is invoked when a remote peer is attempting to connect to
// the server.
func (n *FNatsServer) handleConnection(msg *nats.Msg) {
	if msg.Reply == "" {
		log.Warnf("frugal: discarding invalid connect message %+v", msg)
		return
	}
	hs := &natsConnectionHandshake{}
	if err := json.Unmarshal(msg.Data, hs); err != nil {
		log.Errorf("frugal: could not deserialize connect message %+v", msg)
		return
	}
	if hs.Version != natsV0 {
		log.Errorf("frugal: not a supported connect version %d", hs.Version)
		return
	}
	var (
		heartbeatReply = nats.NewInbox()
		listenTo       = newInbox(msg.Reply)
		tr, err        = n.accept(listenTo, msg.Reply, heartbeatReply)
	)
	if err != nil {
		log.Errorf("frugal: error accepting client transport: %s", err)
		return
	}

	client := &client{tr: tr, stopHeartbeat: make(chan struct{}), heartbeat: heartbeatReply}
	if n.isHeartbeating() {
		n.mu.Lock()
		n.clients[heartbeatReply] = client
		n.mu.Unlock()
	}

	// Connect message consists of "[heartbeat subject] [heartbeat reply subject] [expected interval ms]"
	connectMsg := n.heartbeatSubject + " " + heartbeatReply + " " +
		strconv.FormatInt(int64(n.heartbeatInterval/time.Millisecond), 10)
	if err := n.conn.PublishRequest(msg.Reply, listenTo, []byte(connectMsg)); err != nil {
		log.Errorf("frugal: error publishing transport inbox: %s", err)
		tr.Close()
	} else if n.isHeartbeating() {
		go n.acceptHeartbeat(client)
	}
}

func (n *FNatsServer) remove(heartbeat string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	client, ok := n.clients[heartbeat]
	if !ok {
		return
	}
	client.tr.Close()
	close(client.stopHeartbeat)
	delete(n.clients, heartbeat)
}

func (n *FNatsServer) startHeartbeat() {
	hbTicker := time.NewTicker(n.heartbeatInterval)
	for {
		select {
		case <-hbTicker.C:
			n.mu.Lock()
			clients := len(n.clients)
			n.mu.Unlock()
			if clients == 0 {
				continue
			}
			if err := n.conn.Publish(n.heartbeatSubject, nil); err != nil {
				log.Error("frugal: error publishing heartbeat:", err.Error())
			}
			if err := n.conn.FlushTimeout(n.heartbeatInterval * 3 / 4); err != nil {
				log.Error("frugal: error flushing heartbeat:", err.Error())
			}
		case <-n.quit:
			return
		}
	}
}

func (n *FNatsServer) acceptHeartbeat(client *client) {
	missed := uint(0)
	recvHeartbeat := make(chan struct{}, 1)

	sub, err := n.conn.Subscribe(client.heartbeat, func(msg *nats.Msg) {
		select {
		case recvHeartbeat <- struct{}{}:
		default:
			log.Info("frugal: FNatsServer dropped heartbeat:", client.heartbeat)
		}
	})
	if err != nil {
		log.Error("frugal: error subscribing to heartbeat:", client.heartbeat)
		return
	}
	defer sub.Unsubscribe()

	var wait <-chan time.Time
	for {
		if n.maxMissedHeartbeats > 1 {
			wait = time.After(n.heartbeatInterval)
		} else {
			wait = time.After(n.heartbeatInterval + n.heartbeatInterval/4)
		}
		select {
		case <-wait:
			missed++
			if missed >= n.maxMissedHeartbeats {
				log.Warnf("frugal: client heartbeat expired for heartbeat: %s", client.heartbeat)
				n.remove(client.heartbeat)
				return
			}
		case <-recvHeartbeat:
			missed = 0
		case <-client.stopHeartbeat:
			return
		case <-n.quit:
			return
		}
	}
}

func (n *FNatsServer) accept(listenTo, replyTo, heartbeat string) (FTransport, error) {
	client := newNatsServiceTTransportServer(n.conn, listenTo, replyTo)
	transport := n.transportFactory.GetTransport(client)
	processor := n.processorFactory.GetProcessor(transport)
	protocol := n.protocolFactory.GetProtocol(transport)
	transport.SetRegistry(NewServerRegistry(processor, n.protocolFactory, protocol))
	n.waterMu.RLock()
	transport.SetHighWatermark(n.highWatermark)
	n.waterMu.RUnlock()
	if err := transport.Open(); err != nil {
		return nil, err
	}

	// Cleanup heartbeat when client disconnects.
	go func() {
		select {
		case <-n.quit:
			client.Close()
		case <-transport.Closed():
		}

		n.remove(heartbeat)
	}()

	log.Debug("frugal: client connection accepted with heartbeat:", heartbeat)
	return transport, nil
}

func (n *FNatsServer) isHeartbeating() bool {
	return n.heartbeatInterval > 0
}
