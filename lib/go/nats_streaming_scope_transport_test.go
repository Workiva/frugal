package frugal

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/go-nats-streaming/pb"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStanConn struct {
	mock.Mock
}

func (m *mockStanConn) Publish(subject string, data []byte) error {
	return m.Called(subject, data).Error(0)
}

func (m *mockStanConn) PublishAsync(subject string, data []byte, ah stan.AckHandler) (string, error) {
	called := m.Called(subject, data, ah)
	return called.String(0), called.Error(1)
}

func (m *mockStanConn) PublishWithReply(subject, reply string, data []byte) error {
	return m.Called(subject, reply, data).Error(0)
}

func (m *mockStanConn) PublishAsyncWithReply(subject, reply string, data []byte, ah stan.AckHandler) (string, error) {
	called := m.Called(subject, reply, data, ah)
	return called.String(0), called.Error(1)
}

func (m *mockStanConn) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	called := m.Called(subject, cb, opts)
	return called.Get(0).(stan.Subscription), called.Error(1)
}

func (m *mockStanConn) QueueSubscribe(subject, queue string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	called := m.Called(subject, queue, cb, opts)
	return called.Get(0).(stan.Subscription), called.Error(1)
}

func (m *mockStanConn) Close() error {
	return m.Called().Error(0)
}

func (m *mockStanConn) NatsConn() *nats.Conn {
	return m.Called().Get(0).(*nats.Conn)
}

type mockStanSubscription struct {
	mock.Mock
}

func (m *mockStanSubscription) Unsubscribe() error {
	return m.Called().Error(0)
}

// Ensures LockTopic returns an error if the transport is configured as a
// subscriber.
func TestNatsStreamingScopeLockTopicSubscriberError(t *testing.T) {
	conn := new(mockStanConn)
	sub := new(mockStanSubscription)
	tr := NewFNatsStreamingScopeTransport(conn)
	conn.On(
		"QueueSubscribe",
		"frugal.foo",
		"",
		mock.AnythingOfType("stan.MsgHandler"),
		[]stan.SubscriptionOption(nil),
	).Return(sub, nil)
	tr.Subscribe("foo")

	assert.Error(t, tr.LockTopic("blah"))
	conn.AssertExpectations(t)
}

// Ensures LockTopic returns nil when a publisher successfully locks a topic.
// Subsequent calls will wait on the mutex. Unlock releases the topic.
func TestNatsStreamingScopeLockUnlockTopic(t *testing.T) {
	tr := NewFNatsStreamingScopeTransport(nil)
	assert.Nil(t, tr.LockTopic("foo"))
	acquired := make(chan bool)
	go func() {
		assert.Nil(t, tr.LockTopic("bar"))
		assert.Equal(t, "bar", tr.(*fNatsStreamingScopeTransport).topic)
		acquired <- true
	}()
	assert.Equal(t, "foo", tr.(*fNatsStreamingScopeTransport).topic)
	tr.UnlockTopic()
	<-acquired

	tr.UnlockTopic()
	assert.Equal(t, "", tr.(*fNatsStreamingScopeTransport).topic)
}

// Ensures UnlockTopic returns an error if the transport is a subscriber.
func TestNatsStreamingScopeUnlockTopicSubscriberError(t *testing.T) {
	conn := new(mockStanConn)
	sub := new(mockStanSubscription)
	tr := NewFNatsStreamingScopeTransport(conn)
	conn.On(
		"QueueSubscribe",
		"frugal.foo",
		"",
		mock.AnythingOfType("stan.MsgHandler"),
		[]stan.SubscriptionOption(nil),
	).Return(sub, nil)

	tr.Subscribe("foo")

	assert.Error(t, tr.UnlockTopic())
	conn.AssertExpectations(t)
}

// Ensures Subscribe subscribes to the topic on NATS and puts received frames
// on the read buffer and received in Read calls. Ensure STAN headers are
// injected.
func TestNatsStreamingScopeSubscribeRead(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	sub := new(mockStanSubscription)
	tr := NewFNatsStreamingScopeTransport(conn)
	conn.On(
		"QueueSubscribe",
		"frugal.foo",
		"",
		mock.AnythingOfType("stan.MsgHandler"),
		[]stan.SubscriptionOption(nil),
	).Return(sub, nil)

	assert.Nil(tr.Subscribe("foo"))

	timestamp := int64(1467236243)
	msg := &stan.Msg{
		pb.MsgProto{
			Data:      completeFrugalFrame,
			Sequence:  42,
			Timestamp: timestamp,
		},
		nil,
	}
	tr.(*fNatsStreamingScopeTransport).handleMessage(msg)

	injectedSize := 41
	frameBuff := make([]byte, len(completeFrugalFrame)+injectedSize)
	buff := make([]byte, 10)
	offset := 0
	for i := 0; i < int(math.Ceil(float64(len(frameBuff))/10)); i++ {
		n, err := tr.Read(buff)
		assert.Nil(err)
		assert.Equal(10, n)
		offset += copy(frameBuff[offset:], buff)
	}
	frameBuffWithSize := make([]byte, len(frameBuff)+4)
	binary.BigEndian.PutUint32(frameBuffWithSize, uint32(len(frameBuff)))
	copy(frameBuffWithSize[4:], frameBuff)
	components, err := unmarshalFrame(frameBuffWithSize)
	assert.Nil(err)
	assert.Equal([]byte("this is a request"), components.payload)
	assert.Equal("bar", components.headers["foo"])
	assert.Equal("qux", components.headers["baz"])
	assert.Equal("0", components.headers["_opid"])
	assert.Equal("12345", components.headers["_cid"])
	assert.Equal(strconv.FormatUint(uint64(timestamp), 10), components.headers[NatsTimestampHeader])
	assert.Equal("42", components.headers[NatsSequenceHeader])
	conn.AssertExpectations(t)
}

// Ensures Subscribe subscribes to the topic and queue on NATS when a queue is
// specified.
func TestNatsStreamingScopeSubscribeQueue(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	sub := new(mockStanSubscription)
	queue := "queue"
	tr := NewFNatsStreamingScopeTransportWithQueue(conn, queue)
	conn.On(
		"QueueSubscribe",
		"frugal.foo",
		queue,
		mock.AnythingOfType("stan.MsgHandler"),
		[]stan.SubscriptionOption(nil),
	).Return(sub, nil)

	assert.Nil(tr.Subscribe("foo"))
	conn.AssertExpectations(t)
}

// Ensures Read returns an EOF if the transport is not open.
func TestNatsStreamingScopeReadNotOpen(t *testing.T) {
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)

	n, err := tr.Read(make([]byte, 5))
	assert.Equal(t, 0, n)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.END_OF_FILE, trErr.TypeId())
}

// Ensures Open returns nil on success and writes work.
func TestNatsStreamingScopeOpenPublisherWriteFlush(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)
	frame := make([]byte, 10)
	conn.On(
		"Publish",
		"frugal.foo",
		append([]byte{0, 0, 0, byte(len(frame))}, frame...), // Prepend frame size
	).Return(nil)

	assert.Nil(tr.Open())
	assert.True(tr.IsOpen())
	assert.Nil(tr.LockTopic("foo"))
	n, err := tr.Write(frame)
	assert.Nil(err)
	assert.Equal(10, n)
	assert.Nil(tr.Flush())
	conn.AssertExpectations(t)
}

// Ensures Open returns an ALREADY_OPEN TTransportException if the transport is
// already open.
func TestNatsStreamingScopeOpenAlreadyOpen(t *testing.T) {
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)

	assert.Nil(t, tr.Open())

	err := tr.Open()

	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.ALREADY_OPEN, trErr.TypeId())
}

// Ensures Open returns an error for subscribers with no subject set.
func TestNatsStreamingScopeOpenSubscriberNoSubject(t *testing.T) {
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)
	tr.(*fNatsStreamingScopeTransport).subscriber = true

	assert.Error(t, tr.Open())
}

// Ensures subscribers discard invalid frames (size < 4).
func TestNatsStreamingScopeDiscardInvalidFrame(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	sub := new(mockStanSubscription)
	tr := NewFNatsStreamingScopeTransport(conn)
	conn.On(
		"QueueSubscribe",
		"frugal.blah",
		"",
		mock.AnythingOfType("stan.MsgHandler"),
		[]stan.SubscriptionOption(nil),
	).Return(sub, nil)
	sub.On("Unsubscribe").Return(nil)
	assert.Nil(tr.Subscribe("blah"))

	msg := &stan.Msg{pb.MsgProto{Data: make([]byte, 2)}, nil}
	tr.(*fNatsStreamingScopeTransport).handleMessage(msg)
	assert.Nil(tr.Close())

	buff := make([]byte, 3)
	n, err := tr.Read(buff)
	assert.Equal(0, n)
	trErr := err.(thrift.TTransportException)
	assert.Equal(thrift.END_OF_FILE, trErr.TypeId())
	conn.AssertExpectations(t)
	sub.AssertExpectations(t)
}

// Ensures Close returns nil if the transport is not open.
func TestNatsStreamingScopeCloseNotOpen(t *testing.T) {
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)
	assert.Nil(t, tr.Close())
	assert.False(t, tr.IsOpen())
}

// Ensures Close closes the publisher transport and returns nil.
func TestNatsStreamingScopeClosePublisher(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)
	assert.Nil(tr.LockTopic("foo"))
	assert.Nil(tr.Open())
	assert.True(tr.IsOpen())
	assert.Nil(tr.Close())
	assert.False(tr.IsOpen())
}

// Ensures Close returns an error if the unsubscribe fails.
func TestNatsStreamingScopeCloseSubscriberError(t *testing.T) {
	conn := new(mockStanConn)
	sub := new(mockStanSubscription)
	tr := NewFNatsStreamingScopeTransport(conn)
	conn.On(
		"QueueSubscribe",
		"frugal.foo",
		"",
		mock.AnythingOfType("stan.MsgHandler"),
		[]stan.SubscriptionOption(nil),
	).Return(sub, nil)
	sub.On("Unsubscribe").Return(errors.New("error"))
	assert.Nil(t, tr.Subscribe("foo"))
	assert.Error(t, tr.Close())
	conn.AssertExpectations(t)
	sub.AssertExpectations(t)
}

// Ensures Write returns an error if the transport is not open.
func TestNatsStreamingScopeWriteNotOpen(t *testing.T) {
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)

	n, err := tr.Write(make([]byte, 10))
	assert.Equal(t, 0, n)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Write returns an ErrTooLarge if the written frame exceeds 1MB.
func TestNatsStreamingScopeWriteTooLarge(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)
	assert.Nil(tr.Open())

	n, err := tr.Write(make([]byte, 5))
	assert.Equal(5, n)
	assert.Nil(err)
	n, err = tr.Write(make([]byte, 1024*1024+10))
	assert.Equal(0, n)
	assert.Equal(ErrTooLarge, err)
	assert.Equal(0, tr.(*fNatsStreamingScopeTransport).writeBuffer.Len())
}

// Ensures Flush returns an error if the transport is not open.
func TestNatsStreamingScopeFlushNotOpen(t *testing.T) {
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)

	err := tr.Flush()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Flush returns nil and nothing is sent to NATS when there is no data
// to flush.
func TestNatsStreamingScopeFlushNoData(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)
	assert.Nil(tr.Open())
	assert.Nil(tr.LockTopic("foo"))
	assert.Nil(tr.Flush())
}

// Ensures Flush returns nil and publishes to NATS when data is buffered.
func TestNatsStreamingScopeFlush(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)
	data := make([]byte, 10)
	conn.On(
		"Publish",
		"frugal.foo",
		append([]byte{0, 0, 0, byte(len(data))}, data...), // Prepend frame size
	).Return(nil)
	assert.Nil(tr.Open())
	assert.Nil(tr.LockTopic("foo"))
	_, err := tr.Write(data)
	assert.Nil(err)
	assert.Nil(tr.Flush())
	conn.AssertExpectations(t)
}

// Ensures Flush returns a TTransportException if the publish to NATS fails.
func TestNatsStreamingScopeFlushError(t *testing.T) {
	assert := assert.New(t)
	conn := new(mockStanConn)
	tr := NewFNatsStreamingScopeTransport(conn)
	data := make([]byte, 10)
	err := errors.New("error")
	conn.On(
		"Publish",
		"frugal.foo",
		append([]byte{0, 0, 0, byte(len(data))}, data...), // Prepend frame size
	).Return(err)
	assert.Nil(tr.Open())
	assert.Nil(tr.LockTopic("foo"))
	_, err = tr.Write(data)
	assert.Nil(err)
	err = tr.Flush()
	assert.Error(err)
	trErr := err.(thrift.TTransportException)
	assert.Equal("error", trErr.Error())
	conn.AssertExpectations(t)
}

// Ensures RemainingBytes returns max uint64.
func TestNatsStreamingScopeRemainingBytes(t *testing.T) {
	tr := NewFNatsStreamingScopeTransport(nil)
	assert.Equal(t, ^uint64(0), tr.RemainingBytes())
}
