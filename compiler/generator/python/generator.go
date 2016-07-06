package python

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

const (
	lang             = "py"
	defaultOutputDir = "gen-py"
	tab              = "    "
	tabtab           = tab + tab
	tabtabtab        = tab + tab + tab
	tabtabtabtab     = tab + tab + tab + tab
)

// Generator implements the LanguageGenerator interface for Python.
type Generator struct {
	*generator.BaseGenerator
	outputDir string
	typesFile *os.File
}

// NewGenerator creates a new Python LanguageGenerator.
func NewGenerator(options map[string]string) generator.LanguageGenerator {
	gen := &Generator{&generator.BaseGenerator{Options: options}, "", nil}
	if _, ok := options["tornado"]; ok {
		return &TornadoGenerator{gen}
	}
	return gen
}

// TODO Unimplemented methods

// SetupGenerator performs any setup logic before generation.
func (g *Generator) SetupGenerator(outputDir string) error {
	g.outputDir = outputDir

	typesFile, err := g.GenerateFile("types", outputDir, generator.ObjectFile)
	if err != nil {
		return err
	}
	if err = g.GenerateDocStringComment(typesFile); err != nil {
		return err
	}
	if _, err = typesFile.WriteString("\n\n"); err != nil {
		return err
	}
	if err = g.GenerateTypesImports(typesFile); err != nil {
		return err
	}
	if _, err = typesFile.WriteString("\n\n"); err != nil {
		return err
	}
	g.typesFile = typesFile

	return nil
}

// TeardownGenerator is run after generation.
func (g *Generator) TeardownGenerator() error {
	return g.typesFile.Close()
}

// GenerateConstantsContents generates constants.
func (g *Generator) GenerateConstantsContents(constants []*parser.Constant) error {
	contents := "\n\n"
	contents += "from thrift.Thrift import TType, TMessageType, TException, TApplicationException\n"
	contents += "from f_types import *\n\n"

	for _, constant := range constants {
		value := g.generateConstantValue(constant.Type, constant.Value, "")
		contents += fmt.Sprintf("%s = %s\n", constant.Name, value)
	}

	file, err := g.GenerateFile("constants", g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.GenerateDocStringComment(file); err != nil {
		return err
	}
	_, err = file.WriteString(contents)
	return err
}

func (g *Generator) generateConstantValue(t *parser.Type, value interface{}, ind string) string {
	underlyingType := g.Frugal.UnderlyingType(t)
	// If the value being referenced is of type Identifier, it's referencing
	// another constant. Need to recurse to get that value.
	identifier, ok := value.(parser.Identifier)
	// TODO consolidate this between generators
	if ok {
		name := string(identifier)

		// split based on '.', if present, it should be from an include
		pieces := strings.Split(name, ".")
		if len(pieces) == 1 {
			// From this file
			for _, constant := range g.Frugal.Thrift.Constants {
				if name == constant.Name {
					return g.generateConstantValue(t, constant.Value, ind)
				}
			}
		} else if len(pieces) == 2 {
			// Either from an include, or part of an enum
			for _, enum := range g.Frugal.Thrift.Enums {
				if pieces[0] == enum.Name {
					for _, value := range enum.Values {
						if pieces[1] == value.Name {
							return fmt.Sprintf("%v", value.Value)
						}
					}
					panic(fmt.Sprintf("referenced value '%s' of enum '%s' doesn't exist", pieces[1], pieces[0]))
				}
			}

			// If not part of an enum , it's from an include
			include, ok := g.Frugal.ParsedIncludes[pieces[0]]
			if !ok {
				panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
			}
			for _, constant := range include.Thrift.Constants {
				if pieces[1] == constant.Name {
					return g.generateConstantValue(t, constant.Value, ind)
				}
			}
		} else if len(pieces) == 3 {
			// enum from an include
			include, ok := g.Frugal.ParsedIncludes[pieces[0]]
			if !ok {
				panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
			}
			for _, enum := range include.Thrift.Enums {
				if pieces[1] == enum.Name {
					for _, value := range enum.Values {
						if pieces[2] == value.Name {
							return fmt.Sprintf("%v", value.Value)
						}
					}
					panic(fmt.Sprintf("referenced value '%s' of enum '%s' doesn't exist", pieces[1], pieces[0]))
				}
			}
		}

		panic("referenced constant doesn't exist: " + name)
	}

	if parser.IsThriftPrimitive(underlyingType) || parser.IsThriftContainer(underlyingType) {
		switch underlyingType.Name {
		case "bool":
			return strings.Title(fmt.Sprintf("%v", value))
		case "i8", "byte", "i16", "i32", "i64", "double":
			return fmt.Sprintf("%v", value)
		case "string", "binary":
			return fmt.Sprintf("\"%s\"", value)
		case "list", "set":
			contents := ""
			if underlyingType.Name == "set" {
				contents += "set("
			}
			contents += "[\n"
			for _, v := range value.([]interface{}) {
				val := g.generateConstantValue(underlyingType.ValueType, v, ind+tab)
				contents += fmt.Sprintf(ind+tab+"%s,\n", val)
			}
			contents += ind + "]"
			if underlyingType.Name == "set" {
				contents += ")"
			}
			return contents
		case "map":
			contents := "{\n"
			for _, pair := range value.([]parser.KeyValue) {
				key := g.generateConstantValue(underlyingType.KeyType, pair.Key, ind+tab)
				val := g.generateConstantValue(underlyingType.ValueType, pair.Value, ind+tab)
				contents += fmt.Sprintf(ind+tab+"%s: %s,\n", key, val)
			}
			contents += ind + "}"
			return contents
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return fmt.Sprintf("%d", value)
	} else if g.Frugal.IsStruct(underlyingType) {
		var s *parser.Struct
		for _, potential := range g.Frugal.Thrift.Structs {
			if underlyingType.Name == potential.Name {
				s = potential
				break
			}
		}

		contents := ""

		contents += fmt.Sprintf("%s(**{\n", underlyingType.Name)
		for _, pair := range value.([]parser.KeyValue) {
			name := pair.Key.(string)
			for _, field := range s.Fields {
				if name == field.Name {
					val := g.generateConstantValue(field.Type, pair.Value, ind+tab)
					contents += fmt.Sprintf(tab+ind+"\"%s\": %s,\n", name, val)
				}
			}
		}
		contents += ind + "})"
		return contents
	}

	panic("no entry for type " + underlyingType.Name)
}

// GenerateTypeDef generates the given typedef.
func (g *Generator) GenerateTypeDef(*parser.TypeDef) error {
	// No typedefs in python
	return nil
}

// GenerateEnum generates the given enum.
func (g *Generator) GenerateEnum(enum *parser.Enum) error {
	contents := ""
	contents += fmt.Sprintf("class %s:\n", enum.Name)
	if enum.Comment != nil {
		contents += g.GenerateBlockComment(enum.Comment, tab)
	}
	for _, value := range enum.Values {
		contents += fmt.Sprintf(tab+"%s = %d\n", value.Name, value.Value)
	}
	contents += "\n"

	contents += tab + "_VALUES_TO_NAMES = {\n"
	for _, value := range enum.Values {
		contents += fmt.Sprintf(tabtab+"%d: \"%s\",\n", value.Value, value.Name)
	}
	contents += tab + "}\n\n"

	contents += tab + "_NAMES_TO_VALUES = {\n"
	for _, value := range enum.Values {
		contents += fmt.Sprintf(tabtab+"\"%s\": %d,\n", value.Name, value.Value)
	}
	contents += tab + "}\n\n"

	_, err := g.typesFile.WriteString(contents)
	return err
}

// GenerateStruct generates the given struct.
func (g *Generator) GenerateStruct(*parser.Struct) error {
	return nil
}

// GenerateUnion generates the given union.
func (g *Generator) GenerateUnion(*parser.Struct) error {
	return nil
}

// GenerateException generates the given exception.
func (g *Generator) GenerateException(*parser.Struct) error {
	return nil
}

// GenerateServiceArgsResults generates the args and results objects for the
// given service.
func (g *Generator) GenerateServiceArgsResults(string, string, []*parser.Struct) error {
	return nil
}

// GetOutputDir returns the output directory for generated files.
func (g *Generator) GetOutputDir(dir string) string {
	if pkg, ok := g.Frugal.Thrift.Namespace(lang); ok {
		path := generator.GetPackageComponents(pkg)
		dir = filepath.Join(append([]string{dir}, path...)...)
	} else {
		dir = filepath.Join(dir, g.Frugal.Name)
	}
	return dir
}

// DefaultOutputDir returns the default output directory for generated files.
func (g *Generator) DefaultOutputDir() string {
	dir := defaultOutputDir
	if _, ok := g.Options["tornado"]; ok {
		dir += ".tornado"
	}
	return dir
}

// PostProcess is called after generating each file.
func (g *Generator) PostProcess(f *os.File) error { return nil }

// GenerateDependencies is a no-op.
func (g *Generator) GenerateDependencies(dir string) error {
	return nil
}

// GenerateFile generates the given FileType.
func (g *Generator) GenerateFile(name, outputDir string, fileType generator.FileType) (*os.File, error) {
	switch fileType {
	case generator.PublishFile:
		return g.CreateFile(fmt.Sprintf("f_%s_publisher", name), outputDir, lang, false)
	case generator.SubscribeFile:
		return g.CreateFile(fmt.Sprintf("f_%s_subscriber", name), outputDir, lang, false)
	case generator.CombinedServiceFile:
		return g.CreateFile(fmt.Sprintf("f_%s", name), outputDir, lang, false)
	case generator.ObjectFile:
		return g.CreateFile(fmt.Sprintf("f_%s", name), outputDir, lang, false)
	default:
		return nil, fmt.Errorf("Bad file type for Python generator: %s", fileType)
	}
}

// GenerateDocStringComment generates the autogenerated notice.
func (g *Generator) GenerateDocStringComment(file *os.File) error {
	comment := fmt.Sprintf(
		"#\n"+
			"# Autogenerated by Frugal Compiler (%s)\n"+
			"#\n"+
			"# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING\n"+
			"#",
		globals.Version)

	_, err := file.WriteString(comment)
	return err
}

func (g *Generator) GenerateBlockComment(comment []string, indent string) string {
	contents := indent + "\"\"\"\n"
	for _, line := range comment {
		contents += indent + line + "\n"
	}
	contents += indent + "\"\"\"\n"
	return contents
}

// GenerateServicePackage is a no-op.
func (g *Generator) GenerateServicePackage(file *os.File, s *parser.Service) error {
	return nil
}

// GenerateScopePackage is a no-op.
func (g *Generator) GenerateScopePackage(file *os.File, s *parser.Scope) error {
	return nil
}

func (g *Generator) GenerateTypesImports(file *os.File) error {
	contents := ""
	contents += "from thrift.Thrift import TType, TMessageType, TException, TApplicationException\n"
	for _, include := range g.Frugal.Thrift.Includes {
		includeName, ok := g.Frugal.NamespaceForInclude(filepath.Base(include.Name), lang)
		if !ok {
			includeName = include.Name
		}
		contents += fmt.Sprintf("import %s.f_types\n", includeName)
	}
	contents += "\n\n"
	contents += "from thrift.transport import TTransport\n"
	contents += "from thrift.protocol import TBinaryProtocol, TProtocol\n"
	contents += "try:\n"
	contents += tab + "from thrift.protocol import fastbinary\n"
	contents += "except:\n"
	contents += tab + "fastbinary = None\n"

	_, err := file.WriteString(contents)
	return err
}

// GenerateServiceImports generates necessary imports for the given service.
func (g *Generator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	// TODO
	return nil
}

// GenerateScopeImports generates necessary imports for the given scope.
func (g *Generator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "from thrift.Thrift import TMessageType\n"
	imports += "from frugal.middleware import Method\n"
	_, err := file.WriteString(imports)
	return err
}

// GenerateConstants generates any static constants.
func (g *Generator) GenerateConstants(file *os.File, name string) error {
	return nil
}

// GeneratePublisher generates the publisher for the given scope.
func (g *Generator) GeneratePublisher(file *os.File, scope *parser.Scope) error {
	publisher := ""
	publisher += fmt.Sprintf("class %sPublisher(object):\n", scope.Name)
	if scope.Comment != nil {
		publisher += g.generateDocString(scope.Comment, tab)
	}
	publisher += "\n"

	publisher += tab + fmt.Sprintf("_DELIMITER = '%s'\n\n", globals.TopicDelimiter)

	publisher += tab + "def __init__(self, provider, middleware=None):\n"
	publisher += g.generateDocString([]string{
		fmt.Sprintf("Create a new %sPublisher.\n", scope.Name),
		"Args:",
		tab + "provider: FScopeProvider",
		tab + "middleware: ServiceMiddleware or list of ServiceMiddleware",
	}, tabtab)
	publisher += "\n"

	publisher += tabtab + "if middleware and not isinstance(middleware, list):\n"
	publisher += tabtabtab + "middleware = [middleware]\n"
	publisher += tabtab + "self._transport, protocol_factory = provider.new()\n"
	publisher += tabtab + "self._protocol = protocol_factory.get_protocol(self._transport)\n"
	publisher += tabtab + "self._methods = {\n"
	for _, op := range scope.Operations {
		publisher += tabtabtab + fmt.Sprintf("'publish_%s': Method(self._publish_%s, middleware),\n", op.Name, op.Name)
	}
	publisher += tabtab + "}\n\n"

	if _, ok := g.Options["tornado"]; ok {
		publisher += tab + "@gen.coroutine\n"
	}
	publisher += tab + "def open(self):\n"
	publisher += tabtab
	if _, ok := g.Options["tornado"]; ok {
		publisher += "yield "
	}
	publisher += "self._transport.open()\n\n"

	if _, ok := g.Options["tornado"]; ok {
		publisher += tab + "@gen.coroutine\n"
	}
	publisher += tab + "def close(self):\n"
	publisher += tabtab
	if _, ok := g.Options["tornado"]; ok {
		publisher += "yield "
	}
	publisher += "self._transport.close()\n\n"

	prefix := ""
	for _, op := range scope.Operations {
		publisher += prefix + g.generatePublishMethod(scope, op)
		prefix = "\n\n"
	}

	_, err := file.WriteString(publisher)
	return err
}

func (g *Generator) generatePublishMethod(scope *parser.Scope, op *parser.Operation) string {
	args := ""
	docstr := []string{"Args:", tab + "ctx: FContext"}
	if len(scope.Prefix.Variables) > 0 {
		prefix := ""
		for _, variable := range scope.Prefix.Variables {
			docstr = append(docstr, tab+fmt.Sprintf("%s: string", variable))
			args += prefix + variable
			prefix = ", "
		}
		args += ", "
	}
	docstr = append(docstr, tab+fmt.Sprintf("req: %s", op.Type.Name))
	if op.Comment != nil {
		docstr[0] = "\n" + tabtab + docstr[0]
		docstr = append(op.Comment, docstr...)
	}
	method := tab + fmt.Sprintf("def publish_%s(self, ctx, %sreq):\n", op.Name, args)
	method += g.generateDocString(docstr, tabtab)
	method += tabtab + fmt.Sprintf("self._methods['publish_%s']([ctx, %sreq])\n\n", op.Name, args)

	method += tab + fmt.Sprintf("def _publish_%s(self, ctx, %sreq):\n", op.Name, args)
	method += tabtab + fmt.Sprintf("op = '%s'\n", op.Name)
	method += tabtab + fmt.Sprintf("prefix = %s\n", generatePrefixStringTemplate(scope))
	method += tabtab + fmt.Sprintf("topic = '{}%s{}{}'.format(prefix, self._DELIMITER, op)\n", scope.Name)
	method += tabtab + "oprot = self._protocol\n"
	method += tabtab + "self._transport.lock_topic(topic)\n"
	method += tabtab + "try:\n"
	method += tabtabtab + "oprot.write_request_headers(ctx)\n"
	method += tabtabtab + "oprot.writeMessageBegin(op, TMessageType.CALL, 0)\n"
	method += tabtabtab + "req.write(oprot)\n"
	method += tabtabtab + "oprot.writeMessageEnd()\n"
	method += tabtabtab + "oprot.get_transport().flush()\n"
	method += tabtab + "finally:\n"
	method += tabtabtab + "self._transport.unlock_topic()\n"
	return method
}

func generatePrefixStringTemplate(scope *parser.Scope) string {
	if len(scope.Prefix.Variables) == 0 {
		if scope.Prefix.String == "" {
			return "''"
		}
		return fmt.Sprintf("'%s%s'", scope.Prefix.String, globals.TopicDelimiter)
	}
	template := fmt.Sprintf("'%s%s'.format(", scope.Prefix.Template("{}"), globals.TopicDelimiter)
	prefix := ""
	for _, variable := range scope.Prefix.Variables {
		template += prefix + variable
		prefix = ", "
	}
	template += ")"
	return template
}

// GenerateSubscriber generates the subscriber for the given scope.
func (g *Generator) GenerateSubscriber(file *os.File, scope *parser.Scope) error {
	// TODO
	globals.PrintWarning(fmt.Sprintf("%s: scope subscriber generation is not implemented for Python", scope.Name))
	return nil
}

// GenerateService generates the given service.
func (g *Generator) GenerateService(file *os.File, s *parser.Service) error {
	// TODO
	globals.PrintWarning(fmt.Sprintf("%s: service generation is not implemented for Python", s.Name))
	return nil
}

func (g *Generator) generateServiceInterface(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += fmt.Sprintf("class Iface(%s.Iface):\n", g.getServiceExtendsName(service))
	} else {
		contents += "class Iface(object):\n"
	}
	if service.Comment != nil {
		contents += g.generateDocString(service.Comment, tab)
	}
	contents += "\n"

	for _, method := range service.Methods {
		contents += g.generateMethodSignature(method)
		contents += tabtab + "pass\n\n"
	}

	return contents
}

func (g *Generator) getServiceExtendsName(service *parser.Service) string {
	serviceName := "f_" + service.ExtendsService()
	include := service.ExtendsInclude()
	if include != "" {
		if inc, ok := g.Frugal.NamespaceForInclude(include, lang); ok {
			include = inc
		}
		serviceName = include + "." + serviceName
	}
	return serviceName
}

func (g *Generator) generateProcessor(service *parser.Service) string {
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

func (g *Generator) generateMethodSignature(method *parser.Method) string {
	contents := ""
	docstr := []string{"Args:", tab + "ctx: FContext"}
	for _, arg := range method.Arguments {
		docstr = append(docstr, tab+fmt.Sprintf("%s: %s", arg.Name, g.getPythonTypeName(arg.Type)))
	}
	if method.Comment != nil {
		docstr[0] = "\n" + tabtab + docstr[0]
		docstr = append(method.Comment, docstr...)
	}
	contents += tab + fmt.Sprintf("def %s(self, ctx%s):\n", method.Name, g.generateClientArgs(method.Arguments))
	contents += g.generateDocString(docstr, tabtab)
	return contents
}

func (g *Generator) generateClientArgs(args []*parser.Field) string {
	return g.generateArgs(args, "")
}

func (g *Generator) generateServerArgs(args []*parser.Field) string {
	return g.generateArgs(args, "args.")
}

func (g *Generator) generateArgs(args []*parser.Field, prefix string) string {
	argsStr := ""
	for _, arg := range args {
		argsStr += fmt.Sprintf(", %s%s", prefix, arg.Name)
	}
	return argsStr
}

func (g *Generator) generateDocString(lines []string, tab string) string {
	docstr := tab + "\"\"\"\n"
	for _, line := range lines {
		docstr += tab + line + "\n"
	}
	docstr += tab + "\"\"\"\n"
	return docstr
}

func (g *Generator) getPythonTypeName(t *parser.Type) string {
	t = g.Frugal.UnderlyingType(t)
	switch t.Name {
	case "bool":
		return "boolean"
	case "byte", "i8":
		return "int (signed 8 bits)"
	case "i16":
		return "int (signed 16 bits)"
	case "i32":
		return "int (signed 32 bits)"
	case "i64":
		return "int (signed 64 bits)"
	case "double":
		return "float"
	case "string":
		return "string"
	case "binary":
		return "binary string"
	case "list":
		typ := g.Frugal.UnderlyingType(t.ValueType)
		return fmt.Sprintf("list of %s", g.getPythonTypeName(typ))
	case "set":
		typ := g.Frugal.UnderlyingType(t.ValueType)
		return fmt.Sprintf("set of %s", g.getPythonTypeName(typ))
	case "map":
		return fmt.Sprintf("dict of <%s, %s>",
			g.getPythonTypeName(t.KeyType), g.getPythonTypeName(t.ValueType))
	default:
		// Custom type, either typedef or struct.
		return t.Name
	}
}
