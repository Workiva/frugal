package frugal

import (
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic
// object which operates upon an input stream and writes to an output stream.
// Specifically, an FProcessor is provided to an FServer in order to wire up a
// service handler to process requests.
type FProcessor interface {
	// Process the request from the input protocol and write the response to
	// the output protocol.
	Process(in, out *FProtocol) error

	// AddMiddleware adds the given ServiceMiddleware to the FProcessor. This
	// should only be called before the server is started.
	AddMiddleware(ServiceMiddleware)

	// Annotations returns a map of method name to annotations as defined in
	// the service IDL that is serviced by this processor.
	Annotations() map[string]map[string]string
}

// FBaseProcessor is a base implementation of FProcessor. FProcessors should
// embed this and register FProcessorFunctions. This should only be used by
// generated code.
type FBaseProcessor struct {
	writeMu        sync.Mutex
	processMap     map[string]FProcessorFunction
	annotationsMap map[string]map[string]string
}

// NewFBaseProcessor returns a new FBaseProcessor which FProcessors can extend.
func NewFBaseProcessor() *FBaseProcessor {
	return &FBaseProcessor{
		processMap:     make(map[string]FProcessorFunction),
		annotationsMap: make(map[string]map[string]string),
	}
}

// Process the request from the input protocol and write the response to the
// output protocol.
func (f *FBaseProcessor) Process(iprot, oprot *FProtocol) error {
	ctx, err := iprot.ReadRequestHeader()
	if err != nil {
		return err
	}
	name, _, _, err := iprot.ReadMessageBegin()
	if err != nil {
		return err
	}
	if processor, ok := f.processMap[name]; ok {
		if err := processor.Process(ctx, iprot, oprot); err != nil {
			if _, ok := err.(thrift.TException); ok {
				logger().Errorf(
					"frugal: error occurred while processing request with correlation id %s: %s",
					ctx.CorrelationID(), err.Error())
			} else {
				logger().Errorf(
					"frugal: user handler code returned unhandled error on request with correlation id %s: %s",
					ctx.CorrelationID(), err.Error())
			}
		}
		// Return nil because the server should still send a response to the client.
		return nil
	}

	logger().Warnf("frugal: client invoked unknown function %s on request with correlation id %s",
		name, ctx.CorrelationID())
	if err := iprot.Skip(thrift.STRUCT); err != nil {
		return err
	}
	if err := iprot.ReadMessageEnd(); err != nil {
		return err
	}
	ex := thrift.NewTApplicationException(APPLICATION_EXCEPTION_UNKNOWN_METHOD, "Unknown function "+name)
	f.writeMu.Lock()
	defer f.writeMu.Unlock()
	if err := oprot.WriteResponseHeader(ctx); err != nil {
		return err
	}
	if err := oprot.WriteMessageBegin(name, thrift.EXCEPTION, 0); err != nil {
		return err
	}
	if err := ex.Write(oprot); err != nil {
		return err
	}
	if err := oprot.WriteMessageEnd(); err != nil {
		return err
	}
	if err := oprot.Flush(); err != nil {
		return err
	}
	return nil
}

// AddMiddleware adds the given ServiceMiddleware to the FProcessor. This
// should only be called before the server is started.
func (f *FBaseProcessor) AddMiddleware(middleware ServiceMiddleware) {
	for _, p := range f.processMap {
		p.AddMiddleware(middleware)
	}
}

// AddToProcessorMap registers the given FProcessorFunction.
func (f *FBaseProcessor) AddToProcessorMap(key string, proc FProcessorFunction) {
	f.processMap[key] = proc
}

// AddToAnnotationsMap registers the given annotations to the given method.
func (f *FBaseProcessor) AddToAnnotationsMap(method string, annotations map[string]string) {
	f.annotationsMap[method] = annotations
}

// Annotations returns a map of method name to annotations as defined in
// the service IDL that is serviced by this processor.
func (f *FBaseProcessor) Annotations() map[string]map[string]string {
	annoCopy := make(map[string]map[string]string)
	for k, v := range f.annotationsMap {
		methodCopy := make(map[string]string)
		for mk, mv := range v {
			methodCopy[mk] = mv
		}
		annoCopy[k] = methodCopy
	}
	return annoCopy
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
	// Process the request from the input protocol and write the response to
	// the output protocol.
	Process(ctx FContext, in, out *FProtocol) error

	// AddMiddleware adds the given ServiceMiddleware to the
	// FProcessorFunction. This should only be called before the server is
	// started.
	AddMiddleware(middleware ServiceMiddleware)
}

// FBaseProcessorFunction is a base implementation of FProcessorFunction.
// FProcessorFunctions should embed this. This should only be used by generated
// code.
type FBaseProcessorFunction struct {
	handler *Method
	writeMu *sync.Mutex
}

// NewFBaseProcessorFunction returns a new FBaseProcessorFunction which
// FProcessorFunctions can extend.
func NewFBaseProcessorFunction(writeMu *sync.Mutex, handler *Method) *FBaseProcessorFunction {
	return &FBaseProcessorFunction{handler, writeMu}
}

// GetWriteMutex returns the Mutex which should be used to synchronize access
// to the output FProtocol.
func (f *FBaseProcessorFunction) GetWriteMutex() *sync.Mutex {
	return f.writeMu
}

// AddMiddleware adds the given ServiceMiddleware to the FProcessorFunction.
// This should only be called before the server is started.
func (f *FBaseProcessorFunction) AddMiddleware(middleware ServiceMiddleware) {
	f.handler.AddMiddleware(middleware)
}

// InvokeMethod invokes the handler method.
func (f *FBaseProcessorFunction) InvokeMethod(args []interface{}) Results {
	return f.handler.Invoke(args)
}
