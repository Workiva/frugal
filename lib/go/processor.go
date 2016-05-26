package frugal

import (
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
)

// FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic
// object which operates upon an input stream and writes to an output stream.
// Specifically, an FProcessor is provided to an FServer in order to wire up a
// service handler to process requests.
type FProcessor interface {
	Process(in, out *FProtocol) error
}

// FBaseProcessor is a base implementation of FProcessor. FProcessors should
// embed this and register FProcessorFunctions. This should only be used by
// generated code.
type FBaseProcessor struct {
	writeMu    sync.Mutex
	processMap map[string]FProcessorFunction
}

func NewFBaseProcessor() *FBaseProcessor {
	return &FBaseProcessor{processMap: make(map[string]FProcessorFunction)}
}

func (f *FBaseProcessor) Process(iprot, oprot *FProtocol) error {
	ctx, err := iprot.ReadRequestHeader()
	if err != nil {
		return err
	}
	name, _, _, err := iprot.ReadMessageBegin()
	if err != nil {
		return err
	}
	processor, ok := f.processMap[name]
	if ok {
		err := processor.Process(ctx, iprot, oprot)
		if err != nil {
			log.Warnf("frugal: error processing request with correlation id %s: %s", ctx.CorrelationID(), err.Error())
		}
		return err
	}
	iprot.Skip(thrift.STRUCT)
	iprot.ReadMessageEnd()
	ex := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+name)
	f.writeMu.Lock()
	oprot.WriteResponseHeader(ctx)
	oprot.WriteMessageBegin(name, thrift.EXCEPTION, 0)
	ex.Write(oprot)
	oprot.WriteMessageEnd()
	oprot.Flush()
	f.writeMu.Unlock()
	return ex
}

// AddToProcessorMap registers the given FProcessorFunction.
func (f *FBaseProcessor) AddToProcessorMap(key string, proc FProcessorFunction) {
	f.processMap[key] = proc
}

// GetWriteMutex returns the Mutex which FProcessorFunctions should use to
// synchronize access to the output FProtocol.
func (f *FBaseProcessor) GetWriteMutex() *sync.Mutex {
	return &f.writeMu
}

// FProcessorFunction is used internally by generated code. An FProcessor
// registers an FProcessorFunction for each service method. Like FProcessor, an
// FProcessorFunction exposes a single process call, which is used to handle a
// method invocation.
type FProcessorFunction interface {
	Process(ctx *FContext, in, out *FProtocol) error
}

// FProcessorFactory produces FProcessors and is used by an FServer. It takes a
// TTransport and returns an FProcessor wrapping it.
type FProcessorFactory interface {
	GetProcessor(trans thrift.TTransport) FProcessor
}

type fProcessorFactory struct {
	processor FProcessor
}

func NewFProcessorFactory(p FProcessor) FProcessorFactory {
	return &fProcessorFactory{processor: p}
}

func (p *fProcessorFactory) GetProcessor(trans thrift.TTransport) FProcessor {
	return p.processor
}
