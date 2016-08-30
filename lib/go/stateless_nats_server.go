package frugal

import (
	"bytes"
	"encoding/binary"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

const defaultWorkQueueLen = 64
const defaultWatermark = 5 * time.Second

type frameWrapper struct {
	frameBytes []byte
	timestamp  time.Time
	reply      string
}

// FNatsServerBuilder configures and builds "stateless" nats servers instances.
type FNatsServerBuilder struct {
	conn               *nats.Conn
	processor          FProcessor
	inputProtoFactory  *FProtocolFactory
	outputProtoFactory *FProtocolFactory
	subject            string
	queue              string
	workerCount        uint
	queueLen           uint
	highWatermark      time.Duration
}

// NewFStatelessNatsServerBuilder creates a builder which configures and builds
// "stateless" nats servers instances.
func NewFNatsServerBuilder(conn *nats.Conn, processor FProcessor,
	protoFactory *FProtocolFactory, subject string) *FNatsServerBuilder {
	return &FNatsServerBuilder{
		conn:               conn,
		processor:          processor,
		inputProtoFactory:  protoFactory,
		outputProtoFactory: protoFactory,
		subject:            subject,
		workerCount:        1,
		queueLen:           defaultWorkQueueLen,
		highWatermark:      defaultWatermark,
	}
}

// WithQueueGroup adds a NATS queue group to receive requests on.
func (f *FNatsServerBuilder) WithQueueGroup(queue string) *FNatsServerBuilder {
	f.queue = queue
	return f
}

// WithWorkerCount controls the number of goroutines used to process requests.
func (f *FNatsServerBuilder) WithWorkerCount(workerCount uint) *FNatsServerBuilder {
	f.workerCount = workerCount
	return f
}

// WithQueueLength controls the length of the work queue used to buffer
// requests.
func (f *FNatsServerBuilder) WithQueueLength(queueLength uint) *FNatsServerBuilder {
	f.queueLen = queueLength
	return f
}

// WithHighWatermark controls the time duration requests wait in queue before
// triggering slow consumer logic.
func (f *FNatsServerBuilder) WithHighWatermark(highWatermark time.Duration) *FNatsServerBuilder {
	f.highWatermark = highWatermark
	return f
}

// Build a new configured NATS FServer.
func (f *FNatsServerBuilder) Build() FServer {
	return &fNatsServer{
		conn:               f.conn,
		processor:          f.processor,
		inputProtoFactory:  f.inputProtoFactory,
		outputProtoFactory: f.outputProtoFactory,
		subject:            f.subject,
		queue:              f.queue,
		workerCount:        f.workerCount,
		workC:              make(chan *frameWrapper, f.queueLen),
		quit:               make(chan struct{}),
		highWatermark:      f.highWatermark,
	}
}

// fNatsServer implements FServer by using NATS as the underlying transport.
// Clients must connect with the transport created by NewNatsFTransport.
type fNatsServer struct {
	conn               *nats.Conn
	processor          FProcessor
	inputProtoFactory  *FProtocolFactory
	outputProtoFactory *FProtocolFactory
	subject            string
	queue              string
	workerCount        uint
	workC              chan *frameWrapper
	quit               chan struct{}
	highWatermark      time.Duration
}

// Serve starts the server.
func (f *fNatsServer) Serve() error {
	sub, err := f.conn.QueueSubscribe(f.subject, f.queue, f.handler)
	if err != nil {
		return err
	}

	for i := uint(0); i < f.workerCount; i++ {
		go f.worker()
	}

	logger().Info("frugal: server running...")
	<-f.quit
	logger().Info("frugal: server stopping...")

	sub.Unsubscribe()

	return nil
}

// Stop the server.
func (f *fNatsServer) Stop() error {
	close(f.quit)
	return nil
}

// handler is invoked when a request is received. The request is placed on the
// work channel which is processed by a worker goroutine.
func (f *fNatsServer) handler(msg *nats.Msg) {
	if msg.Reply == "" {
		logger().Warn("frugal: discarding invalid NATS request (no reply)")
		return
	}
	select {
	case f.workC <- &frameWrapper{frameBytes: msg.Data, timestamp: time.Now(), reply: msg.Reply}:
	case <-f.quit:
		return
	}
}

// worker should be called as a goroutine. It reads requests off the work
// channel and processes them.
func (f *fNatsServer) worker() {
	for {
		select {
		case <-f.quit:
			return
		case frame := <-f.workC:
			dur := time.Since(frame.timestamp)
			if dur > f.highWatermark {
				logger().Warnf("frugal: frame spent %+v in the transport buffer, your consumer might be backed up", dur)
			}
			if err := f.processFrame(frame.frameBytes, frame.reply); err != nil {
				logger().Errorf("frugal: error processing frame: %s", err.Error())
			}
		}
	}
}

// processFrame invokes the FProcessor and sends the response on the given
// subject.
func (f *fNatsServer) processFrame(frame []byte, reply string) error {
	// Read and process frame.
	input := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(frame[4:])} // Discard frame size
	// Buffer 1MB - 4 bytes since frame size is copied directly.
	output := NewFBoundedMemoryBuffer(natsMaxMessageSize - 4)
	if err := f.processor.Process(
		f.inputProtoFactory.GetProtocol(input),
		f.outputProtoFactory.GetProtocol(output)); err != nil {
		return err
	}

	if output.Len() == 0 {
		return nil
	}

	// Add frame size (4-byte uint32).
	response := make([]byte, output.Len()+4)
	binary.BigEndian.PutUint32(response, uint32(output.Len()))
	copy(response[4:], output.Bytes())

	// Send response.
	return f.conn.Publish(reply, response)
}
