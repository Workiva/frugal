package frugal

import (
	"errors"
	"io"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var frame = []byte{
	0, 0, 0, 70, 0, 0, 0, 0, 65, 0, 0, 0, 5, 104, 101, 108, 108, 111, 0, 0, 0,
	5, 119, 111, 114, 108, 100, 0, 0, 0, 5, 95, 111, 112, 105, 100, 0, 0, 0, 1,
	48, 0, 0, 0, 4, 95, 99, 105, 100, 0, 0, 0, 21, 105, 89, 65, 71, 67, 74, 72,
	66, 87, 67, 75, 76, 74, 66, 115, 106, 107, 100, 111, 104, 98,
}

type mockTTransport struct {
	mock.Mock
	reads     chan []byte
	readError error
}

func (m *mockTTransport) Open() error {
	return m.Called().Error(0)
}

func (m *mockTTransport) Close() error {
	return m.Called().Error(0)
}

func (m *mockTTransport) IsOpen() bool {
	return m.Called().Bool(0)
}

func (m *mockTTransport) Read(b []byte) (int, error) {
	if m.readError != nil {
		return 0, m.readError
	}
	read, ok := <-m.reads
	if !ok {
		return 0, thrift.NewTTransportExceptionFromError(io.EOF)
	}
	copy(b, read)
	num := len(b)
	if len(read) < num {
		num = len(read)
	}
	return num, nil
}

func (m *mockTTransport) Write(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockTTransport) Flush() error {
	return m.Called().Error(0)
}

func (m *mockTTransport) RemainingBytes() uint64 {
	return m.Called().Get(0).(uint64)
}

type mockTTransportFactory struct {
	mock.Mock
}

func (m *mockTTransportFactory) GetTransport(tr thrift.TTransport) thrift.TTransport {
	return m.Called(tr).Get(0).(thrift.TTransport)
}

// Ensures NewTFramedTransportFactory creates a tFramedTransportFactory with
// the default max length and GetTransport calls the underlying
// TTransportFactory.
func TestTFramedTransportFactory(t *testing.T) {
	mockTrFactory := new(mockTTransportFactory)
	trFactory := NewTFramedTransportFactory(mockTrFactory)
	mockTr := new(mockTTransport)
	mockTrFactory.On("GetTransport", mockTr).Return(mockTr)

	tr := trFactory.GetTransport(mockTr)

	assert.Equal(t, mockTr, tr.(*TFramedTransport).transport)
	assert.Equal(t, uint32(defaultMaxLength), tr.(*TFramedTransport).maxLength)
	mockTrFactory.AssertExpectations(t)
}

// Ensures NewTFramedTransportFactory creates a tFramedTransportFactory with
// the specified max length and GetTransport calls the underlying
// TTransportFactory.
func TestTFramedTransportFactoryMaxLength(t *testing.T) {
	mockTrFactory := new(mockTTransportFactory)
	maxLength := uint32(1024)
	trFactory := NewTFramedTransportFactoryMaxLength(mockTrFactory, maxLength)
	mockTr := new(mockTTransport)
	mockTrFactory.On("GetTransport", mockTr).Return(mockTr)

	tr := trFactory.GetTransport(mockTr)

	assert.Equal(t, mockTr, tr.(*TFramedTransport).transport)
	assert.Equal(t, maxLength, tr.(*TFramedTransport).maxLength)
	mockTrFactory.AssertExpectations(t)
}

// Ensures Open calls through to the underlying transport.
func TestOpen(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewTFramedTransport(mockTr)
	mockTr.On("Open").Return(nil)

	assert.Nil(t, tr.Open())
	assert.Equal(t, uint32(defaultMaxLength), tr.maxLength)
	mockTr.AssertExpectations(t)
}

// Ensures Open calls through to the underlying transport and returns the error
// returned by that transport.
func TestOpenError(t *testing.T) {
	mockTr := new(mockTTransport)
	maxLength := uint32(1024)
	tr := NewTFramedTransportMaxLength(mockTr, maxLength)
	err := errors.New("error")
	mockTr.On("Open").Return(err)

	assert.Equal(t, err, tr.Open())
	assert.Equal(t, maxLength, tr.maxLength)
	mockTr.AssertExpectations(t)
}

// Ensures Close calls through to the underlying transport.
func TestClose(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewTFramedTransport(mockTr)
	mockTr.On("Close").Return(nil)

	assert.Nil(t, tr.Close())
	mockTr.AssertExpectations(t)
}

// Ensures Close calls through to the underlying transport and returns the
// error returned by that transport.
func TestCloseError(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewTFramedTransport(mockTr)
	err := errors.New("error")
	mockTr.On("Close").Return(err)

	assert.Equal(t, err, tr.Close())
	mockTr.AssertExpectations(t)
}

// Ensures IsOpen calls through to the underlying transport.
func TestIsOpen(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewTFramedTransport(mockTr)
	mockTr.On("IsOpen").Return(true)

	assert.True(t, tr.IsOpen())
	mockTr.AssertExpectations(t)
}

// Ensures Read reads a complete frame into the provided buffer from the
// underlying transport.
func TestRead(t *testing.T) {
	mockTr := new(mockTTransport)
	reads := make(chan []byte, 2)
	reads <- frame[0:4]
	reads <- frame[4:]
	close(reads)
	mockTr.reads = reads
	tr := NewTFramedTransport(mockTr)
	mockTr.On("Read", make([]byte, 4096)).Return(4, nil).Once()
	mockTr.On("Read", append(frame[0:4], make([]byte, 4092)...)).Return(len(frame), nil).Once()

	buff := make([]byte, len(frame)-4)
	n, err := tr.Read(buff)

	assert.Nil(t, err)
	assert.Equal(t, len(frame)-4, n)
	assert.Equal(t, frame[4:], buff)
}

// Ensure Read returns an error if the provided buffer is larger than the
// frame.
func TestReadLargeBuffer(t *testing.T) {
	mockTr := new(mockTTransport)
	reads := make(chan []byte, 2)
	reads <- frame[0:4]
	reads <- frame[4:]
	close(reads)
	mockTr.reads = reads
	tr := NewTFramedTransport(mockTr)

	buff := make([]byte, len(frame)+10)
	n, err := tr.Read(buff)

	assert.Error(t, err)
	assert.Equal(t, "frugal: not enough frame (size 70) to read 84 bytes", err.Error())
	assert.Equal(t, len(frame)-4, n)
	assert.Equal(t, append(frame[4:], make([]byte, 14)...), buff)
}

// Ensures Read returns an error if it fails to read the frame size header.
func TestReadHeaderError(t *testing.T) {
	mockTr := new(mockTTransport)
	mockTr.readError = errors.New("error")
	tr := NewTFramedTransport(mockTr)

	buff := make([]byte, len(frame)-4)
	n, err := tr.Read(buff)

	assert.Equal(t, mockTr.readError, err)
	assert.Equal(t, 0, n)
}

// Ensures Write writes the bytes to the transport buffer.
func TestWrite(t *testing.T) {
	tr := NewTFramedTransport(new(mockTTransport))
	buff := make([]byte, 10)

	n, err := tr.Write(buff)

	assert.Equal(t, 10, n)
	assert.Nil(t, err)
	assert.Equal(t, buff, tr.buf.Bytes())
}

// Ensures Flush copies the framed buffered bytes to the underlying transport
// and then flushes it.
func TestFlush(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewTFramedTransport(mockTr)
	buff := make([]byte, 10)
	_, err := tr.Write(buff)
	assert.Nil(t, err)
	mockTr.On("Write", []byte{0, 0, 0, 10}).Return(4, nil)
	mockTr.On("Write", buff).Return(len(buff), nil)
	mockTr.On("Flush").Return(nil)

	assert.Nil(t, tr.Flush())
	mockTr.AssertExpectations(t)
}

// Ensures Flush returns an error if writing the frame size to the underlying
// transport fails.
func TestFlushFrameSizeError(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewTFramedTransport(mockTr)
	buff := make([]byte, 10)
	_, err := tr.Write(buff)
	assert.Nil(t, err)
	mockTr.On("Write", []byte{0, 0, 0, 10}).Return(0, errors.New("error"))

	assert.Error(t, tr.Flush())
	mockTr.AssertExpectations(t)
}

// Ensures Flush returns an error if copying the buffered bytes to the
// underlying transport fails.
func TestFlushWriteError(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewTFramedTransport(mockTr)
	buff := make([]byte, 10)
	_, err := tr.Write(buff)
	assert.Nil(t, err)
	mockTr.On("Write", []byte{0, 0, 0, 10}).Return(4, nil)
	mockTr.On("Write", buff).Return(0, errors.New("error"))

	assert.Error(t, tr.Flush())
	mockTr.AssertExpectations(t)
}

// Ensures Flush returns an error if flushing the underlying transport fails.
func TestFlushError(t *testing.T) {
	mockTr := new(mockTTransport)
	tr := NewTFramedTransport(mockTr)
	buff := make([]byte, 10)
	_, err := tr.Write(buff)
	assert.Nil(t, err)
	mockTr.On("Write", []byte{0, 0, 0, 10}).Return(4, nil)
	mockTr.On("Write", buff).Return(len(buff), nil)
	mockTr.On("Flush").Return(errors.New("error"))

	assert.Error(t, tr.Flush())
	mockTr.AssertExpectations(t)
}

// Ensures RemainingBytes returns the remaining frame size.
func TestRemainingBytes(t *testing.T) {
	mockTr := new(mockTTransport)
	reads := make(chan []byte, 2)
	reads <- frame[0:4]
	reads <- frame[4:]
	close(reads)
	mockTr.reads = reads
	tr := NewTFramedTransport(mockTr)
	mockTr.On("Read", make([]byte, 4096)).Return(4, nil).Once()
	mockTr.On("Read", append(frame[0:4], make([]byte, 4092)...)).Return(len(frame), nil).Once()

	assert.Equal(t, uint64(0), tr.RemainingBytes())

	buff := make([]byte, len(frame)-10)
	_, err := tr.Read(buff)
	assert.Nil(t, err)

	assert.Equal(t, uint64(6), tr.RemainingBytes())
}
