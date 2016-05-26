package frugal

import (
	"errors"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// _opid0_cid123[1,"ping",1,0,{}]
var pingFrame = []byte{
	0, 0, 0, 0, 29, 0, 0, 0, 5, 95, 111, 112, 105, 100, 0, 0, 0, 1, 48, 0, 0,
	0, 4, 95, 99, 105, 100, 0, 0, 0, 3, 49, 50, 51, 91, 49, 44, 34, 112, 105,
	110, 103, 34, 44, 49, 44, 48, 44, 123, 125, 93,
}

type pingProcessor struct {
	t             *testing.T
	expectedProto *FProtocol
	called        bool
	err           error
}

func (p *pingProcessor) Process(ctx *FContext, iprot, oprot *FProtocol) error {
	p.called = true
	assert.Equal(p.t, p.expectedProto, iprot)
	assert.Equal(p.t, p.expectedProto, oprot)
	assert.Equal(p.t, "123", ctx.CorrelationID())
	return p.err
}

// Ensures FBaseProcessor invokes the correct FProcessorFunction and returns
// nil on success.
func TestFBaseProcessorHappyPath(t *testing.T) {
	mockTransport := new(mockTTransport)
	reads := make(chan []byte, 3)
	reads <- pingFrame[0:5]  // version and headers size
	reads <- pingFrame[5:34] // FContext headers
	reads <- pingFrame[34:]  // request body
	mockTransport.reads = reads
	proto := &FProtocol{thrift.NewTJSONProtocol(mockTransport)}
	processor := NewFBaseProcessor()
	processorFunction := &pingProcessor{t: t, expectedProto: proto}
	processor.AddToProcessorMap("ping", processorFunction)

	assert.Nil(t, processor.Process(proto, proto))
	assert.True(t, processorFunction.called)
}

// Ensures FBaseProcessor invokes the correct FProcessorFunction and returns
// an error on failure.
func TestFBaseProcessorError(t *testing.T) {
	mockTransport := new(mockTTransport)
	reads := make(chan []byte, 3)
	reads <- pingFrame[0:5]  // version and headers size
	reads <- pingFrame[5:34] // FContext headers
	reads <- pingFrame[34:]  // request body
	mockTransport.reads = reads
	proto := &FProtocol{thrift.NewTJSONProtocol(mockTransport)}
	processor := NewFBaseProcessor()
	err := errors.New("error")
	processorFunction := &pingProcessor{t: t, expectedProto: proto, err: err}
	processor.AddToProcessorMap("ping", processorFunction)

	assert.Equal(t, err, processor.Process(proto, proto))
	assert.True(t, processorFunction.called)
}

// Ensures FBaseProcessor returns a TTransportException if the transport read
// fails.
func TestFBaseProcessorReadError(t *testing.T) {
	mockTransport := new(mockTTransport)
	err := errors.New("error")
	mockTransport.readError = err
	proto := &FProtocol{thrift.NewTJSONProtocol(mockTransport)}
	processor := NewFBaseProcessor()

	err = processor.Process(proto, proto)
	assert.Error(t, err)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, int(thrift.UNKNOWN_TRANSPORT_EXCEPTION), int(trErr.TypeId()))
}

// Ensures FBaseProcessor writes and returns an UNKNOWN_METHOD
// TApplicationException if there is no registered FProcessorFunction.
func TestFBaseProcessorNoProcessorFunction(t *testing.T) {
	mockTransport := new(mockTTransport)
	reads := make(chan []byte, 3)
	reads <- pingFrame[0:5]  // version and headers size
	reads <- pingFrame[5:34] // FContext headers
	reads <- pingFrame[34:]  // request body
	mockTransport.reads = reads
	// _opid0
	responseCtx := []byte{0, 0, 0, 0, 14, 0, 0, 0, 5, 95, 111, 112, 105, 100, 0, 0, 0, 1, 48}
	mockTransport.On("Write", responseCtx).Return(len(responseCtx), nil).Once()
	// [1,"ping",3,0,{"1":{"str":"Unknown function ping"},"2":{"i32":1}}]
	responseBody := []byte{
		91, 49, 44, 34, 112, 105, 110, 103, 34, 44, 51, 44, 48, 44, 123, 34,
		49, 34, 58, 123, 34, 115, 116, 114, 34, 58, 34, 85, 110, 107, 110, 111,
		119, 110, 32, 102, 117, 110, 99, 116, 105, 111, 110, 32, 112, 105, 110,
		103, 34, 125, 44, 34, 50, 34, 58, 123, 34, 105, 51, 50, 34, 58, 49,
		125, 125, 93,
	}
	mockTransport.On("Write", responseBody).Return(len(responseBody), nil).Once()
	mockTransport.On("Flush").Return(nil)
	proto := &FProtocol{thrift.NewTJSONProtocol(mockTransport)}
	processor := NewFBaseProcessor()

	err := processor.Process(proto, proto)
	assert.Error(t, err)
	appErr := err.(thrift.TApplicationException)
	assert.Equal(t, int(thrift.UNKNOWN_METHOD), int(appErr.TypeId()))
	assert.Equal(t, "Unknown function ping", appErr.Error())
	mockTransport.AssertExpectations(t)
}
