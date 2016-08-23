package python

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

// AsyncIOGenerator implements the LanguageGenerator interface for Python using
// AsyncIO.
type AsyncIOGenerator struct {
	*Generator
}

// GenerateServiceImports generates necessary imports for the given service.
func (a *AsyncIOGenerator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	imports := "import asyncio\n"
	imports += "from datetime import timedelta\n"
	imports += "import inspect\n\n"

	imports += "from frugal.aio.processor import FBaseProcessor\n"
	imports += "from frugal.aio.processor import FProcessorFunction\n"
	imports += "from frugal.aio.registry import FClientRegistry\n"
	imports += "from frugal.middleware import Method\n"
	imports += "from thrift.Thrift import TApplicationException\n"
	imports += "from thrift.Thrift import TMessageType\n"

	// Import include modules.
	for _, include := range s.ReferencedIncludes() {
		namespace, ok := a.Frugal.NamespaceForInclude(include, lang)
		if !ok {
			namespace = include
		}
		imports += fmt.Sprintf("import %s\n", namespace)
	}

	// Import this service's modules.
	namespace, ok := a.Frugal.Thrift.Namespace(lang)
	if !ok {
		namespace = a.Frugal.Name
	}
	imports += fmt.Sprintf("from %s.%s import *\n", namespace, s.Name)
	imports += fmt.Sprintf("from %s.ttypes import *\n", namespace)

	_, err := file.WriteString(imports)
	return err
}

// GenerateScopeImports generates necessary imports for the given scope.
func (a *AsyncIOGenerator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "import sys\n"
	imports += "import traceback\n\n"

	imports += "from thrift.Thrift import TApplicationException\n"
	imports += "from thrift.Thrift import TMessageType\n"
	imports += "from thrift.Thrift import TType\n"
	imports += "from frugal.middleware import Method\n"
	imports += "from frugal.subscription import FSubscription\n\n"

	namespace, ok := a.Frugal.Thrift.Namespace(lang)
	if !ok {
		namespace = a.Frugal.Name
	}
	imports += fmt.Sprintf("from %s.ttypes import *\n", namespace)
	_, err := file.WriteString(imports)
	return err
}

// GenerateService generates the given service.
func (a *AsyncIOGenerator) GenerateService(file *os.File, s *parser.Service) error {
	if err := a.exposeServiceModule(filepath.Dir(file.Name()), s); err != nil {
		return err
	}
	contents := ""
	contents += a.generateServiceInterface(s)
	contents += a.generateClient(s)
	contents += a.generateServer(s)

	_, err := file.WriteString(contents)
	return err
}

func (a *AsyncIOGenerator) exposeServiceModule(path string, service *parser.Service) error {
	initFile := fmt.Sprintf("%s%s__init__.py", path, string(os.PathSeparator))
	init, err := ioutil.ReadFile(initFile)
	if err != nil {
		return err
	}
	initStr := string(init)
	initStr += fmt.Sprintf("\nfrom . import f_%s\n", service.Name)
	initStr += fmt.Sprintf("from .f_%s import *\n", service.Name)
	init = []byte(initStr)
	return ioutil.WriteFile(initFile, init, os.ModePerm)
}

func (a *AsyncIOGenerator) generateClient(service *parser.Service) string {
	contents := "\n"
	if service.Extends != "" {
		contents += fmt.Sprintf("class Client(%s.Client, Iface):\n\n", a.getServiceExtendsName(service))
	} else {
		contents += "class Client(Iface):\n\n"
	}

	contents += tab + "def __init__(self, transport, protocol_factory, middleware=None):\n"
	contents += a.generateDocString([]string{
		"Create a new Client with a transport and protocol factory.\n",
		"Args:",
		tab + "transport: FTransport",
		tab + "protocol_factory: FProtocolFactory",
		tab + "middleware: ServiceMiddleware or list of ServiceMiddleware",
	}, tabtab)
	contents += tabtab + "if middleware and not isinstance(middleware, list):\n"
	contents += tabtabtab + "middleware = [middleware]\n"
	if service.Extends != "" {
		contents += tabtab + "super(Client, self).__init__(transport, protocol_factory,\n"
		contents += tabtab + "                             middleware=middleware)\n"
		contents += tabtab + "self._methods.update("
	} else {
		contents += tabtab + "transport.set_registry(FClientRegistry())\n"
		contents += tabtab + "self._transport = transport\n"
		contents += tabtab + "self._protocol_factory = protocol_factory\n"
		contents += tabtab + "self._oprot = protocol_factory.get_protocol(transport)\n"
		contents += tabtab + "self._write_lock = asyncio.Lock()\n"
		contents += tabtab + "self._methods = "
	}
	contents += "{\n"
	for _, method := range service.Methods {
		contents += tabtabtab + fmt.Sprintf("'%s': Method(self._%s, middleware),\n", method.Name, method.Name)
	}
	contents += tabtab + "}"
	if service.Extends != "" {
		contents += ")"
	}
	contents += "\n\n"

	for _, method := range service.Methods {
		contents += a.generateClientMethod(method)
	}
	contents += "\n"

	return contents
}

func (a *AsyncIOGenerator) generateClientMethod(method *parser.Method) string {
	contents := ""
	contents += a.generateMethodSignature(method)
	contents += tabtab + fmt.Sprintf("await self._methods['%s']([ctx%s])\n\n",
		method.Name, a.generateClientArgs(method.Arguments))

	if method.Oneway {
		contents += tab + fmt.Sprintf("async def _%s(self, ctx%s):\n", method.Name, a.generateClientArgs(method.Arguments))
		contents += tabtab + fmt.Sprintf("await self._send_%s(ctx%s)\n\n", method.Name, a.generateClientArgs(method.Arguments))
		contents += a.generateClientSendMethod(method)
		return contents
	}

	contents += tab + fmt.Sprintf("async def _%s(self, ctx%s):\n", method.Name, a.generateClientArgs(method.Arguments))
	contents += tabtab + "timeout = ctx.get_timeout() / 1000.0\n"
	contents += tabtab + "future = asyncio.Future()\n"
	contents += tabtab + "timed_future = asyncio.wait_for(future, timeout)\n"
	contents += tabtab + fmt.Sprintf("await self._transport.register(ctx, self._recv_%s(ctx, future))\n", method.Name)
	contents += tabtab + fmt.Sprintf("await self._send_%s(ctx%s)\n\n", method.Name, a.generateClientArgs(method.Arguments))
	contents += tabtab + "try:\n"
	contents += tabtabtab + "result = await timed_future\n"
	contents += tabtab + "finally:\n"
	contents += tabtabtab + "await self._transport.unregister(ctx)\n"
	contents += tabtab + "return result\n\n"
	contents += a.generateClientSendMethod(method)
	contents += a.generateClientRecvMethod(method)

	return contents
}

func (a *AsyncIOGenerator) generateClientSendMethod(method *parser.Method) string {
	contents := ""
	contents += tab + fmt.Sprintf("async def _send_%s(self, ctx%s):\n", method.Name, a.generateClientArgs(method.Arguments))
	contents += tabtab + "oprot = self._oprot\n"
	contents += tabtab + "async with self._write_lock:\n"
	contents += tabtabtab + "oprot.write_request_headers(ctx)\n"
	contents += tabtabtab + fmt.Sprintf("oprot.writeMessageBegin('%s', TMessageType.CALL, 0)\n", method.Name)
	contents += tabtabtab + fmt.Sprintf("args = %s_args()\n", method.Name)
	for _, arg := range method.Arguments {
		contents += tabtabtab + fmt.Sprintf("args.%s = %s\n", arg.Name, arg.Name)
	}
	contents += tabtabtab + "args.write(oprot)\n"
	contents += tabtabtab + "oprot.writeMessageEnd()\n"
	contents += tabtabtab + "await oprot.get_transport().flush()\n\n"

	return contents
}

func (a *AsyncIOGenerator) generateClientRecvMethod(method *parser.Method) string {
	contents := tab + fmt.Sprintf("def _recv_%s(self, ctx, future):\n", method.Name)
	contents += tabtab + fmt.Sprintf("def %s_callback(transport):\n", method.Name)
	contents += tabtabtab + "iprot = self._protocol_factory.get_protocol(transport)\n"
	contents += tabtabtab + "iprot.read_response_headers(ctx)\n"
	contents += tabtabtab + "_, mtype, _ = iprot.readMessageBegin()\n"
	contents += tabtabtab + "if mtype == TMessageType.EXCEPTION:\n"
	contents += tabtabtabtab + "x = TApplicationException()\n"
	contents += tabtabtabtab + "x.read(iprot)\n"
	contents += tabtabtabtab + "iprot.readMessageEnd()\n"
	contents += tabtabtabtab + "future.set_exception(x)\n"
	contents += tabtabtabtab + "raise x\n"
	contents += tabtabtab + fmt.Sprintf("result = %s_result()\n", method.Name)
	contents += tabtabtab + "result.read(iprot)\n"
	contents += tabtabtab + "iprot.readMessageEnd()\n"
	for _, err := range method.Exceptions {
		contents += tabtabtab + fmt.Sprintf("if result.%s is not None:\n", err.Name)
		contents += tabtabtabtab + fmt.Sprintf("future.set_exception(result.%s)\n", err.Name)
		contents += tabtabtabtab + "return\n"
	}
	if method.ReturnType == nil {
		contents += tabtabtab + "future.set_result(None)\n"
	} else {
		contents += tabtabtab + "if result.success is not None:\n"
		contents += tabtabtabtab + "future.set_result(result.success)\n"
		contents += tabtabtabtab + "return\n"
		contents += tabtabtab + fmt.Sprintf(
			"x = TApplicationException(TApplicationException.MISSING_RESULT, \"%s failed: unknown result\")\n", method.Name)
		contents += tabtabtab + "future.set_exception(x)\n"
		contents += tabtabtab + "raise x\n"
	}
	contents += tabtab + fmt.Sprintf("return %s_callback\n\n", method.Name)

	return contents
}

func (a *AsyncIOGenerator) generateServer(service *parser.Service) string {
	contents := ""
	contents += a.generateProcessor(service)
	for _, method := range service.Methods {
		contents += a.generateProcessorFunction(method)
	}

	return contents
}

func (g *AsyncIOGenerator) generateProcessor(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += fmt.Sprintf("class Processor(%s.Processor):\n\n", g.getServiceExtendsName(service))
	} else {
		contents += "class Processor(FBaseProcessor):\n\n"
	}

	contents += tab + "def __init__(self, handler):\n"
	contents += g.generateDocString([]string{
		"Create a new Processor.\n",
		"Args:",
		tab + "handler: Iface",
	}, tabtab)
	if service.Extends != "" {
		contents += tabtab + "super(Processor, self).__init__(handler)\n"
	} else {
		contents += tabtab + "super(Processor, self).__init__()\n"
	}
	for _, method := range service.Methods {
		contents += tabtab + fmt.Sprintf("self.add_to_processor_map('%s', _%s(handler, self.get_write_lock()))\n",
			method.Name, method.Name)
	}
	contents += "\n\n"
	return contents
}

func (a *AsyncIOGenerator) generateProcessorFunction(method *parser.Method) string {
	contents := ""
	contents += fmt.Sprintf("class _%s(FProcessorFunction):\n\n", method.Name)
	contents += tab + "def __init__(self, handler, lock):\n"
	contents += tabtab + "self._handler = handler\n"
	contents += tabtab + "self._write_lock = lock\n\n"

	contents += tab + "async def process(self, ctx, iprot, oprot):\n"
	contents += tabtab + fmt.Sprintf("args = %s_args()\n", method.Name)
	contents += tabtab + "args.read(iprot)\n"
	contents += tabtab + "iprot.readMessageEnd()\n"
	if !method.Oneway {
		contents += tabtab + fmt.Sprintf("result = %s_result()\n", method.Name)
	}
	indent := tabtab
	if len(method.Exceptions) > 0 {
		indent += tab
		contents += tabtab + "try:\n"
	}
	contents += indent + fmt.Sprintf("ret = self._handler.%s(ctx%s)\n",
		method.Name, a.generateServerArgs(method.Arguments))
	contents += indent + "if inspect.iscoroutine(ret):\n"
	contents += indent + tab + "ret = await ret\n"
	if method.ReturnType != nil {
		contents += indent + "result.success = ret\n"
	}
	for _, err := range method.Exceptions {
		contents += tabtab + fmt.Sprintf("except %s as %s:\n", a.qualifiedTypeName(err.Type), err.Name)
		contents += tabtabtab + fmt.Sprintf("result.%s = %s\n", err.Name, err.Name)
	}
	if !method.Oneway {
		contents += tabtab + "async with self._write_lock:\n"
		contents += tabtabtab + "oprot.write_response_headers(ctx)\n"
		contents += tabtabtab + fmt.Sprintf("oprot.writeMessageBegin('%s', TMessageType.REPLY, 0)\n", method.Name)
		contents += tabtabtab + "result.write(oprot)\n"
		contents += tabtabtab + "oprot.writeMessageEnd()\n"
		contents += tabtabtab + "oprot.get_transport().flush()\n"
	}
	contents += "\n\n"

	return contents
}

// GenerateSubscriber generates the subscriber for the given scope.
func (a *AsyncIOGenerator) GenerateSubscriber(file *os.File, scope *parser.Scope) error {
	subscriber := ""
	subscriber += fmt.Sprintf("class %sSubscriber(object):\n", scope.Name)
	if scope.Comment != nil {
		subscriber += a.generateDocString(scope.Comment, tab)
	}
	subscriber += "\n"

	subscriber += tab + fmt.Sprintf("_DELIMITER = '%s'\n\n", globals.TopicDelimiter)

	subscriber += tab + "def __init__(self, provider, middleware=None):\n"
	subscriber += a.generateDocString([]string{
		fmt.Sprintf("Create a new %sSubscriber.\n", scope.Name),
		"Args:",
		tab + "provider: FScopeProvider",
		tab + "middleware: ServiceMiddleware or list of ServiceMiddleware",
	}, tabtab)
	subscriber += "\n"
	subscriber += tabtab + "if middleware and not isinstance(middleware, list):\n"
	subscriber += tabtabtab + "middleware = [middleware]\n"
	subscriber += tabtab + "self._middleware = middleware\n"
	subscriber += tabtab + "self._transport, self._protocol_factory = provider.new()\n\n"

	for _, op := range scope.Operations {
		subscriber += a.generateSubscribeMethod(scope, op)
		subscriber += "\n\n"
	}

	_, err := file.WriteString(subscriber)
	return err
}

func (a *AsyncIOGenerator) generateSubscribeMethod(scope *parser.Scope, op *parser.Operation) string {
	args := ""
	docstr := []string{}
	if len(scope.Prefix.Variables) > 0 {
		docstr = append(docstr, "Args:")
		prefix := ""
		for _, variable := range scope.Prefix.Variables {
			docstr = append(docstr, tab+fmt.Sprintf("%s: string", variable))
			args += prefix + variable
			prefix = ", "
		}
		args += ", "
	}
	docstr = append(docstr, tab+fmt.Sprintf("%s_handler: function which takes FContext and %s", op.Name, op.Type))
	if op.Comment != nil {
		docstr[0] = "\n" + tabtab + docstr[0]
		docstr = append(op.Comment, docstr...)
	}
	method := ""
	method += tab + fmt.Sprintf("async def subscribe_%s(self, %s%s_handler):\n", op.Name, args, op.Name)
	method += a.generateDocString(docstr, tabtab)
	method += "\n"

	method += tabtab + fmt.Sprintf("op = '%s'\n", op.Name)
	method += tabtab + fmt.Sprintf("prefix = %s\n", generatePrefixStringTemplate(scope))
	method += tabtab + fmt.Sprintf("topic = '{}%s{}{}'.format(prefix, self._DELIMITER, op)\n\n", scope.Name)

	method += tabtab + fmt.Sprintf(
		"await self._transport.subscribe(topic, self._recv_%s(self._protocol_factory, op, %s_handler))\n\n",
		op.Name, op.Name)

	method += tab + fmt.Sprintf("def _recv_%s(self, protocol_factory, op, handler):\n", op.Name)
	method += tabtab + "method = Method(handler, self._middleware)\n\n"

	method += tabtab + "def callback(transport):\n"
	method += tabtabtab + "iprot = protocol_factory.get_protocol(transport)\n"
	method += tabtabtab + "ctx = iprot.read_request_headers()\n"
	method += tabtabtab + "mname, _, _ = iprot.readMessageBegin()\n"
	method += tabtabtab + "if mname != op:\n"
	method += tabtabtabtab + "iprot.skip(TType.STRUCT)\n"
	method += tabtabtabtab + "iprot.readMessageEnd()\n"
	method += tabtabtabtab + "raise TApplicationException(TApplicationException.UNKNOWN_METHOD)\n"
	method += tabtabtab + fmt.Sprintf("req = %s()\n", op.Type.Name)
	method += tabtabtab + "req.read(iprot)\n"
	method += tabtabtab + "iprot.readMessageEnd()\n"
	method += tabtabtab + "try:\n"
	method += tabtabtabtab + "method([ctx, req])\n"
	method += tabtabtab + "except:\n"
	method += tabtabtabtab + "traceback.print_exc()\n"
	method += tabtabtabtab + "sys.exit(1)\n\n"

	method += tabtab + "return callback\n\n"

	return method
}
