/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package python

import (
	"fmt"
	"os"
	"strings"

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
	imports += "from frugal.exceptions import TApplicationExceptionType\n"
	imports += "from frugal.exceptions import TTransportExceptionType\n"
	imports += "from frugal.middleware import Method\n"
	imports += "from frugal.transport import TMemoryOutputBuffer\n"
	imports += "from frugal.util.deprecate import deprecated\n"
	imports += "from thrift.Thrift import TApplicationException\n"
	imports += "from thrift.Thrift import TMessageType\n"
	imports += "from thrift.transport.TTransport import TTransportException\n"

	imports += a.generateServiceExtendsImport(s)
	if imp, err := a.generateServiceIncludeImports(s); err != nil {
		return err
	} else {
		imports += imp
	}

	_, err := file.WriteString(imports)
	return err
}

// GenerateScopeImports generates necessary imports for the given scope.
func (a *AsyncIOGenerator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "import inspect\n"
	imports += "import sys\n"
	imports += "import traceback\n\n"

	imports += "from thrift.Thrift import TApplicationException\n"
	imports += "from thrift.Thrift import TMessageType\n"
	imports += "from thrift.Thrift import TType\n"
	imports += "from frugal.exceptions import TApplicationExceptionType\n"
	imports += "from frugal.middleware import Method\n"
	imports += "from frugal.subscription import FSubscription\n"
	imports += "from frugal.transport import TMemoryOutputBuffer\n\n"

	imports += "from .ttypes import *\n"
	_, err := file.WriteString(imports)
	return err
}

// GenerateService generates the given service.
func (a *AsyncIOGenerator) GenerateService(file *os.File, s *parser.Service) error {
	contents := ""
	contents += a.generateServiceInterface(s)
	contents += a.generateClient(s)
	contents += a.generateServer(s)
	contents += a.generateServiceArgsResults(s)

	_, err := file.WriteString(contents)
	return err
}

func (a *AsyncIOGenerator) generateClient(service *parser.Service) string {
	contents := "\n"
	if service.Extends != "" {
		contents += fmt.Sprintf("class Client(%s.Client, Iface):\n\n", a.getServiceExtendsName(service))
	} else {
		contents += "class Client(Iface):\n\n"
	}

	contents += tab + "def __init__(self, provider, middleware=None):\n"
	contents += a.generateDocString([]string{
		"Create a new Client with an FServiceProvider containing a transport",
		"and protocol factory.\n",
		"Args:",
		tab + "provider: FServiceProvider",
		tab + "middleware: ServiceMiddleware or list of ServiceMiddleware",
	}, tabtab)
	contents += tabtab + "middleware = middleware or []\n"
	contents += tabtab + "if middleware and not isinstance(middleware, list):\n"
	contents += tabtabtab + "middleware = [middleware]\n"
	if service.Extends != "" {
		contents += tabtab + "super(Client, self).__init__(provider, middleware=middleware)\n"
		contents += tabtab + "middleware += provider.get_middleware()\n"
		contents += tabtab + "self._methods.update("
	} else {
		contents += tabtab + "self._transport = provider.get_transport()\n"
		contents += tabtab + "self._protocol_factory = provider.get_protocol_factory()\n"
		contents += tabtab + "middleware += provider.get_middleware()\n"
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
	contents += tabtab + fmt.Sprintf("return await self._methods['%s']([ctx%s])\n\n",
		method.Name, a.generateClientArgs(method.Arguments))

	contents += tab + fmt.Sprintf("async def _%s(self, ctx%s):\n", method.Name, a.generateClientArgs(method.Arguments))
	contents += tabtab + "memory_buffer = TMemoryOutputBuffer(self._transport.get_request_size_limit())\n"
	contents += tabtab + "oprot = self._protocol_factory.get_protocol(memory_buffer)\n"
	contents += tabtab + "oprot.write_request_headers(ctx)\n"
	contents += tabtab + fmt.Sprintf("oprot.writeMessageBegin('%s', TMessageType.CALL, 0)\n", parser.LowercaseFirstLetter(method.Name))
	contents += tabtab + fmt.Sprintf("args = %s_args()\n", method.Name)
	for _, arg := range method.Arguments {
		contents += tabtab + fmt.Sprintf("args.%s = %s\n", arg.Name, arg.Name)
	}
	contents += tabtab + "args.write(oprot)\n"
	contents += tabtab + "oprot.writeMessageEnd()\n"

	if method.Oneway {
		contents += tabtab + "await self._transport.oneway(ctx, memory_buffer.getvalue())\n\n"
		return contents
	}
	contents += tabtab + "response_transport = await self._transport.request(ctx, memory_buffer.getvalue())\n\n"

	contents += tabtab + "iprot = self._protocol_factory.get_protocol(response_transport)\n"
	contents += tabtab + "iprot.read_response_headers(ctx)\n"
	contents += tabtab + "_, mtype, _ = iprot.readMessageBegin()\n"
	contents += tabtab + "if mtype == TMessageType.EXCEPTION:\n"
	contents += tabtabtab + "x = TApplicationException()\n"
	contents += tabtabtab + "x.read(iprot)\n"
	contents += tabtabtab + "iprot.readMessageEnd()\n"
	contents += tabtabtab + "if x.type == TApplicationExceptionType.RESPONSE_TOO_LARGE:\n"
	contents += tabtabtabtab + "raise TTransportException(type=TTransportExceptionType.RESPONSE_TOO_LARGE, message=x.message)\n"
	contents += tabtabtab + "raise x\n"
	contents += tabtab + fmt.Sprintf("result = %s_result()\n", method.Name)
	contents += tabtab + "result.read(iprot)\n"
	contents += tabtab + "iprot.readMessageEnd()\n"
	for _, err := range method.Exceptions {
		contents += tabtab + fmt.Sprintf("if result.%s is not None:\n", err.Name)
		contents += tabtabtab + fmt.Sprintf("raise result.%s\n", err.Name)
	}
	if method.ReturnType != nil {
		contents += tabtab + "if result.success is not None:\n"
		contents += tabtabtab + "return result.success\n"
		contents += tabtab + fmt.Sprintf(
			"raise TApplicationException(TApplicationExceptionType.MISSING_RESULT, \"%s failed: unknown result\")\n\n", method.Name)
	}
	return contents
}

func (a *AsyncIOGenerator) generateServer(service *parser.Service) string {
	contents := ""
	contents += a.generateProcessor(service)
	for _, method := range service.Methods {
		contents += a.generateProcessorFunction(method)
	}
	contents += a.generateWriteApplicationException()

	return contents
}

func (g *AsyncIOGenerator) generateProcessor(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += fmt.Sprintf("class Processor(%s.Processor):\n\n", g.getServiceExtendsName(service))
	} else {
		contents += "class Processor(FBaseProcessor):\n\n"
	}

	contents += tab + "def __init__(self, handler, middleware=None):\n"
	contents += g.generateDocString([]string{
		"Create a new Processor.\n",
		"Args:",
		tab + "handler: Iface",
	}, tabtab)

	contents += tabtab + "if middleware and not isinstance(middleware, list):\n"
	contents += tabtabtab + "middleware = [middleware]\n\n"

	if service.Extends != "" {
		contents += tabtab + "super(Processor, self).__init__(handler, middleware=middleware)\n"
	} else {
		contents += tabtab + "super(Processor, self).__init__()\n"
	}
	for _, method := range service.Methods {
		methodLower := parser.LowercaseFirstLetter(method.Name)
		contents += tabtab + fmt.Sprintf("self.add_to_processor_map('%s', _%s(Method(handler.%s, middleware), self.get_write_lock()))\n",
			methodLower, method.Name, method.Name)
		if len(method.Annotations) > 0 {
			annotations := make([]string, len(method.Annotations))
			for i, annotation := range method.Annotations {
				annotations[i] = fmt.Sprintf("'%s': '%s'", annotation.Name, annotation.Value)
			}
			contents += tabtab +
				fmt.Sprintf("self.add_to_annotations_map('%s', {%s})\n", methodLower, strings.Join(annotations, ", "))
		}
	}
	contents += "\n\n"

	return contents
}

func (a *AsyncIOGenerator) generateProcessorFunction(method *parser.Method) string {
	methodLower := parser.LowercaseFirstLetter(method.Name)
	contents := ""
	contents += fmt.Sprintf("class _%s(FProcessorFunction):\n\n", method.Name)
	contents += tab + "def __init__(self, handler, lock):\n"
	contents += tabtab + fmt.Sprintf("super(_%s, self).__init__(handler, lock)\n", method.Name)
	contents += "\n"

	if _, ok := method.Annotations.Deprecated(); ok {
		contents += tab + "@deprecated\n"
	}
	contents += tab + "async def process(self, ctx, iprot, oprot):\n"
	contents += tabtab + fmt.Sprintf("args = %s_args()\n", method.Name)
	contents += tabtab + "args.read(iprot)\n"
	contents += tabtab + "iprot.readMessageEnd()\n"
	if !method.Oneway {
		contents += tabtab + fmt.Sprintf("result = %s_result()\n", method.Name)
	}
	contents += tabtab + "try:\n"
	contents += tabtabtab + fmt.Sprintf("ret = self._handler([ctx%s])\n",
		a.generateServerArgs(method.Arguments))
	contents += tabtabtab + "if inspect.iscoroutine(ret):\n"
	contents += tabtabtabtab + "ret = await ret\n"
	if method.ReturnType != nil {
		contents += tabtabtab + "result.success = ret\n"
	}
	contents += tabtab + "except TApplicationException as ex:\n"
	contents += tabtabtab + "async with self._lock:\n"
	contents += tabtabtabtab +
		fmt.Sprintf("_write_application_exception(ctx, oprot, \"%s\", exception=ex)\n",
			methodLower)
	contents += tabtabtabtab + "return\n"
	for _, err := range method.Exceptions {
		contents += tabtab + fmt.Sprintf("except %s as %s:\n", a.qualifiedTypeName(err.Type), err.Name)
		contents += tabtabtab + fmt.Sprintf("result.%s = %s\n", err.Name, err.Name)
	}
	contents += tabtab + "except Exception as e:\n"
	if !method.Oneway {
		contents += tabtabtab + "async with self._lock:\n"
		contents += tabtabtabtab + fmt.Sprintf("e = _write_application_exception(ctx, oprot, \"%s\", ex_code=TApplicationExceptionType.INTERNAL_ERROR, message=e.args[0])\n", methodLower)
	}
	contents += tabtabtab + "raise e from None\n"
	if !method.Oneway {
		contents += tabtab + "async with self._lock:\n"
		contents += tabtabtab + "try:\n"
		contents += tabtabtabtab + "oprot.write_response_headers(ctx)\n"
		contents += tabtabtabtab + fmt.Sprintf("oprot.writeMessageBegin('%s', TMessageType.REPLY, 0)\n", methodLower)
		contents += tabtabtabtab + "result.write(oprot)\n"
		contents += tabtabtabtab + "oprot.writeMessageEnd()\n"
		contents += tabtabtabtab + "oprot.get_transport().flush()\n"
		contents += tabtabtab + "except TTransportException as e:\n"
		contents += tabtabtabtab + "# catch a request too large error because the TMemoryOutputBuffer always throws that if too much data is written\n"
		contents += tabtabtabtab + "if e.type == TTransportExceptionType.REQUEST_TOO_LARGE:\n"
		contents += tabtabtabtabtab + fmt.Sprintf(
			"raise _write_application_exception(ctx, oprot, \"%s\", ex_code=TApplicationExceptionType.RESPONSE_TOO_LARGE, message=e.message)\n", methodLower)
		contents += tabtabtabtab + "else:\n"
		contents += tabtabtabtabtab + "raise e\n"
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
	subscriber += tabtab + "middleware = middleware or []\n"
	subscriber += tabtab + "if middleware and not isinstance(middleware, list):\n"
	subscriber += tabtabtab + "middleware = [middleware]\n"
	subscriber += tabtab + "middleware += provider.get_middleware()\n"
	subscriber += tabtab + "self._middleware = middleware\n"
	subscriber += tabtab + "self._provider = provider\n\n"

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

	method += tabtab + "transport, protocol_factory = self._provider.new_subscriber()\n"
	method += tabtab + fmt.Sprintf(
		"await transport.subscribe(topic, self._recv_%s(protocol_factory, op, %s_handler))\n",
		op.Name, op.Name)
	method += tabtab + "return FSubscription(topic, transport)\n\n"

	method += tab + fmt.Sprintf("def _recv_%s(self, protocol_factory, op, handler):\n", op.Name)
	method += tabtab + "method = Method(handler, self._middleware)\n\n"

	method += tabtab + "async def callback(transport):\n"
	method += tabtabtab + "iprot = protocol_factory.get_protocol(transport)\n"
	method += tabtabtab + "ctx = iprot.read_request_headers()\n"
	method += tabtabtab + "mname, _, _ = iprot.readMessageBegin()\n"
	method += tabtabtab + "if mname != op:\n"
	method += tabtabtabtab + "iprot.skip(TType.STRUCT)\n"
	method += tabtabtabtab + "iprot.readMessageEnd()\n"
	method += tabtabtabtab + "raise TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD)\n"
	method += a.generateReadFieldRec(parser.FieldFromType(op.Type, "req"), false, tabtabtab)
	method += tabtabtab + "iprot.readMessageEnd()\n"
	method += tabtabtab + "try:\n"
	method += tabtabtabtab + "ret = method([ctx, req])\n"
	method += tabtabtabtab + "if inspect.iscoroutine(ret):\n"
	method += tabtabtabtabtab + "await ret\n"
	method += tabtabtab + "except:\n"
	method += tabtabtabtab + "traceback.print_exc()\n"
	method += tabtabtabtab + "sys.exit(1)\n\n"

	method += tabtab + "return callback\n\n"

	return method
}
