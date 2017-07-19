package python

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	tabtabtabtabtab  = tab + tab + tab + tab + tab
)

type concurrencyModel int

const (
	synchronous concurrencyModel = iota
	tornado
	asyncio
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
	switch getAsyncOpt(options) {
	case tornado:
		return &TornadoGenerator{gen}
	case asyncio:
		return &AsyncIOGenerator{gen}
	}
	return gen
}

// SetupGenerator performs any setup logic before generation.
func (g *Generator) SetupGenerator(outputDir string) error {
	g.outputDir = outputDir

	dir := g.outputDir
	for filepath.Dir(dir) != "." {
		file, err := g.GenerateFile("__init__", dir, generator.ObjectFile)
		file.Close()
		if err != nil {
			return err
		}

		dir = filepath.Dir(dir)
	}

	// create types file
	typesFile, err := g.GenerateFile("ttypes", outputDir, generator.ObjectFile)
	if err != nil {
		return err
	}
	if err = g.GenerateDocStringComment(typesFile); err != nil {
		return err
	}
	if _, err = typesFile.WriteString("\n\n"); err != nil {
		return err
	}
	if err = g.GenerateTypesImports(typesFile, false); err != nil {
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
	file, err := g.GenerateFile("constants", g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	contents := "\n\n"
	contents += "from thrift.Thrift import TType, TMessageType, TException, TApplicationException\n"
	contents += "from .ttypes import *\n\n"

	for _, include := range g.Frugal.Includes {
		namespace := g.getPackageNamespace(filepath.Base(include.Name))
		contents += fmt.Sprintf("import %s.ttypes\n", namespace)
		contents += fmt.Sprintf("import %s.constants\n", namespace)
		contents += "\n"
	}

	for _, constant := range constants {
		_, value := g.generateConstantValue(constant.Type, constant.Value, "")
		contents += fmt.Sprintf("%s = %s\n", constant.Name, value)
	}

	if err = g.GenerateDocStringComment(file); err != nil {
		return err
	}
	_, err = file.WriteString(contents)
	return err
}

func (g *Generator) generateConstantValue(t *parser.Type, value interface{}, ind string) (parser.IdentifierType, string) {
	if value == nil {
		return parser.NonIdentifier, "None"
	}

	// If the value being referenced is of type Identifier, it's referencing
	// another constant.
	identifier, ok := value.(parser.Identifier)
	if ok {
		idCtx := g.Frugal.ContextFromIdentifier(identifier)
		switch idCtx.Type {
		case parser.LocalConstant:
			return idCtx.Type, idCtx.Constant.Name
		case parser.LocalEnum:
			return idCtx.Type, fmt.Sprintf("%s.%s", idCtx.Enum.Name, idCtx.EnumValue.Name)
		case parser.IncludeConstant:
			include := idCtx.Include.Name
			include = g.getPackageNamespace(include)
			return idCtx.Type, fmt.Sprintf("%s.constants.%s", include, idCtx.Constant.Name)
		case parser.IncludeEnum:
			include := idCtx.Include.Name
			include = g.getPackageNamespace(include)
			return idCtx.Type, fmt.Sprintf("%s.ttypes.%s.%s", include, idCtx.Enum.Name, idCtx.EnumValue.Name)
		default:
			panic(fmt.Sprintf("The Identifier %s has unexpected type %d", identifier, idCtx.Type))
		}
	}

	underlyingType := g.Frugal.UnderlyingType(t)
	if underlyingType.IsPrimitive() || underlyingType.IsContainer() {
		switch underlyingType.Name {
		case "bool":
			return parser.NonIdentifier, strings.Title(fmt.Sprintf("%v", value))
		case "i8", "byte", "i16", "i32", "i64", "double":
			return parser.NonIdentifier, fmt.Sprintf("%v", value)
		case "string", "binary":
			return parser.NonIdentifier, fmt.Sprintf("%s", strconv.Quote(value.(string)))
		case "list", "set":
			contents := ""
			if underlyingType.Name == "set" {
				contents += "set("
			}
			contents += "[\n"
			for _, v := range value.([]interface{}) {
				_, val := g.generateConstantValue(underlyingType.ValueType, v, ind+tab)
				contents += fmt.Sprintf(ind+tab+"%s,\n", val)
			}
			contents += ind + "]"
			if underlyingType.Name == "set" {
				contents += ")"
			}
			return parser.NonIdentifier, contents
		case "map":
			contents := "{\n"
			for _, pair := range value.([]parser.KeyValue) {
				_, key := g.generateConstantValue(underlyingType.KeyType, pair.Key, ind+tab)
				_, val := g.generateConstantValue(underlyingType.ValueType, pair.Value, ind+tab)
				contents += fmt.Sprintf(ind+tab+"%s: %s,\n", key, val)
			}
			contents += ind + "}"
			return parser.NonIdentifier, contents
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return parser.NonIdentifier, fmt.Sprintf("%d", value)
	} else if g.Frugal.IsStruct(underlyingType) {
		s := g.Frugal.FindStruct(underlyingType)
		if s == nil {
			panic("no struct for type " + underlyingType.Name)
		}

		contents := ""

		contents += fmt.Sprintf("%s(**{\n", g.qualifiedTypeName(underlyingType))
		for _, pair := range value.([]parser.KeyValue) {
			name := pair.KeyToString()
			for _, field := range s.Fields {
				if name == field.Name {
					_, val := g.generateConstantValue(field.Type, pair.Value, ind+tab)
					contents += fmt.Sprintf(tab+ind+"\"%s\": %s,\n", name, val)
				}
			}
		}
		contents += ind + "})"
		return parser.NonIdentifier, contents
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
	contents += fmt.Sprintf("class %s(int):\n", enum.Name)
	if enum.Comment != nil {
		contents += g.generateDocString(enum.Comment, tab)
	}
	for _, value := range enum.Values {
		if value.Comment != nil {
			contents += g.generateDocString(value.Comment, tab)
		}
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
func (g *Generator) GenerateStruct(s *parser.Struct) error {
	_, err := g.typesFile.WriteString(g.generateStruct(s))
	return err
}

// GenerateUnion generates the given union.
func (g *Generator) GenerateUnion(union *parser.Struct) error {
	// TODO 2.0 consider adding validation only one field is set,
	// similar to other languages
	_, err := g.typesFile.WriteString(g.generateStruct(union))
	return err
}

// GenerateException generates the given exception.
func (g *Generator) GenerateException(exception *parser.Struct) error {
	_, err := g.typesFile.WriteString(g.generateStruct(exception))
	return err
}

// generateServiceArgsResults generates the args and results objects for the
// given service.
func (g *Generator) generateServiceArgsResults(service *parser.Service) string {
	contents := ""
	for _, s := range g.GetServiceMethodTypes(service) {
		contents += g.generateStruct(s)
	}
	return contents
}

// generateStruct generates a python representation of a thrift struct
func (g *Generator) generateStruct(s *parser.Struct) string {
	contents := ""

	extends := "(object)"
	if s.Type == parser.StructTypeException {
		extends = "(TException)"
	}
	contents += fmt.Sprintf("class %s%s:\n", s.Name, extends)
	contents += g.generateClassDocstring(s)

	contents += g.generateDefaultMarkers(s)
	contents += g.generateInit(s)

	contents += g.generateRead(s)
	contents += g.generateWrite(s)

	contents += g.generateValidate(s)
	contents += g.generateMagicMethods(s)

	return contents
}

// generateDefaultMarkers generates marker objects to provide as defaults to
// an __init__ method. The __init__ method can then determine if the default
// was provided and generate the constant appropriately. Don't generate the
// constant as a class variable to avoid order of declaration issues.
func (g *Generator) generateDefaultMarkers(s *parser.Struct) string {
	contents := ""
	importConstants := false
	for _, field := range s.Fields {
		if field.Default != nil {
			underlyingType := g.Frugal.UnderlyingType(field.Type)
			// use 'object()' as a marker value to avoid instantiating
			// a class defined later in the file
			defaultVal := "object()"
			idType := parser.NonIdentifier
			if underlyingType.IsPrimitive() || g.Frugal.IsEnum(underlyingType) {
				idType, defaultVal = g.generateConstantValue(underlyingType, field.Default, tab)
				if idType == parser.LocalConstant {
					importConstants = true
					defaultVal = fmt.Sprintf("constants.%s", defaultVal)
				}
			}
			contents += fmt.Sprintf(tab+"_DEFAULT_%s_MARKER = %s\n", field.Name, defaultVal)
		}
	}
	if importConstants {
		contents = tab + "from . import constants\n" + contents
	}
	return contents
}

// generateInit generates the init method for a class.
func (g *Generator) generateInit(s *parser.Struct) string {
	if len(s.Fields) == 0 {
		return ""
	}

	contents := ""
	argList := ""
	for _, field := range s.Fields {
		defaultVal := "None"
		if field.Default != nil {
			defaultVal = fmt.Sprintf("_DEFAULT_%s_MARKER", field.Name)
		}
		argList += fmt.Sprintf(", %s=%s", field.Name, defaultVal)
	}
	contents += fmt.Sprintf(tab+"def __init__(self%s):\n", argList)
	fieldContents := ""
	importConstants := false
	for _, field := range s.Fields {
		underlyingType := g.Frugal.UnderlyingType(field.Type)
		if !underlyingType.IsPrimitive() && !g.Frugal.IsEnum(underlyingType) && field.Default != nil {
			fieldContents += fmt.Sprintf(tabtab+"if %s is self._DEFAULT_%s_MARKER:\n", field.Name, field.Name)
			idType, val := g.generateConstantValue(field.Type, field.Default, tabtabtab)
			if idType == parser.LocalConstant {
				importConstants = true
				val = fmt.Sprintf("constants.%s", val)
			}
			fieldContents += fmt.Sprintf(tabtabtab+"%s = %s\n", field.Name, val)
		}
		fieldContents += fmt.Sprintf(tabtab+"self.%s = %s\n", field.Name, field.Name)
	}
	if importConstants {
		contents += tabtab + "from . import constants\n"
	}
	contents += fieldContents
	contents += "\n"
	return contents
}

// generateClassDocstring generates a docstring for class. This includes a
// description of the class, if present, a list of attributes, and descriptions
// of each attribute, if present.
func (g *Generator) generateClassDocstring(s *parser.Struct) string {
	lines := []string{}
	if s.Comment != nil {
		lines = append(lines, s.Comment...)
		lines = append(lines, "")
	}

	if len(s.Fields) > 0 {
		lines = append(lines, "Attributes:")
		for _, field := range s.Fields {
			line := fmt.Sprintf(" - %s", field.Name)
			if len(field.Comment) > 0 {
				line = fmt.Sprintf("%s: %s", line, field.Comment[0])
				lines = append(lines, line)
				lines = append(lines, field.Comment[1:]...)
			} else {
				lines = append(lines, line)
			}
		}
	}

	if len(lines) == 0 {
		return ""
	}

	return g.generateDocString(lines, tab)
}

// generateRead generates the read method for a struct.
func (g *Generator) generateRead(s *parser.Struct) string {
	contents := ""
	contents += tab + "def read(self, iprot):\n"
	contents += tabtab + "iprot.readStructBegin()\n"
	contents += tabtab + "while True:\n"
	contents += tabtabtab + "(fname, ftype, fid) = iprot.readFieldBegin()\n"
	contents += tabtabtab + "if ftype == TType.STOP:\n"
	contents += tabtabtabtab + "break\n"
	ifstatement := "if"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtab+"%s fid == %d:\n", ifstatement, field.ID)
		contents += fmt.Sprintf(tabtabtabtab+"if ftype == %s:\n", g.getTType(field.Type))
		contents += g.generateReadFieldRec(field, true, tabtabtabtabtab)
		contents += tabtabtabtab + "else:\n"
		contents += tabtabtabtabtab + "iprot.skip(ftype)\n"
		ifstatement = "elif"
	}
	contents += tabtabtab + "else:\n"
	contents += tabtabtabtab + "iprot.skip(ftype)\n"
	contents += tabtabtab + "iprot.readFieldEnd()\n"
	contents += tabtab + "iprot.readStructEnd()\n"
	contents += tabtab + "self.validate()\n\n"
	return contents
}

// generateWrite generates the write method for a struct.
func (g *Generator) generateWrite(s *parser.Struct) string {
	contents := ""
	contents += tab + "def write(self, oprot):\n"
	contents += tabtab + "self.validate()\n"
	contents += fmt.Sprintf(tabtab+"oprot.writeStructBegin('%s')\n", s.Name)
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"if self.%s is not None:\n", field.Name)
		contents += fmt.Sprintf(tabtabtab+"oprot.writeFieldBegin('%s', %s, %d)\n", field.Name, g.getTType(field.Type), field.ID)
		contents += g.generateWriteFieldRec(field, true, tabtabtab)
		contents += fmt.Sprintf(tabtabtab + "oprot.writeFieldEnd()\n")
	}

	contents += tabtab + "oprot.writeFieldStop()\n"
	contents += tabtab + "oprot.writeStructEnd()\n\n"
	return contents
}

// generateValidate generates a validate method for a class. This ensures
// required fields are present.
func (g *Generator) generateValidate(s *parser.Struct) string {
	contents := ""
	contents += tab + "def validate(self):\n"
	if s.Type != parser.StructTypeUnion {
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				contents += fmt.Sprintf(tabtab+"if self.%s is None:\n", field.Name)
				contents += fmt.Sprintf(tabtabtab+"raise TProtocol.TProtocolException(type=TProtocol.TProtocolException.INVALID_DATA, message='Required field %s is unset!')\n", field.Name)
			}
		}
	} else {
		contents += tabtab + "set_fields = 0\n"
		for _, field := range s.Fields {
			contents += fmt.Sprintf(tabtab+"if self.%s is not None:\n", field.Name)
			contents += tabtabtab + "set_fields += 1\n"
		}
		contents += tabtab + "if set_fields != 1:\n"
		contents += fmt.Sprintf(tabtabtab + "raise TProtocol.TProtocolException(type=TProtocol.TProtocolException.INVALID_DATA, message='The union did not have exactly one field set, {} were set'.format(set_fields))\n")
	}

	contents += tabtab + "return\n\n"
	return contents
}

// generateMagicMethods generates magic methods for the class, such as
// '__hash__', '__repr__', '__eq__', and '__ne__'.
func (g *Generator) generateMagicMethods(s *parser.Struct) string {
	contents := ""
	if s.Type == parser.StructTypeException {
		contents += tab + "def __str__(self):\n"
		contents += tabtab + "return repr(self)\n\n"
	}

	contents += tab + "def __hash__(self):\n"
	contents += tabtab + "value = 17\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"value = (value * 31) ^ hash(self.%s)\n", field.Name)
	}
	contents += tabtab + "return value\n\n"

	contents += tab + "def __repr__(self):\n"
	contents += tabtab + "L = ['%s=%r' % (key, value)\n"
	contents += tabtabtab + "for key, value in self.__dict__.items()]\n"
	contents += tabtab + "return '%s(%s)' % (self.__class__.__name__, ', '.join(L))\n\n"

	contents += tab + "def __eq__(self, other):\n"
	contents += tabtab + "return isinstance(other, self.__class__) and self.__dict__ == other.__dict__\n\n"

	contents += tab + "def __ne__(self, other):\n"
	contents += tabtab + "return not (self == other)\n\n"
	return contents
}

// generateSpecArgs is a recursive function that returns the type of the
// argument in the format thrift_spec requires.
func (g *Generator) generateSpecArgs(t *parser.Type) string {
	underlyingType := g.Frugal.UnderlyingType(t)

	if underlyingType.IsPrimitive() {
		return "None"
	} else if underlyingType.IsContainer() {
		switch underlyingType.Name {
		case "list", "set":
			return fmt.Sprintf("(%s, %s)", g.getTType(underlyingType.ValueType), g.generateSpecArgs(underlyingType.ValueType))
		case "map":
			return fmt.Sprintf("(%s, %s, %s, %s)",
				g.getTType(underlyingType.KeyType), g.generateSpecArgs(underlyingType.KeyType),
				g.getTType(underlyingType.ValueType), g.generateSpecArgs(underlyingType.ValueType))
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return "None"
	} else if g.Frugal.IsStruct(underlyingType) {
		qualifiedName := g.qualifiedTypeName(underlyingType)
		return fmt.Sprintf("(%s, %s.thrift_spec)", qualifiedName, qualifiedName)
	}

	panic("unrecognized type: " + t.Name)
}

// generateReadFieldRec recursively generates code to read a field.
func (g *Generator) generateReadFieldRec(field *parser.Field, first bool, ind string) string {
	contents := ""

	prefix := ""
	if first {
		prefix = "self."
	}
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isEnum := g.Frugal.IsEnum(underlyingType)
	if underlyingType.IsPrimitive() {
		thriftType := ""
		switch underlyingType.Name {
		case "bool", "byte", "i16", "i32", "i64", "double", "string", "binary":
			thriftType = strings.Title(underlyingType.Name)
		case "i8":
			thriftType = "Byte"
		default:
			panic("unknown type: " + underlyingType.Name)
		}
		contents += fmt.Sprintf(ind+"%s%s = iprot.read%s()\n", prefix, field.Name, thriftType)
	} else if isEnum {
		contents += fmt.Sprintf(ind+"%s%s = %s(iprot.readI32())\n", prefix, field.Name, g.qualifiedTypeName(underlyingType))
	} else if g.Frugal.IsStruct(underlyingType) {
		g.qualifiedTypeName(underlyingType)
		contents += fmt.Sprintf(ind+"%s%s = %s()\n", prefix, field.Name, g.qualifiedTypeName(underlyingType))
		contents += fmt.Sprintf(ind+"%s%s.read(iprot)\n", prefix, field.Name)
	} else if underlyingType.IsContainer() {
		sizeElem := g.GetElem()
		valElem := g.GetElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)

		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf(ind+"%s%s = []\n", prefix, field.Name)
			contents += fmt.Sprintf(ind+"(_, %s) = iprot.readListBegin()\n", sizeElem)
			contents += fmt.Sprintf(ind+"for _ in range(%s):\n", sizeElem)
			contents += g.generateReadFieldRec(valField, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s.append(%s)\n", prefix, field.Name, valElem)
			contents += fmt.Sprintf(ind + "iprot.readListEnd()\n")
		case "set":
			contents += fmt.Sprintf(ind+"%s%s = set()\n", prefix, field.Name)
			contents += fmt.Sprintf(ind+"(_, %s) = iprot.readSetBegin()\n", sizeElem)
			contents += fmt.Sprintf(ind+"for _ in range(%s):\n", sizeElem)
			contents += g.generateReadFieldRec(valField, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s)\n", prefix, field.Name, valElem)
			contents += fmt.Sprintf(ind + "iprot.readSetEnd()\n")
		case "map":
			contents += fmt.Sprintf(ind+"%s%s = {}\n", prefix, field.Name)
			contents += fmt.Sprintf(ind+"(_, _, %s) = iprot.readMapBegin()\n", sizeElem)
			contents += fmt.Sprintf(ind+"for _ in range(%s):\n", sizeElem)
			keyElem := g.GetElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			contents += g.generateReadFieldRec(keyField, false, ind+tab)
			contents += g.generateReadFieldRec(valField, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s[%s] = %s\n", prefix, field.Name, keyElem, valElem)
			contents += fmt.Sprintf(ind + "iprot.readMapEnd()\n")
		}
	}

	return contents
}

// generateReadFieldRec recursively generates code to write a field.
func (g *Generator) generateWriteFieldRec(field *parser.Field, first bool, ind string) string {
	contents := ""

	prefix := ""
	if first {
		prefix = "self."
	}
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isEnum := g.Frugal.IsEnum(underlyingType)
	if underlyingType.IsPrimitive() || isEnum {
		thriftType := ""
		switch underlyingType.Name {
		case "bool", "byte", "i16", "i32", "i64", "double", "string", "binary":
			thriftType = strings.Title(underlyingType.Name)
		case "i8":
			thriftType = "Byte"
		default:
			if isEnum {
				thriftType = "I32"
			} else {
				panic("unknown type: " + underlyingType.Name)
			}
		}
		contents += fmt.Sprintf(ind+"oprot.write%s(%s%s)\n", thriftType, prefix, field.Name)
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(ind+"%s%s.write(oprot)\n", prefix, field.Name)
	} else if underlyingType.IsContainer() {
		valElem := g.GetElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)
		valTType := g.getTType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf(ind+"oprot.writeListBegin(%s, len(%s%s))\n", valTType, prefix, field.Name)
			contents += fmt.Sprintf(ind+"for %s in %s%s:\n", valElem, prefix, field.Name)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += ind + "oprot.writeListEnd()\n"
		case "set":
			contents += fmt.Sprintf(ind+"oprot.writeSetBegin(%s, len(%s%s))\n", valTType, prefix, field.Name)
			contents += fmt.Sprintf(ind+"for %s in %s%s:\n", valElem, prefix, field.Name)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += ind + "oprot.writeSetEnd()\n"
		case "map":
			keyElem := g.GetElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			keyTType := g.getTType(underlyingType.KeyType)
			contents += fmt.Sprintf(ind+"oprot.writeMapBegin(%s, %s, len(%s%s))\n", keyTType, valTType, prefix, field.Name)
			contents += fmt.Sprintf(ind+"for %s, %s in %s%s.items():\n", keyElem, valElem, prefix, field.Name)
			contents += g.generateWriteFieldRec(keyField, false, ind+tab)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += ind + "oprot.writeMapEnd()\n"
		}
	}

	return contents
}

// GetOutputDir returns the output directory for generated files.
func (g *Generator) GetOutputDir(dir string) string {
	if namespace := g.Frugal.Namespace(lang); namespace != nil {
		path := generator.GetPackageComponents(namespace.Value)
		dir = filepath.Join(append([]string{dir}, path...)...)
	} else {
		dir = filepath.Join(dir, g.Frugal.Name)
	}
	return dir
}

// DefaultOutputDir returns the default output directory for generated files.
func (g *Generator) DefaultOutputDir() string {
	dir := defaultOutputDir
	switch getAsyncOpt(g.Options) {
	case tornado:
		dir += ".tornado"
	case asyncio:
		dir += ".asyncio"
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
		return g.CreateFile(fmt.Sprintf("%s", name), outputDir, lang, false)
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

// GenerateServicePackage is a no-op.
func (g *Generator) GenerateServicePackage(file *os.File, s *parser.Service) error {
	return nil
}

// GenerateScopePackage is a no-op.
func (g *Generator) GenerateScopePackage(file *os.File, s *parser.Scope) error {
	return nil
}

func (g *Generator) GenerateTypesImports(file *os.File, isArgsOrResult bool) error {
	contents := ""
	contents += "from thrift.Thrift import TType, TMessageType, TException, TApplicationException\n"
	for _, include := range g.Frugal.Includes {
		includeName := g.getPackageNamespace(filepath.Base(include.Name))
		contents += fmt.Sprintf("import %s.ttypes\n", includeName)
		contents += fmt.Sprintf("import %s.constants\n", includeName)
	}
	contents += "\n"
	if isArgsOrResult {
		contents += "from .ttypes import *\n"
	}
	contents += "from thrift.transport import TTransport\n"
	contents += "from thrift.protocol import TBinaryProtocol, TProtocol\n"

	_, err := file.WriteString(contents)
	return err
}

// GenerateServiceImports generates necessary imports for the given service.
func (g *Generator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	imports := "from threading import Lock\n\n"

	imports += "from frugal.middleware import Method\n"
	imports += "from frugal.exceptions import TApplicationExceptionType\n"
	imports += "from frugal.exceptions import TTransportExceptionType\n"
	imports += "from frugal.processor import FBaseProcessor\n"
	imports += "from frugal.processor import FProcessorFunction\n"
	imports += "from frugal.util.deprecate import deprecated\n"
	imports += "from thrift.Thrift import TApplicationException\n"
	imports += "from thrift.Thrift import TMessageType\n"
	imports += "from thrift.transport.TTransport import TTransportException\n\n"

	imports += g.generateServiceExtendsImport(s)
	if imp, err := g.generateServiceIncludeImports(s); err != nil {
		return err
	} else {
		imports += imp
	}

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) generateServiceExtendsImport(s *parser.Service) string {
	if s.Extends == "" {
		// No super service
		return ""
	}

	if strings.Contains(s.Extends, ".") {
		// From an include
		extendsSlice := strings.SplitN(s.Extends, ".", 2)
		namespace := g.getPackageNamespace(extendsSlice[0])
		return fmt.Sprintf("import %s.f_%s\n", namespace, extendsSlice[1])
	}

	// From the same file
	return fmt.Sprintf("from . import f_%s\n", s.Extends)
}

func (g *Generator) generateServiceIncludeImports(s *parser.Service) (string, error) {
	imports := ""

	// Import include modules.
	includes, err := s.ReferencedIncludes()
	if err != nil {
		return "", err
	}
	for _, include := range includes {
		namespace := g.getPackageNamespace(filepath.Base(include.Name))
		imports += fmt.Sprintf("import %s.ttypes\n", namespace)
		imports += fmt.Sprintf("import %s.constants\n", namespace)
	}

	// Import this service's modules.
	imports += "from .ttypes import *\n"

	return imports, nil
}

// GenerateScopeImports generates necessary imports for the given scope.
func (g *Generator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "from thrift.Thrift import TMessageType\n"
	imports += "from frugal.middleware import Method\n"
	imports += "from frugal.transport import TMemoryOutputBuffer\n"
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

	publisher += tabtab + "middleware = middleware or []\n"
	publisher += tabtab + "if middleware and not isinstance(middleware, list):\n"
	publisher += tabtabtab + "middleware = [middleware]\n"
	publisher += tabtab + "middleware += provider.get_middleware()\n"
	publisher += tabtab + "self._transport, self._protocol_factory = provider.new_publisher()\n"
	publisher += tabtab + "self._methods = {\n"
	for _, op := range scope.Operations {
		publisher += tabtabtab + fmt.Sprintf("'publish_%s': Method(self._publish_%s, middleware),\n", op.Name, op.Name)
	}
	publisher += tabtab + "}\n\n"

	asyncOpt := getAsyncOpt(g.Options)
	publisher += tab
	switch asyncOpt {
	case tornado:
		publisher += "@gen.coroutine\n" + tab
	case asyncio:
		publisher += "async "
	}
	publisher += "def open(self):\n"

	publisher += tabtab
	switch asyncOpt {
	case tornado:
		publisher += "yield "
	case asyncio:
		publisher += "await "
	}
	publisher += "self._transport.open()\n\n"

	publisher += tab
	switch asyncOpt {
	case tornado:
		publisher += "@gen.coroutine\n" + tab
	case asyncio:
		publisher += "async "
	}
	publisher += "def close(self):\n"

	publisher += tabtab
	switch asyncOpt {
	case tornado:
		publisher += "yield "
	case asyncio:
		publisher += "await "
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
	asyncOpt := getAsyncOpt(g.Options)
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

	method := tab
	switch asyncOpt {
	case tornado:
		method += "@gen.coroutine\n" + tab
	case asyncio:
		method += "async "
	}
	method += fmt.Sprintf("def publish_%s(self, ctx, %sreq):\n", op.Name, args)
	method += g.generateDocString(docstr, tabtab)
	method += tabtab
	switch asyncOpt {
	case tornado:
		method += "yield "
	case asyncio:
		method += "await "
	}
	method += fmt.Sprintf("self._methods['publish_%s']([ctx, %sreq])\n\n", op.Name, args)

	method += tab
	switch asyncOpt {
	case tornado:
		method += "@gen.coroutine\n" + tab
	case asyncio:
		method += "async "
	}
	method += fmt.Sprintf("def _publish_%s(self, ctx, %sreq):\n", op.Name, args)
	// Inject the prefix variables into the FContext to send
	for _, prefixVar := range scope.Prefix.Variables {
		method += fmt.Sprintf(tabtab+"ctx.set_request_header('_topic_%s', %s)\n", prefixVar, prefixVar)
	}
	method += tabtab + fmt.Sprintf("op = '%s'\n", op.Name)
	method += tabtab + fmt.Sprintf("prefix = %s\n", generatePrefixStringTemplate(scope))
	method += tabtab + fmt.Sprintf("topic = '{}%s{}{}'.format(prefix, self._DELIMITER, op)\n", scope.Name)
	method += tabtab + "buffer = TMemoryOutputBuffer(self._transport.get_publish_size_limit())\n"
	method += tabtab + "oprot = self._protocol_factory.get_protocol(buffer)\n"
	method += tabtab + "oprot.write_request_headers(ctx)\n"
	method += tabtab + "oprot.writeMessageBegin(op, TMessageType.CALL, 0)\n"
	method += g.generateWriteFieldRec(parser.FieldFromType(op.Type, "req"), false, tabtab)
	method += tabtab + "oprot.writeMessageEnd()\n"

	method += tabtab
	switch asyncOpt {
	case tornado:
		method += "yield "
	case asyncio:
		method += "await "
	}
	method += "self._transport.publish(topic, buffer.getvalue())\n"
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
	contents := ""
	contents += g.generateServiceInterface(s)
	contents += g.generateClient(s)
	contents += g.generateServer(s)
	contents += g.generateServiceArgsResults(s)

	_, err := file.WriteString(contents)
	return err
}

func (g *Generator) generateClient(service *parser.Service) string {
	contents := "\n"
	contents += g.generateClientConstructor(service, false)
	for _, method := range service.Methods {
		contents += g.generateClientMethod(method)
	}
	return contents
}

func (g *Generator) generateClientMethod(method *parser.Method) string {
	contents := ""
	contents += g.generateMethodSignature(method)
	contents += tabtab + fmt.Sprintf("return self._methods['%s']([ctx%s])\n\n",
		method.Name, g.generateClientArgs(method.Arguments))

	contents += tab + fmt.Sprintf("def _%s(self, ctx%s):\n", method.Name, g.generateClientArgs(method.Arguments))
	contents += tabtab + fmt.Sprintf("self._send_%s(ctx%s)\n", method.Name, g.generateClientArgs(method.Arguments))
	if method.Oneway {
		contents += "\n"
	} else {
		contents += tabtab
		if method.ReturnType != nil {
			contents += "return "
		}
		contents += fmt.Sprintf("self._recv_%s(ctx)\n\n", method.Name)
	}

	contents += g.generateClientSendMethod(method)
	contents += g.generateClientRecvMethod(method)

	return contents
}

func (g *Generator) generateClientSendMethod(method *parser.Method) string {
	contents := ""
	contents += tab + fmt.Sprintf("def _send_%s(self, ctx%s):\n", method.Name, g.generateClientArgs(method.Arguments))
	contents += tabtab + "oprot = self._oprot\n"
	contents += tabtab + "with self._write_lock:\n"
	contents += tabtabtab + "oprot.get_transport().set_timeout(ctx.timeout)\n"
	contents += tabtabtab + "oprot.write_request_headers(ctx)\n"
	contents += tabtabtab + fmt.Sprintf("oprot.writeMessageBegin('%s', TMessageType.CALL, 0)\n", parser.LowercaseFirstLetter(method.Name))
	contents += tabtabtab + fmt.Sprintf("args = %s_args()\n", method.Name)
	for _, arg := range method.Arguments {
		contents += tabtabtab + fmt.Sprintf("args.%s = %s\n", arg.Name, arg.Name)
	}
	contents += tabtabtab + "args.write(oprot)\n"
	contents += tabtabtab + "oprot.writeMessageEnd()\n"
	contents += tabtabtab + "oprot.get_transport().flush()\n\n"

	return contents
}

func (g *Generator) generateClientRecvMethod(method *parser.Method) string {
	contents := tab + fmt.Sprintf("def _recv_%s(self, ctx):\n", method.Name)
	contents += tabtab + "self._iprot.read_response_headers(ctx)\n"
	contents += tabtab + "_, mtype, _ = self._iprot.readMessageBegin()\n"
	contents += tabtab + "if mtype == TMessageType.EXCEPTION:\n"
	contents += tabtabtab + "x = TApplicationException()\n"
	contents += tabtabtab + "x.read(self._iprot)\n"
	contents += tabtabtab + "self._iprot.readMessageEnd()\n"
	contents += tabtabtab + "if x.type == TApplicationExceptionType.RESPONSE_TOO_LARGE:\n"
	contents += tabtabtabtab + "raise TTransportException(type=TTransportExceptionType.RESPONSE_TOO_LARGE, message=x.message)\n"
	contents += tabtabtab + "raise x\n"
	contents += tabtab + fmt.Sprintf("result = %s_result()\n", method.Name)
	contents += tabtab + "result.read(self._iprot)\n"
	contents += tabtab + "self._iprot.readMessageEnd()\n"
	for _, err := range method.Exceptions {
		contents += tabtab + fmt.Sprintf("if result.%s is not None:\n", err.Name)
		contents += tabtabtab + fmt.Sprintf("raise result.%s\n", err.Name)
	}
	if method.ReturnType == nil {
		contents += tabtab + "return\n\n"
		return contents
	}
	contents += tabtab + "if result.success is not None:\n"
	contents += tabtabtab + "return result.success\n"
	contents += tabtab + fmt.Sprintf(
		"x = TApplicationException(TApplicationExceptionType.MISSING_RESULT, \"%s failed: unknown result\")\n", method.Name)
	contents += tabtab + "raise x\n\n"

	return contents
}

func (g *Generator) generateClientConstructor(service *parser.Service, async bool) string {
	contents := ""
	if service.Extends != "" {
		contents += fmt.Sprintf("class Client(%s.Client, Iface):\n\n", g.getServiceExtendsName(service))
	} else {
		contents += "class Client(Iface):\n\n"
	}

	contents += tab + "def __init__(self, provider, middleware=None):\n"
	docstring := []string{
		"Create a new Client with an FServiceProvider containing a transport",
		"and protocol factory.\n",
		"Args:",
	}
	if async {
		docstring = append(docstring, tab+"provider: FServiceProvider")
	} else {
		docstring = append(docstring, tab+"provider: FServiceProvider with TSynchronousTransport")
	}
	docstring = append(docstring, tab+"middleware: ServiceMiddleware or list of ServiceMiddleware")
	contents += g.generateDocString(docstring, tabtab)
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
		contents += tabtab + "self._oprot = self._protocol_factory.get_protocol(self._transport)\n"
		if !async {
			contents += tabtab + "self._iprot = self._protocol_factory.get_protocol(self._transport)\n"
		}
		contents += tabtab + "self._write_lock = Lock()\n"
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
	return contents
}

func (g *Generator) generateServer(service *parser.Service) string {
	contents := ""
	contents += g.generateProcessor(service)
	for _, method := range service.Methods {
		contents += g.generateProcessorFunction(method)
	}
	contents += g.generateWriteApplicationException()

	return contents
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
		include := g.getPackageNamespace(include)
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

func (g *Generator) generateProcessorFunction(method *parser.Method) string {
	methodLower := parser.LowercaseFirstLetter(method.Name)
	contents := ""
	contents += fmt.Sprintf("class _%s(FProcessorFunction):\n\n", method.Name)
	contents += tab + "def __init__(self, handler, lock):\n"
	contents += tabtab + fmt.Sprintf("super(_%s, self).__init__(handler, lock)\n", method.Name)
	contents += "\n"

	if _, ok := method.Annotations.Deprecated(); ok {
		contents += tab + "@deprecated\n"
	}
	contents += tab + "def process(self, ctx, iprot, oprot):\n"
	contents += tabtab + fmt.Sprintf("args = %s_args()\n", method.Name)
	contents += tabtab + "args.read(iprot)\n"
	contents += tabtab + "iprot.readMessageEnd()\n"
	if !method.Oneway {
		contents += tabtab + fmt.Sprintf("result = %s_result()\n", method.Name)
	}
	contents += tabtab + "try:\n"
	if method.ReturnType == nil {
		contents += tabtabtab + fmt.Sprintf("self._handler([ctx%s])\n",
			g.generateServerArgs(method.Arguments))
	} else {
		contents += tabtabtab + fmt.Sprintf("result.success = self._handler([ctx%s])\n",
			g.generateServerArgs(method.Arguments))
	}
	for _, err := range method.Exceptions {
		contents += tabtab + fmt.Sprintf("except %s as %s:\n", g.qualifiedTypeName(err.Type), err.Name)
		contents += tabtabtab + fmt.Sprintf("result.%s = %s\n", err.Name, err.Name)
	}
	contents += tabtab + "except TApplicationException as ex:\n"
	contents += tabtabtab + "with self._lock:\n"
	contents += tabtabtabtab +
		fmt.Sprintf("_write_application_exception(ctx, oprot, \"%s\", exception=ex)\n",
			methodLower)
	contents += tabtabtabtab + "return\n"
	contents += tabtab + "except Exception as e:\n"
	if !method.Oneway {
		contents += tabtabtab + "with self._lock:\n"
		contents += tabtabtabtab + fmt.Sprintf("_write_application_exception(ctx, oprot, \"%s\", ex_code=TApplicationExceptionType.INTERNAL_ERROR, message=e.message)\n", methodLower)
	}
	contents += tabtabtab + "raise\n"
	if !method.Oneway {
		contents += tabtab + "with self._lock:\n"
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
			"raise _write_application_exception(ctx, oprot, \"%s\", ex_code=TApplicationExceptionType.RESPONSE_TOO_LARGE, message=e.args[0])\n", methodLower)
		contents += tabtabtabtab + "else:\n"
		contents += tabtabtabtabtab + "raise e\n"
	}
	contents += "\n\n"

	return contents
}

func (g *Generator) generateWriteApplicationException() string {
	contents := "def _write_application_exception(ctx, oprot, method, ex_code=None, message=None, exception=None):\n"
	contents += tab + "if exception is not None:\n"
	contents += tabtab + "x = exception\n"
	contents += tab + "else:\n"
	contents += tabtab + "x = TApplicationException(type=ex_code, message=message)\n"
	contents += tab + "oprot.write_response_headers(ctx)\n"
	contents += tab + "oprot.writeMessageBegin(method, TMessageType.EXCEPTION, 0)\n"
	contents += tab + "x.write(oprot)\n"
	contents += tab + "oprot.writeMessageEnd()\n"
	contents += tab + "oprot.get_transport().flush()\n"
	contents += tab + "return x"
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

	deprecationValue, deprecated := method.Annotations.Deprecated()
	if deprecationValue != "" && deprecated {
		docstr = append(docstr, "", fmt.Sprintf("deprecated: %s", deprecationValue))
	}

	if deprecated {
		contents += tab + "@deprecated\n"
	}

	contents += tab
	if getAsyncOpt(g.Options) == asyncio {
		contents += "async "
	}

	contents += fmt.Sprintf("def %s(self, ctx%s):\n", method.Name, g.generateClientArgs(method.Arguments))
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

func (g *Generator) qualifiedTypeName(t *parser.Type) string {
	param := t.ParamName()
	include := t.IncludeName()

	if include == "" {
		return param
	}

	namespace := g.getPackageNamespace(include)
	return fmt.Sprintf("%s.ttypes.%s", namespace, param)
}

func (g *Generator) getTType(t *parser.Type) string {
	underlyingType := g.Frugal.UnderlyingType(t)

	ttype := ""
	switch underlyingType.Name {
	case "bool", "byte", "double", "i16", "i32", "i64", "list", "set", "map", "string":
		ttype = strings.ToUpper(underlyingType.Name)
	case "binary":
		ttype = "STRING"
	default:
		if g.Frugal.IsStruct(t) {
			ttype = "STRUCT"
		} else if g.Frugal.IsEnum(t) {
			ttype = "I32"
		} else {
			panic("unrecognized type: " + underlyingType.Name)
		}
	}
	return "TType." + ttype
}

func (g *Generator) getPackageNamespace(include string) string {
	name := include
	if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
		name = namespace.Value
	}
	return g.Options["package_prefix"] + name
}

func getAsyncOpt(options map[string]string) concurrencyModel {
	if _, ok := options["tornado"]; ok {
		return tornado
	} else if _, ok := options["asyncio"]; ok {
		return asyncio
	}
	return synchronous
}
