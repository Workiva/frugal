package frugal

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

const (
	queue                      = "rpc"
	defaultMaxMissedHeartbeats = 3
	minHeartbeatInterval       = 20 * time.Second
	frugalPrefix               = "frugal."
)

type client struct {
	tr            thrift.TTransport
	stopHeartbeat chan struct{}
	heartbeat     string
}

func newFrugalInbox() string {
	return fmt.Sprintf("%s%s", frugalPrefix, nats.NewInbox())
}

// FNatsServer implements FServer by using NATS as the underlying transport.
type FNatsServer struct {
	conn                *nats.Conn
	subject             string
	heartbeatSubject    string
	heartbeatInterval   time.Duration
	maxMissedHeartbeats uint
	clients             map[string]*client
	mu                  sync.Mutex
	quit                chan struct{}
	processorFactory    FProcessorFactory
	transportFactory    FTransportFactory
	protocolFactory     *FProtocolFactory
}

// NewFNatsServer creates a new FNatsServer which listens for requests on the
// given subject.
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

func NewFNatsServerFactory(
	conn *nats.Conn,
	subject string,
	heartbeatInterval time.Duration,
	maxMissedHeartbeats uint,
	processorFactory FProcessorFactory,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	if heartbeatInterval < minHeartbeatInterval {
		heartbeatInterval = minHeartbeatInterval
	}

	return &FNatsServer{
		conn:                conn,
		subject:             subject,
		heartbeatSubject:    nats.NewInbox(),
		heartbeatInterval:   heartbeatInterval,
		maxMissedHeartbeats: maxMissedHeartbeats,
		clients:             make(map[string]*client),
		processorFactory:    processorFactory,
		transportFactory:    transportFactory,
		protocolFactory:     protocolFactory,
		quit:                make(chan struct{}, 1),
	}
}

// Serve starts the server.
func (n *FNatsServer) Serve() error {
	sub, err := n.conn.QueueSubscribe(n.subject, queue, func(msg *nats.Msg) {
		if msg.Reply == "" {
			log.Printf("frugal: discarding invalid connect message %+v\n", msg)
			return
		}
		hs := &natsConnectionHandshake{}
		if err := json.Unmarshal(msg.Data, hs); err != nil {
			log.Printf("frugal: could not deserialize connect message %+v\n", msg)
			return
		}
		if hs.Version != natsV0 {
			log.Printf("frugal: not a supported connect version %d", hs.Version)
			return
		}
		var (
			heartbeatReply = nats.NewInbox()
			listenTo       = newFrugalInbox()
			tr, err        = n.accept(listenTo, msg.Reply, heartbeatReply)
		)
		if err != nil {
			log.Println("frugal: error accepting client transport:", err)
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
			strconv.FormatInt(int64(n.heartbeatInterval.Seconds())*1000, 10)
		if err := n.conn.PublishRequest(msg.Reply, listenTo, []byte(connectMsg)); err != nil {
			log.Println("frugal: error publishing transport inbox:", err)
			tr.Close()
		} else if n.isHeartbeating() {
			go n.acceptHeartbeat(client)
		}
	})
	if err != nil {
		return err
	}

	n.conn.Flush()
	if n.isHeartbeating() {
		go n.startHeartbeat()
	}

	log.Println("frugal: server running...")
	<-n.quit
	if n.conn.Status() != nats.CONNECTED {
		log.Println("frugal: Nats is already disconnected!")
		return nil
	}
	return sub.Unsubscribe()
}

// Stop the server.
func (n *FNatsServer) Stop() error {
	close(n.quit)
	return nil
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
	for {
		select {
		case <-time.After(n.heartbeatInterval):
			n.mu.Lock()
			clients := len(n.clients)
			n.mu.Unlock()
			if clients == 0 {
				continue
			}
			if err := n.conn.Publish(n.heartbeatSubject, nil); err != nil {
				log.Println("frugal: error publishing heartbeat:", err.Error())
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
		}
	})
	if err != nil {
		log.Println("frugal: error subscribing to heartbeat", client.heartbeat)
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-time.After(n.heartbeatInterval + heartbeatGracePeriod):
			missed++
			if missed >= n.maxMissedHeartbeats {
				log.Println("frugal: client heartbeat expired")
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
	if err := transport.Open(); err != nil {
		return nil, err
	}

	processor := n.processorFactory.GetProcessor(transport)
	protocol := n.protocolFactory.GetProtocol(transport)
	transport.SetRegistry(NewServerRegistry(processor, n.protocolFactory, protocol))

	// Cleanup heartbeat when client disconnects.
	go func() {
		select {
		case <-n.quit:
			client.Close()
		case <-transport.Closed():
		}

		n.remove(heartbeat)
	}()

	return transport, nil
}

func (n *FNatsServer) isHeartbeating() bool {
	return n.heartbeatInterval > 0
}
