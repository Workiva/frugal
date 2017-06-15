package dartlang

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	"gopkg.in/yaml.v2"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

const (
	lang                  = "dart"
	defaultOutputDir      = "gen-dart"
	serviceSuffix         = "_service"
	scopeSuffix           = "_scope"
	minimumDartVersion    = "1.13.0"
	tab                   = "  "
	tabtab                = tab + tab
	tabtabtab             = tab + tab + tab
	tabtabtabtab          = tab + tab + tab + tab
	tabtabtabtabtab       = tab + tab + tab + tab + tab
	tabtabtabtabtabtab    = tab + tab + tab + tab + tab + tab
	tabtabtabtabtabtabtab = tab + tab + tab + tab + tab + tab + tab
	libraryPrefixOption   = "library_prefix"
	useVendorOption       = "use_vendor"
)

// Generator implements the LanguageGenerator interface for Dart.
type Generator struct {
	*generator.BaseGenerator
	outputDir string
}

// NewGenerator creates a new Dart LanguageGenerator.
func NewGenerator(options map[string]string) generator.LanguageGenerator {
	return &Generator{
		BaseGenerator: &generator.BaseGenerator{Options: options},
		outputDir:     "",
	}
}

func (g *Generator) getLibraryName() string {
	if ns := g.Frugal.Namespace(lang); ns != nil {
		return parser.LowercaseFirstLetter(toLibraryName(ns.Value))
	}
	return parser.LowercaseFirstLetter(g.Frugal.Name)
}

// SetupGenerator performs any setup logic before generation.
func (g *Generator) SetupGenerator(outputDir string) error {
	g.outputDir = outputDir

	if g.getLibraryPrefix() == "" {
		libDir := filepath.Join(outputDir, "lib", "src")
		if err := os.MkdirAll(libDir, 0777); err != nil {
			return err
		}
	}

	libraryName := g.getLibraryName()
	file, err := os.Create(g.getExportFilePath(outputDir))
	if err != nil {
		return err
	}
	defer file.Close()

	if err := g.GenerateDocStringComment(file); err != nil {
		return err
	}

	contents := ""
	contents += fmt.Sprintf("\n\nlibrary %s;\n\n", libraryName)

	if len(g.Frugal.Constants) > 0 {
		constantsName := fmt.Sprintf("%sConstants", snakeToCamel(libraryName))
		contents += g.createExport(constantsName, false)
	}
	for _, s := range g.Frugal.Structs {
		contents += g.createExport(s.Name, false)
	}
	for _, union := range g.Frugal.Unions {
		contents += g.createExport(union.Name, false)
	}
	for _, exception := range g.Frugal.Exceptions {
		contents += g.createExport(exception.Name, false)
	}
	for _, enum := range g.Frugal.Enums {
		contents += g.createExport(enum.Name, true)
	}

	_, err = file.WriteString(contents)
	return err
}

func (g *Generator) createExport(structName string, isEnum bool) string {
	srcDir := "src"
	if _, ok := g.Options[libraryPrefixOption]; ok {
		srcDir = g.getLibraryName()
	}
	if !isEnum || !g.useEnums() {
		return fmt.Sprintf("export '%s/f_%s.dart' show %s;\n",
			srcDir, toFileName(structName), structName)
	}
	return fmt.Sprintf("export '%s/f_%s.dart' show %s, serialize%s, deserialize%s;\n",
		srcDir, toFileName(structName), structName, structName, structName)
}

// TeardownGenerator is run after generation.
func (g *Generator) TeardownGenerator() error { return nil }

// GetOutputDir returns the output directory for generated files.
func (g *Generator) GetOutputDir(dir string) string {
	if namespace := g.Frugal.Namespace(lang); namespace != nil {
		dir = filepath.Join(dir, toLibraryName(namespace.Value))
	} else {
		dir = filepath.Join(dir, g.Frugal.Name)
	}
	return dir
}

// DefaultOutputDir returns the default output directory for generated files.
func (g *Generator) DefaultOutputDir() string {
	return defaultOutputDir
}

// PostProcess is called after generating each file.
func (g *Generator) PostProcess(f *os.File) error { return nil }

// GenerateDependencies modifies the pubspec.yaml as needed.
func (g *Generator) GenerateDependencies(dir string) error {
	if _, ok := g.Options[libraryPrefixOption]; !ok {
		if err := g.addToPubspec(dir); err != nil {
			return err
		}
	}
	if err := g.exportClasses(dir); err != nil {
		return err
	}
	return nil
}

type pubspec struct {
	Name         string                      `yaml:"name"`
	Version      string                      `yaml:"version"`
	Description  string                      `yaml:"description"`
	Environment  env                         `yaml:"environment"`
	Dependencies map[interface{}]interface{} `yaml:"dependencies"`
}

type env struct {
	SDK string `yaml:"sdk"`
}

type dep struct {
	Hosted  hostedDep `yaml:"hosted,omitempty"`
	Git     gitDep    `yaml:"git,omitempty"`
	Path    string    `yaml:"path,omitempty"`
	Version string    `yaml:"version,omitempty"`
}

type hostedDep struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type gitDep struct {
	URL string `yaml:"url"`
	Ref string `yaml:"ref"`
}

func (g *Generator) addToPubspec(dir string) error {
	pubFilePath := filepath.Join(dir, "pubspec.yaml")

	deps := map[interface{}]interface{}{
		"logging": "^0.11.2",
		"thrift": dep{
			Hosted:  hostedDep{Name: "thrift", URL: "https://pub.workiva.org"},
			Version: "^0.0.7",
		},
	}

	if g.Frugal.ContainsFrugalDefinitions() {
		deps["frugal"] = dep{
			Hosted:  hostedDep{Name: "frugal", URL: "https://pub.workiva.org"},
			Version: fmt.Sprintf("^%s", globals.Version),
		}
	}

	includesSet := make(map[string]bool) // include.Name ---> include.Annotations.Vendor()

	scopeIncludes, err := g.Frugal.ReferencedScopeIncludes()
	if err != nil {
		return err
	}
	for _, include := range scopeIncludes {
		_, vendored := include.Annotations.Vendor()
		includesSet[include.Name] = vendored
	}
	servIncludes, err := g.Frugal.ReferencedServiceIncludes()
	if err != nil {
		return err
	}
	for _, include := range servIncludes {
		_, vendored := include.Annotations.Vendor()
		includesSet[include.Name] = vendored
	}
	includes := make([]string, 0, len(includesSet))
	for include := range includesSet {
		includes = append(includes, include)
	}
	sort.Strings(includes)

	for _, include := range includes {
		name := include
		namespace := g.Frugal.NamespaceForInclude(include, lang)
		if namespace != nil {
			name = namespace.Value
		}

		if g.useVendor() && includesSet[include] {
			vendorPath, _ := namespace.Annotations.Vendor()
			deps[toLibraryName(vendorPath)] = dep{
				Hosted:  hostedDep{Name: toLibraryName(vendorPath), URL: "https://pub.workiva.org"},
				Version: "any",
			}
		} else {
			deps[toLibraryName(name)] = dep{Path: "../" + toLibraryName(name)}
		}
	}

	namespace := g.Frugal.Namespace(lang)
	name := g.Frugal.Name
	if namespace != nil {
		name = namespace.Value
	}

	ps := &pubspec{
		Name:        toLibraryName(name),
		Version:     globals.Version,
		Description: "Autogenerated by the frugal compiler",
		Environment: env{
			SDK: "^" + minimumDartVersion,
		},
		Dependencies: deps,
	}

	d, err := yaml.Marshal(&ps)
	if err != nil {
		return err
	}
	// create and write to new file
	newPubFile, err := os.Create(pubFilePath)
	defer newPubFile.Close()
	if err != nil {
		return err
	}
	if _, err := newPubFile.Write(d); err != nil {
		return err
	}
	return nil
}

func (g *Generator) getExportFilePath(dir string) string {
	libName := g.getLibraryName()
	dartFile := fmt.Sprintf("%s.%s", libName, lang)

	if _, ok := g.Options[libraryPrefixOption]; ok {
		return filepath.Join(dir, "..", dartFile)
	}
	return filepath.Join(dir, "lib", dartFile)
}

func (g *Generator) exportClasses(dir string) error {
	filename := g.getLibraryName()
	mainFilePath := g.getExportFilePath(dir)
	mainFile, err := os.OpenFile(mainFilePath, syscall.O_RDWR, 0777)
	defer mainFile.Close()
	if err != nil {
		return err
	}

	exports := "\n"
	for _, service := range g.Frugal.Services {
		servSrcDir := "src"
		if _, ok := g.Options[libraryPrefixOption]; ok {
			servSrcDir = filename
		}

		servTitle := strings.Title(service.Name)
		exports += fmt.Sprintf("export '%s/%s%s%s.%s' show F%s;\n",
			servSrcDir, generator.FilePrefix, toFileName(service.Name), serviceSuffix, lang, servTitle)
		exports += fmt.Sprintf("export '%s/%s%s%s.%s' show F%sClient;\n",
			servSrcDir, generator.FilePrefix, toFileName(service.Name), serviceSuffix, lang, servTitle)
	}
	for _, scope := range g.Frugal.Scopes {
		scopeSrcDir := "src"
		if _, ok := g.Options[libraryPrefixOption]; ok {
			scopeSrcDir = filename
		}
		scopeTitle := strings.Title(scope.Name)
		exports += fmt.Sprintf("export '%s/%s%s%s.%s' show %sPublisher, %sSubscriber;\n",
			scopeSrcDir, generator.FilePrefix, toFileName(scope.Name), scopeSuffix, lang, scopeTitle, scopeTitle)
	}
	stat, err := mainFile.Stat()
	if err != nil {
		return err
	}
	_, err = mainFile.WriteAt([]byte(exports), stat.Size())
	return err
}

// GenerateFile generates the given FileType.
func (g *Generator) GenerateFile(name, outputDir string, fileType generator.FileType) (*os.File, error) {
	if _, ok := g.Options[libraryPrefixOption]; !ok {
		outputDir = filepath.Join(outputDir, "lib")
		outputDir = filepath.Join(outputDir, "src")
	}
	switch fileType {
	case generator.CombinedServiceFile:
		return g.CreateFile(toFileName(name)+serviceSuffix, outputDir, lang, true)
	case generator.CombinedScopeFile:
		return g.CreateFile(toFileName(name)+scopeSuffix, outputDir, lang, true)
	case generator.ObjectFile:
		file, err := g.CreateFile(toFileName(name), outputDir, lang, true)
		if err != nil {
			return file, err
		}
		if err = g.GenerateDocStringComment(file); err != nil {
			return file, err
		}
		if _, err = file.WriteString("\n\n"); err != nil {
			return file, err
		}
		return file, nil

	default:
		return nil, fmt.Errorf("Bad file type for dartlang generator: %s", fileType)
	}
}

// GenerateDocStringComment generates the autogenerated notice.
func (g *Generator) GenerateDocStringComment(file *os.File) error {
	comment := fmt.Sprintf(
		"// Autogenerated by Frugal Compiler (%s)\n"+
			"// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING",
		globals.Version)

	_, err := file.WriteString(comment)
	return err
}

// GenerateConstantsContents generates constants.
func (g *Generator) GenerateConstantsContents(constants []*parser.Constant) error {
	if len(constants) == 0 {
		return nil
	}

	className := fmt.Sprintf("%sConstants", snakeToCamel(g.getLibraryName()))
	file, err := g.GenerateFile(className, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.writeThriftImports(file); err != nil {
		return err
	}

	// Need utf8 for binary constants
	_, err = file.WriteString("import 'dart:convert' show UTF8;\n\n")
	if err != nil {
		return err
	}

	contents := ""
	contents += fmt.Sprintf("class %s {\n", className)
	for _, constant := range constants {
		if constant.Comment != nil {
			contents += g.GenerateInlineComment(constant.Comment, tab+"/")
		}
		value := g.generateConstantValue(constant.Type, constant.Value, "")
		contents += fmt.Sprintf(tab+"static final %s %s = %s;\n",
			g.getDartTypeFromThriftType(constant.Type), constant.Name, value)
	}
	contents += "}\n"
	_, err = file.WriteString(contents)
	return err
}

func (g *Generator) generateConstantValue(t *parser.Type, value interface{}, ind string) string {
	underlyingType := g.Frugal.UnderlyingType(t)

	// If the value being referenced is of type Identifier, it's referencing
	// another constant.
	identifier, ok := value.(parser.Identifier)
	if ok {
		idCtx := g.Frugal.ContextFromIdentifier(identifier)
		switch idCtx.Type {
		case parser.LocalConstant:
			return fmt.Sprintf("t_%s.%sConstants.%s", toLibraryName(g.getNamespaceOrName()),
				snakeToCamel(g.getLibraryName()), idCtx.Constant.Name)
		case parser.LocalEnum:
			return fmt.Sprintf("t_%s.%s.%s", toLibraryName(g.getNamespaceOrName()),
				idCtx.Enum.Name, idCtx.EnumValue.Name)
		case parser.IncludeConstant:
			include := idCtx.Include.Name
			if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
				include = namespace.Value
			}
			nsLibName := toLibraryName(include)
			return fmt.Sprintf("t_%s.%sConstants.%s", nsLibName, snakeToCamel(nsLibName), idCtx.Constant.Name)
		case parser.IncludeEnum:
			include := idCtx.Include.Name
			if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
				include = namespace.Value
			}
			return fmt.Sprintf("t_%s.%s.%s", toLibraryName(include), idCtx.Enum.Name, idCtx.EnumValue.Name)
		default:
			panic(fmt.Sprintf("The Identifier %s has unexpected type %d", identifier, idCtx.Type))
		}
	}

	if underlyingType.IsPrimitive() || underlyingType.IsContainer() {
		switch underlyingType.Name {
		case "bool", "i8", "byte", "i16", "i32", "i64", "double":
			return fmt.Sprintf("%v", value)
		case "string":
			return fmt.Sprintf("%s", strconv.Quote(value.(string)))
		case "binary":
			return fmt.Sprintf("new Uint8List.fromList(UTF8.encode('%s'))", value)
		case "list", "set":
			contents := ""
			if underlyingType.Name == "set" {
				contents += fmt.Sprintf("new Set<%s>.from(", underlyingType.ValueType)
			}
			contents += "[\n"
			for _, v := range value.([]interface{}) {
				val := g.generateConstantValue(underlyingType.ValueType, v, ind+tab)
				contents += fmt.Sprintf(tabtab+ind+"%s,\n", val)
			}
			contents += tab + ind + "]"
			if underlyingType.Name == "set" {
				contents += ")"
			}
			return contents
		case "map":
			contents := "{\n"
			for _, pair := range value.([]parser.KeyValue) {
				key := g.generateConstantValue(underlyingType.KeyType, pair.Key, ind+tab)
				val := g.generateConstantValue(underlyingType.ValueType, pair.Value, ind+tab)
				contents += fmt.Sprintf(tabtab+ind+"%s: %s,\n", key, val)
			}
			contents += tab + ind + "}"
			return contents
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return fmt.Sprintf("%d", value)
	} else if g.Frugal.IsStruct(underlyingType) {
		s := g.Frugal.FindStruct(underlyingType)
		if s == nil {
			panic("no struct for type " + underlyingType.Name)
		}

		contents := ""

		dartType := g.getDartTypeFromThriftType(parser.TypeFromStruct(s))
		contents += fmt.Sprintf("new %s()", dartType)
		for _, pair := range value.([]parser.KeyValue) {
			name := pair.KeyToString()
			for _, field := range s.Fields {
				if name == field.Name {
					val := g.generateConstantValue(field.Type, pair.Value, ind+tab)
					contents += fmt.Sprintf("\n%s..%s = %s", tabtab+ind, toFieldName(name), val)
				}
			}
		}

		return contents
	}

	panic("no entry for type " + underlyingType.Name)
}

// GenerateTypeDef generates the given typedef.
func (g *Generator) GenerateTypeDef(*parser.TypeDef) error {
	// No typedefs in dart
	return nil
}

// GenerateEnum generates the given enum.
func (g *Generator) GenerateEnum(enum *parser.Enum) error {
	contents := ""
	file, err := g.GenerateFile(enum.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if g.useEnums() {
		contents += g.generateEnumUsingEnums(enum)
	} else {
		contents += g.generateEnumUsingClasses(enum)
	}

	file.WriteString(contents)

	return nil
}

func (g *Generator) generateEnumUsingClasses(enum *parser.Enum) string {
	contents := ""
	contents += fmt.Sprintf("class %s {\n", enum.Name)
	for _, field := range enum.Values {
		contents += fmt.Sprintf(tab+"static const int %s = %d;\n", field.Name, field.Value)
	}
	contents += "\n"

	contents += tab + "static final Set<int> VALID_VALUES = new Set.from([\n"
	for _, field := range enum.Values {
		contents += fmt.Sprintf(tabtab+"%s,\n", field.Name)
	}
	contents += tab + "]);\n\n"

	contents += tab + "static final Map<int, String> VALUES_TO_NAMES = {\n"
	for _, field := range enum.Values {
		contents += fmt.Sprintf(tabtab+"%s: '%s',\n", field.Name, field.Name)
	}
	contents += tab + "};\n"
	contents += "}\n"
	return contents
}

func (g *Generator) generateEnumUsingEnums(enum *parser.Enum) string {
	contents := ""
	contents += fmt.Sprintf("enum %s {\n", enum.Name)
	for _, field := range enum.Values {
		contents += fmt.Sprintf(tab+"%s,\n", field.Name)
	}
	contents += "}\n\n"

	contents += fmt.Sprintf("int serialize%s(%s variant) {\n", enum.Name, enum.Name)
	contents += tab + "switch (variant) {\n"
	for _, field := range enum.Values {
		contents += fmt.Sprintf(tabtab+"case %s.%s:\n", enum.Name, field.Name)
		contents += fmt.Sprintf(tabtabtab+"return %d;\n", field.Value)
	}
	contents += tab + "}\n"
	contents += "}\n\n"

	contents += fmt.Sprintf("%s deserialize%s(int value) {\n", enum.Name, enum.Name)
	contents += tab + "switch (value) {\n"
	for _, field := range enum.Values {
		contents += fmt.Sprintf(tabtab+"case %d:\n", field.Value)
		contents += fmt.Sprintf(tabtabtab+"return %s.%s;\n", enum.Name, field.Name)
	}
	contents += tabtab + "default:\n"
	contents += fmt.Sprintf(tabtabtab+"throw new thrift.TProtocolError(thrift.TProtocolErrorType.UNKNOWN, \"Invalid value '$value' for enum '%s'\");", enum.Name)

	contents += tab + "}\n"
	contents += "}\n"
	return contents
}

// GenerateStruct generates the given struct.
func (g *Generator) GenerateStruct(s *parser.Struct) error {
	file, err := g.GenerateFile(s.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.writeThriftImports(file); err != nil {
		return err
	}

	contents := g.generateStruct(s)
	_, err = file.WriteString(contents)
	return err
}

// GenerateUnion generates the given union.
func (g *Generator) GenerateUnion(s *parser.Struct) error {
	return g.GenerateStruct(s)
}

// GenerateException generates the given exception.
func (g *Generator) GenerateException(s *parser.Struct) error {
	return g.GenerateStruct(s)
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

func (g *Generator) generateStruct(s *parser.Struct) string {
	contents := ""

	// Class declaration
	if s.Comment != nil {
		contents += g.GenerateInlineComment(s.Comment, "/")
	}

	contents += fmt.Sprintf("class %s ", s.Name)
	if s.Type == parser.StructTypeException {
		contents += "extends Error "
	}
	contents += "implements thrift.TBase {\n"

	// Struct and field descriptors
	contents += fmt.Sprintf(tab+"static final thrift.TStruct _STRUCT_DESC = new thrift.TStruct(\"%s\");\n", s.Name)
	for _, field := range s.Fields {
		constantName := toScreamingCapsConstant(field.Name)
		contents += fmt.Sprintf(tab+"static final thrift.TField _%s_FIELD_DESC = new thrift.TField(\"%s\", %s, %d);\n",
			constantName, field.Name, g.getEnumFromThriftType(field.Type), field.ID)
	}
	contents += "\n"

	// Fields
	for _, field := range s.Fields {
		if field.Comment != nil {
			contents += g.GenerateInlineComment(field.Comment, tab+"/")
		}
		contents += fmt.Sprintf(tab+"%s _%s%s;\n",
			g.getDartTypeFromThriftType(field.Type), toFieldName(field.Name), g.generateInitValue(field))
		contents += fmt.Sprintf(tab+"static const int %s = %d;\n", strings.ToUpper(field.Name), field.ID)
	}
	contents += "\n"

	// Is set helpers for primitive types.
	for _, field := range s.Fields {
		if g.isDartPrimitive(field.Type) {
			contents += fmt.Sprintf(tab+"bool __isset_%s = false;\n", toFieldName(field.Name))
		}
	}

	// Constructor
	contents += "\n"
	contents += fmt.Sprintf(tab+"%s() {\n", s.Name)
	for _, field := range s.Fields {
		if field.Default != nil {
			value := g.generateConstantValue(field.Type, field.Default, tab)
			contents += fmt.Sprintf(tabtab+"this.%s = %s;\n", toFieldName(field.Name), value)
		}
	}
	contents += tab + "}\n\n"

	// methods for getting/setting fields
	contents += g.generateFieldMethods(s)

	// read
	contents += g.generateRead(s)

	// write
	contents += g.generateWrite(s)

	// to string
	contents += g.generateToString(s)

	// validate
	contents += g.generateValidate(s)

	contents += "}\n"
	return contents
}

func (g *Generator) generateInitValue(field *parser.Field) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	if !underlyingType.IsPrimitive() || field.Modifier == parser.Optional {
		return ""
	}

	switch underlyingType.Name {
	case "bool":
		return " = false"
	case "byte", "i8", "i16", "i32", "i64":
		return " = 0"
	case "double":
		return " = 0.0"
	default:
		return ""
	}
}

func (g *Generator) generateFieldMethods(s *parser.Struct) string {
	// Getters and setters for each field
	contents := ""
	for _, field := range s.Fields {
		dartType := g.getDartTypeFromThriftType(field.Type)
		dartPrimitive := g.isDartPrimitive(field.Type)
		fName := toFieldName(field.Name)
		titleName := strings.Title(field.Name)

		if field.Comment != nil {
			contents += g.GenerateInlineComment(field.Comment, tab+"/")
		}
		contents += fmt.Sprintf(tab+"%s get %s => this._%s;\n\n", dartType, fName, fName)
		if field.Comment != nil {
			contents += g.GenerateInlineComment(field.Comment, tab+"/")
		}
		contents += fmt.Sprintf(tab+"set %s(%s %s) {\n", fName, dartType, fName)
		contents += fmt.Sprintf(tabtab+"this._%s = %s;\n", fName, fName)
		if dartPrimitive {
			contents += fmt.Sprintf(tabtab+"this.__isset_%s = true;\n", fName)
		}
		contents += tab + "}\n\n"

		if dartPrimitive {
			contents += fmt.Sprintf(tab+"bool isSet%s() => this.__isset_%s;\n\n", titleName, fName)
			contents += fmt.Sprintf(tab+"unset%s() {\n", titleName)
			contents += fmt.Sprintf(tabtab+"this.__isset_%s = false;\n", fName)
			contents += tab + "}\n\n"
		} else {
			contents += fmt.Sprintf(tab+"bool isSet%s() => this.%s != null;\n\n", titleName, fName)
			contents += fmt.Sprintf(tab+"unset%s() {\n", titleName)
			contents += fmt.Sprintf(tabtab+"this.%s = null;\n", fName)
			contents += tab + "}\n\n"
		}
	}

	// getFieldValue
	contents += tab + "getFieldValue(int fieldID) {\n"
	contents += tabtab + "switch (fieldID) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", strings.ToUpper(field.Name))
		contents += fmt.Sprintf(tabtabtabtab+"return this.%s;\n", toFieldName(field.Name))
	}
	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "throw new ArgumentError(\"Field $fieldID doesn't exist!\");\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	// setFieldValue
	contents += tab + "setFieldValue(int fieldID, Object value) {\n"
	contents += tabtab + "switch(fieldID) {\n"
	for _, field := range s.Fields {
		fName := toFieldName(field.Name)
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", strings.ToUpper(field.Name))
		contents += tabtabtabtab + "if(value == null) {\n"
		contents += fmt.Sprintf(tabtabtabtabtab+"unset%s();\n", strings.Title(field.Name))
		contents += tabtabtabtab + "} else {\n"
		contents += fmt.Sprintf(tabtabtabtabtab+"this.%s = value as %s;\n", fName, g.getDartTypeFromThriftType(field.Type))
		contents += tabtabtabtab + "}\n"
		contents += tabtabtabtab + "break;\n\n"
	}
	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "throw new ArgumentError(\"Field $fieldID doesn't exist!\");\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	// isSet
	contents += tab + "// Returns true if the field corresponding to fieldID is set (has been assigned a value) and false otherwise\n"
	contents += tab + "bool isSet(int fieldID) {\n"
	contents += tabtab + "switch(fieldID) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", strings.ToUpper(field.Name))
		contents += fmt.Sprintf(tabtabtabtab+"return isSet%s();\n", strings.Title(field.Name))
	}
	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "throw new ArgumentError(\"Field $fieldID doesn't exist!\");\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateRead(s *parser.Struct) string {
	contents := tab + "read(thrift.TProtocol iprot) {\n"
	contents += tabtab + "thrift.TField field;\n"
	contents += tabtab + "iprot.readStructBegin();\n"
	contents += tabtab + "while(true) {\n"
	contents += tabtabtab + "field = iprot.readFieldBegin();\n"
	contents += tabtabtab + "if(field.type == thrift.TType.STOP) {\n"
	contents += tabtabtabtab + "break;\n"
	contents += tabtabtab + "}\n"
	contents += tabtabtab + "switch(field.id) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtabtab+"case %s:\n", strings.ToUpper(field.Name))
		t := g.getEnumFromThriftType(g.Frugal.UnderlyingType(field.Type))
		contents += fmt.Sprintf(tabtabtabtabtab+"if(field.type == %s) {\n", t)
		contents += g.generateReadFieldRec(field, true, tabtabtabtabtabtab)
		contents += tabtabtabtabtab + "} else {\n"
		contents += tabtabtabtabtabtab + "thrift.TProtocolUtil.skip(iprot, field.type);\n"
		contents += tabtabtabtabtab + "}\n"
		contents += tabtabtabtabtab + "break;\n"
	}
	contents += tabtabtabtab + "default:\n"
	contents += tabtabtabtabtab + "thrift.TProtocolUtil.skip(iprot, field.type);\n"
	contents += tabtabtabtabtab + "break;\n"
	contents += tabtabtab + "}\n"
	contents += tabtabtab + "iprot.readFieldEnd();\n"
	contents += tabtab + "}\n"
	contents += tabtab + "iprot.readStructEnd();\n\n"

	// validate primitives
	contents += tabtab + "// check for required fields of primitive type, which can't be checked in the validate method\n"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required && g.isDartPrimitive(field.Type) {
			fName := toFieldName(field.Name)
			contents += fmt.Sprintf(tabtab+"if(!__isset_%s) {\n", fName)
			contents += fmt.Sprintf(tabtabtab+"throw new thrift.TProtocolError(thrift.TProtocolErrorType.UNKNOWN, \"Required field '%s' was not present in struct %s\");\n", fName, s.Name)
			contents += tabtab + "}\n"
		}
	}
	contents += tabtab + "validate();\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateReadFieldRec(field *parser.Field, first bool, ind string) string {
	contents := ""

	prefix := ""
	dartType := g.getDartTypeFromThriftType(field.Type)
	if !first {
		prefix = dartType + " "
	}

	fName := toFieldName(field.Name)
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	primitive := g.isDartPrimitive(underlyingType)
	if underlyingType.IsPrimitive() {
		thriftType := ""
		switch underlyingType.Name {
		case "bool":
			thriftType = "Bool"
		case "byte", "i8":
			thriftType = "Byte"
		case "i16":
			thriftType = "I16"
		case "i32":
			thriftType = "I32"
		case "i64":
			thriftType = "I64"
		case "double":
			thriftType = "Double"
		case "string":
			thriftType = "String"
		case "binary":
			thriftType = "Binary"
		default:
			panic("unknown thrift type: " + underlyingType.Name)
		}

		contents += fmt.Sprintf(ind+"%s%s = iprot.read%s();\n", prefix, fName, thriftType)
		if primitive && first {
			contents += fmt.Sprintf(ind+"this.__isset_%s = true;\n", fName)
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		if g.useEnums() {
			contents += fmt.Sprintf(ind+"%s%s = %s.deserialize%s(iprot.readI32());\n",
				prefix, fName, g.includeQualifier(underlyingType), underlyingType.Name)
		} else {
			contents += fmt.Sprintf(ind+"%s%s = iprot.readI32();\n", prefix, fName)
		}

		if first {
			contents += fmt.Sprintf(ind+"this.__isset_%s = true;\n", fName)
		}
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(ind+"%s%s = new %s();\n", prefix, fName, dartType)
		contents += fmt.Sprintf(ind+"%s.read(iprot);\n", fName)
	} else if underlyingType.IsContainer() {
		containerElem := g.GetElem()
		valElem := g.GetElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)
		valContents := g.generateReadFieldRec(valField, false, ind+tab)
		counterElem := g.GetElem()
		dartType := g.getDartTypeFromThriftType(underlyingType)

		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf(ind+"thrift.TList %s = iprot.readListBegin();\n", containerElem)
			contents += fmt.Sprintf(ind+"%s%s = new %s();\n", prefix, fName, dartType)
			contents += fmt.Sprintf(ind+"for(int %s = 0; %s < %s.length; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += fmt.Sprintf(tab+ind+"%s.add(%s);\n", fName, valElem)
			contents += ind + "}\n"
			contents += ind + "iprot.readListEnd();\n"
		case "set":
			contents += fmt.Sprintf(ind+"thrift.TSet %s = iprot.readSetBegin();\n", containerElem)
			contents += fmt.Sprintf(ind+"%s%s = new %s();\n", prefix, fName, dartType)
			contents += fmt.Sprintf(ind+"for(int %s = 0; %s < %s.length; ++%s) {\n",
				counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += fmt.Sprintf(tab+ind+"%s.add(%s);\n", fName, valElem)
			contents += ind + "}\n"
			contents += ind + "iprot.readSetEnd();\n"
		case "map":
			contents += fmt.Sprintf(ind+"thrift.TMap %s = iprot.readMapBegin();\n", containerElem)
			contents += fmt.Sprintf(ind+"%s%s = new %s();\n", prefix, fName, dartType)
			contents += fmt.Sprintf(ind+"for(int %s = 0; %s < %s.length; ++%s) {\n",
				counterElem, counterElem, containerElem, counterElem)
			keyElem := g.GetElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			contents += g.generateReadFieldRec(keyField, false, ind+tab)
			contents += valContents
			contents += fmt.Sprintf(tab+ind+"%s[%s] = %s;\n", fName, keyElem, valElem)
			contents += ind + "}\n"
			contents += ind + "iprot.readMapEnd();\n"
		default:
			panic("unrecognized thrift type: " + underlyingType.Name)
		}
	}

	return contents
}

func (g *Generator) generateWrite(s *parser.Struct) string {
	contents := tab + "write(thrift.TProtocol oprot) {\n"
	contents += tabtab + "validate();\n\n"
	contents += tabtab + "oprot.writeStructBegin(_STRUCT_DESC);\n"
	for _, field := range s.Fields {
		fName := toFieldName(field.Name)
		optional := field.Modifier == parser.Optional
		nullable := !g.isDartPrimitive(g.Frugal.UnderlyingType(field.Type))
		ind := ""
		if optional || nullable {
			ind = tab
			contents += tabtab + "if("
			if optional {
				contents += fmt.Sprintf("isSet%s()", strings.Title(field.Name))
			}
			if optional && nullable {
				contents += " && "
			}
			if nullable {
				contents += fmt.Sprintf("this.%s != null", fName)
			}
			contents += ") {\n"
		}

		contents += fmt.Sprintf(tabtab+ind+"oprot.writeFieldBegin(_%s_FIELD_DESC);\n", toScreamingCapsConstant(field.Name))
		contents += g.generateWriteFieldRec(field, true, ind)
		contents += fmt.Sprintf(tabtab + ind + "oprot.writeFieldEnd();\n")

		if optional || nullable {
			contents += tabtab + "}\n"
		}
	}
	contents += tabtab + "oprot.writeFieldStop();\n"
	contents += tabtab + "oprot.writeStructEnd();\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateWriteFieldRec(field *parser.Field, first bool, ind string) string {
	contents := ""

	fName := toFieldName(field.Name)
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	if underlyingType.IsPrimitive() {
		write := tabtab + ind + "oprot.write"
		switch underlyingType.Name {
		case "bool":
			write += "Bool(%s);\n"
		case "byte", "i8":
			write += "Byte(%s);\n"
		case "i16":
			write += "I16(%s);\n"
		case "i32":
			write += "I32(%s);\n"
		case "i64":
			write += "I64(%s);\n"
		case "double":
			write += "Double(%s);\n"
		case "string":
			write += "String(%s);\n"
		case "binary":
			write += "Binary(%s);\n"
		default:
			panic("unknown thrift type: " + underlyingType.Name)
		}

		contents += fmt.Sprintf(write, fName)
	} else if g.Frugal.IsEnum(underlyingType) {
		if g.useEnums() {
			contents += fmt.Sprintf(tabtab+"oprot.writeI32(%s.serialize%s(%s));\n",
				g.includeQualifier(underlyingType), underlyingType.Name, fName)
		} else {
			contents += fmt.Sprintf(tabtab+ind+"oprot.writeI32(%s);\n", fName)
		}
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(tabtab+ind+"%s.write(oprot);\n", fName)
	} else if underlyingType.IsContainer() {
		valEnumType := g.getEnumFromThriftType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			valElem := g.GetElem()
			valField := parser.FieldFromType(underlyingType.ValueType, valElem)
			contents += fmt.Sprintf(tabtab+ind+"oprot.writeListBegin(new thrift.TList(%s, %s.length));\n", valEnumType, fName)
			contents += fmt.Sprintf(tabtab+ind+"for(var %s in %s) {\n", valElem, fName)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += tabtab + ind + "}\n"
			contents += tabtab + ind + "oprot.writeListEnd();\n"
		case "set":
			valElem := g.GetElem()
			valField := parser.FieldFromType(underlyingType.ValueType, valElem)
			contents += fmt.Sprintf(tabtab+ind+"oprot.writeSetBegin(new thrift.TSet(%s, %s.length));\n", valEnumType, fName)
			contents += fmt.Sprintf(tabtab+ind+"for(var %s in %s) {\n", valElem, fName)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += tabtab + ind + "}\n"
			contents += tabtab + ind + "oprot.writeSetEnd();\n"
		case "map":
			keyEnumType := g.getEnumFromThriftType(underlyingType.KeyType)
			keyElem := g.GetElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			valField := parser.FieldFromType(underlyingType.ValueType, fmt.Sprintf("%s[%s]", fName, keyElem))
			contents += fmt.Sprintf(tabtab+ind+"oprot.writeMapBegin(new thrift.TMap(%s, %s, %s.length));\n", keyEnumType, valEnumType, fName)
			contents += fmt.Sprintf(tabtab+ind+"for(var %s in %s.keys) {\n", keyElem, fName)
			contents += g.generateWriteFieldRec(keyField, false, ind+tab)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += tabtab + ind + "}\n"
			contents += tabtab + ind + "oprot.writeMapEnd();\n"
		default:
			panic("unknown type: " + underlyingType.Name)
		}
	}

	return contents
}

func (g *Generator) generateToString(s *parser.Struct) string {
	contents := tab + "String toString() {\n"
	contents += fmt.Sprintf(tabtab+"StringBuffer ret = new StringBuffer(\"%s(\");\n\n", s.Name)
	first := true
	for _, field := range s.Fields {
		fName := toFieldName(field.Name)
		optional := field.Modifier == parser.Optional
		ind := ""
		optInd := ""
		if optional {
			contents += fmt.Sprintf(tabtab+"if(isSet%s()) {\n", strings.Title(field.Name))
			ind += tab
			optInd = tab
		}

		if !first {
			contents += tabtab + ind + "ret.write(\", \");\n"
		}
		contents += fmt.Sprintf(tabtab+ind+"ret.write(\"%s:\");\n", fName)

		if !g.isDartPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtab+ind+"if(this.%s == null) {\n", fName)
			contents += tabtabtab + ind + "ret.write(\"null\");\n"
			contents += tabtab + ind + "} else {\n"
			ind += tab
		}

		underlyingType := g.Frugal.UnderlyingType(field.Type)
		if underlyingType.Name == "binary" {
			// Don't want to write out
			contents += tabtab + ind + "ret.write(\"BINARY\");\n"
		} else if g.Frugal.IsEnum(underlyingType) {
			contents += fmt.Sprintf(tabtab+ind+"String %s_name = %s.VALUES_TO_NAMES[this.%s];\n",
				fName, g.qualifiedTypeName(field.Type), fName)
			contents += fmt.Sprintf(tabtab+ind+"if(%s_name != null) {\n", fName)
			contents += fmt.Sprintf(tabtabtab+ind+"ret.write(%s_name);\n", fName)
			contents += tabtabtab + ind + "ret.write(\" (\");\n"
			contents += tabtab + ind + "}\n"
			contents += tabtab + ind + fmt.Sprintf("ret.write(this.%s);\n", fName)
			contents += fmt.Sprintf(tabtab+ind+"if(%s_name != null) {\n", fName)
			contents += tabtabtab + ind + "ret.write(\")\");\n"
			contents += tabtab + ind + "}\n"
		} else {
			contents += fmt.Sprintf(tabtab+ind+"ret.write(this.%s);\n", fName)
		}

		if !g.isDartPrimitive(field.Type) {
			contents += tabtab + optInd + "}\n"
		}

		if optional {
			contents += tabtab + "}\n"
		}
		contents += "\n"
		first = false
	}
	contents += tabtab + "ret.write(\")\");\n\n"
	contents += tabtab + "return ret.toString();\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateValidate(s *parser.Struct) string {
	contents := tab + "validate() {\n"

	if s.Type != parser.StructTypeUnion {
		contents += tabtab + "// check for required fields\n"
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				fName := toFieldName(field.Name)
				if !g.isDartPrimitive(field.Type) {
					contents += fmt.Sprintf(tabtab+"if(%s == null) {\n", fName)
					contents += fmt.Sprintf(tabtabtab+"throw new thrift.TProtocolError(thrift.TProtocolErrorType.INVALID_DATA, \"Required field '%s' was not present in struct %s\");\n", fName, s.Name)
					contents += tabtab + "}\n"
				}
			}
		}
	} else {
		contents += tabtab + "// check exactly one field is set\n"
		contents += tabtab + "int setFields = 0;\n"
		for _, field := range s.Fields {
			contents += fmt.Sprintf(tabtab+"if(isSet%s()) {\n", strings.Title(field.Name))
			contents += tabtabtab + "setFields++;\n"
			contents += tabtab + "}\n"
		}
		contents += tabtab + "if(setFields != 1) {\n"
		contents += tabtabtab + "throw new thrift.TProtocolError(thrift.TProtocolErrorType.INVALID_DATA, \"The union did not have exactly one field set, $setFields were set\");\n"
		contents += tabtab + "}\n"
	}

	if !g.useEnums() {
		contents += tabtab + "// check that fields of type enum have valid values\n"
		for _, field := range s.Fields {
			if g.Frugal.IsEnum(field.Type) {
				fName := toFieldName(field.Name)
				isSetCheck := fmt.Sprintf("isSet%s()", strings.Title(field.Name))
				contents += fmt.Sprintf(tabtab+"if(%s && !%s.VALID_VALUES.contains(%s)) {\n",
					isSetCheck, g.qualifiedTypeName(field.Type), fName)
				contents += fmt.Sprintf(tabtabtab+"throw new thrift.TProtocolError(thrift.TProtocolErrorType.INVALID_DATA, \"The field '%s' has been assigned the invalid value $%s\");\n", fName, fName)
				contents += tabtab + "}\n"
			}
		}
	}

	contents += tab + "}\n"
	return contents
}

// GenerateServicePackage generates the package for the given service.
func (g *Generator) GenerateServicePackage(file *os.File, s *parser.Service) error {
	return nil
}

// GenerateScopePackage generates the package for the given scope.
func (g *Generator) GenerateScopePackage(file *os.File, s *parser.Scope) error {
	return nil
}

// GenerateObjectPackage generates the package for the given name.
func (g *Generator) GenerateObjectPackage(file *os.File, name string) error {
	return nil
}

// GenerateThriftImports generates necessary imports for Thrift.
func (g *Generator) GenerateThriftImports() (string, error) {
	imports := "import 'dart:typed_data' show Uint8List;\n"
	imports += "import 'package:thrift/thrift.dart' as thrift;\n"
	// Import the current package
	imports += g.getImportDeclaration(g.getNamespaceOrName(), g.getPackagePrefix())

	// Import includes
	for _, include := range g.Frugal.Includes {
		if imp, err := g.generateIncludeImport(include); err != nil {
			return "", err
		} else {
			imports += imp
		}
	}

	imports += "\n"
	return imports, nil
}

func (g *Generator) writeThriftImports(file *os.File) error {
	imports, err := g.GenerateThriftImports()
	if err != nil {
		return err
	}

	_, err = file.WriteString(imports)
	if err != nil {
		return err
	}

	return nil
}

// GenerateServiceImports generates necessary imports for the given service.
func (g *Generator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	imports := "import 'dart:async';\n\n"
	imports += "import 'dart:typed_data' show Uint8List;\n"
	imports += "import 'package:logging/logging.dart' as logging;\n"
	imports += "import 'package:thrift/thrift.dart' as thrift;\n"
	imports += "import 'package:frugal/frugal.dart' as frugal;\n\n"
	// import included packages
	includes, err := s.ReferencedIncludes()
	if err != nil {
		return err
	}
	for _, include := range includes {
		if imp, err := g.generateIncludeImport(include); err != nil {
			return err
		} else {
			imports += imp
		}
	}

	// Import same package.
	imports += g.getImportDeclaration(g.getNamespaceOrName(), g.getPackagePrefix())

	_, err = file.WriteString(imports)
	return err
}

// GenerateScopeImports generates necessary imports for the given scope.
func (g *Generator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "import 'dart:async';\n\n"
	imports += "import 'package:thrift/thrift.dart' as thrift;\n"
	imports += "import 'package:frugal/frugal.dart' as frugal;\n\n"
	// import included packages
	scopeIncludes, err := s.ReferencedIncludes()
	if err != nil {
		return err
	}
	for _, include := range scopeIncludes {
		if imp, err := g.generateIncludeImport(include); err != nil {
			return err
		} else {
			imports += imp
		}
	}

	// Import same package.
	imports += g.getImportDeclaration(g.getNamespaceOrName(), g.getPackagePrefix())

	_, err = file.WriteString(imports)
	return err
}

// GenerateConstants generates any static constants.
func (g *Generator) GenerateConstants(file *os.File, name string) error {
	constants := fmt.Sprintf("const String delimiter = '%s';", globals.TopicDelimiter)
	_, err := file.WriteString(constants)
	return err
}

// GeneratePublisher generates the publisher for the given scope.
func (g *Generator) GeneratePublisher(file *os.File, scope *parser.Scope) error {
	publishers := ""
	if scope.Comment != nil {
		publishers += g.GenerateInlineComment(scope.Comment, "/")
	}
	publishers += fmt.Sprintf("class %sPublisher {\n", strings.Title(scope.Name))
	publishers += tab + "frugal.FPublisherTransport transport;\n"
	publishers += tab + "frugal.FProtocolFactory protocolFactory;\n"
	publishers += tab + "Map<String, frugal.FMethod> _methods;\n"

	publishers += fmt.Sprintf(tab+"%sPublisher(frugal.FScopeProvider provider, [List<frugal.Middleware> middleware]) {\n", strings.Title(scope.Name))
	publishers += tabtab + "transport = provider.publisherTransportFactory.getTransport();\n"
	publishers += tabtab + "protocolFactory = provider.protocolFactory;\n"
	publishers += tabtab + "var combined = middleware ?? [];\n"
	publishers += tabtab + "combined.addAll(provider.middleware);\n"
	publishers += tabtab + "this._methods = {};\n"
	for _, operation := range scope.Operations {
		publishers += fmt.Sprintf(tabtab+"this._methods['%s'] = new frugal.FMethod(this._publish%s, '%s', 'publish%s', combined);\n",
			operation.Name, operation.Name, strings.Title(scope.Name), operation.Name)
	}
	publishers += tab + "}\n\n"

	publishers += tab + "Future open() {\n"
	publishers += tabtab + "return transport.open();\n"
	publishers += tab + "}\n\n"

	publishers += tab + "Future close() {\n"
	publishers += tabtab + "return transport.close();\n"
	publishers += tab + "}\n\n"

	args := ""
	argsWithoutTypes := ""
	if len(scope.Prefix.Variables) > 0 {
		for _, variable := range scope.Prefix.Variables {
			args = fmt.Sprintf("%sString %s, ", args, variable)
			argsWithoutTypes = fmt.Sprintf("%s%s, ", argsWithoutTypes, variable)
		}
	}
	prefix := ""
	for _, op := range scope.Operations {
		publishers += prefix
		prefix = "\n\n"
		if op.Comment != nil {
			publishers += g.GenerateInlineComment(op.Comment, tab+"/")
		}

		publishers += fmt.Sprintf(tab+"Future publish%s(frugal.FContext ctx, %s%s req) {\n", op.Name, args, g.getDartTypeFromThriftType(op.Type))

		publishers += fmt.Sprintf(tabtab+"return this._methods['%s']([ctx, %sreq]);\n", op.Name, argsWithoutTypes)
		publishers += tab + "}\n\n"

		publishers += fmt.Sprintf(tab+"Future _publish%s(frugal.FContext ctx, %s%s req) async {\n", op.Name, args, g.getDartTypeFromThriftType(op.Type))

		// Inject the prefix variables into the FContext to send
		for _, prefixVar := range scope.Prefix.Variables {
			publishers += fmt.Sprintf(tabtab+"ctx.addRequestHeader('_topic_%s', %s);\n", prefixVar, prefixVar)
		}

		publishers += tabtab + fmt.Sprintf("var op = \"%s\";\n", op.Name)
		publishers += tabtab + fmt.Sprintf("var prefix = \"%s\";\n", generatePrefixStringTemplate(scope))
		publishers += tabtab + "var topic = \"${prefix}" + strings.Title(scope.Name) + "${delimiter}${op}\";\n"
		publishers += tabtab + "var memoryBuffer = new frugal.TMemoryOutputBuffer(transport.publishSizeLimit);\n"
		publishers += tabtab + "var oprot = protocolFactory.getProtocol(memoryBuffer);\n"
		publishers += tabtab + "var msg = new thrift.TMessage(op, thrift.TMessageType.CALL, 0);\n"
		publishers += tabtab + "oprot.writeRequestHeader(ctx);\n"
		publishers += tabtab + "oprot.writeMessageBegin(msg);\n"
		publishers += g.generateWriteFieldRec(parser.FieldFromType(op.Type, "req"), false, "")
		publishers += tabtab + "oprot.writeMessageEnd();\n"
		publishers += tabtab + "await transport.publish(topic, memoryBuffer.writeBytes);\n"
		publishers += tab + "}\n"
	}

	publishers += "}\n"

	_, err := file.WriteString(publishers)
	return err
}

func generatePrefixStringTemplate(scope *parser.Scope) string {
	if scope.Prefix.String == "" {
		return ""
	}
	template := ""
	template += scope.Prefix.Template("%s")
	template += globals.TopicDelimiter
	if len(scope.Prefix.Variables) == 0 {
		return template
	}
	vars := make([]interface{}, len(scope.Prefix.Variables))
	for i, variable := range scope.Prefix.Variables {
		vars[i] = fmt.Sprintf("${%s}", variable)
	}
	template = fmt.Sprintf(template, vars...)
	return template
}

// GenerateSubscriber generates the subscriber for the given scope.
func (g *Generator) GenerateSubscriber(file *os.File, scope *parser.Scope) error {
	subscribers := ""
	if scope.Comment != nil {
		subscribers += g.GenerateInlineComment(scope.Comment, "/")
	}
	subscribers += fmt.Sprintf("class %sSubscriber {\n", strings.Title(scope.Name))
	subscribers += tab + "final frugal.FScopeProvider provider;\n"
	subscribers += tab + "final List<frugal.Middleware> _middleware;\n\n"

	subscribers += tab + fmt.Sprintf("%sSubscriber(this.provider, [List<frugal.Middleware> middleware])\n", strings.Title(scope.Name))
	subscribers += tabtabtab + ": this._middleware = middleware ?? [] {\n"
	subscribers += tabtab + "this._middleware.addAll(provider.middleware);\n"
	subscribers += "}\n\n"

	args := ""
	if len(scope.Prefix.Variables) > 0 {
		for _, variable := range scope.Prefix.Variables {
			args = fmt.Sprintf("%sString %s, ", args, variable)
		}
	}
	prefix := ""
	for _, op := range scope.Operations {
		subscribers += prefix
		prefix = "\n\n"
		if op.Comment != nil {
			subscribers += g.GenerateInlineComment(op.Comment, tab+"/")
		}
		subscribers += fmt.Sprintf(tab+"Future<frugal.FSubscription> subscribe%s(%sdynamic on%s(frugal.FContext ctx, %s req)) async {\n",
			op.Name, args, op.Type.ParamName(), g.getDartTypeFromThriftType(op.Type))
		subscribers += fmt.Sprintf(tabtab+"var op = \"%s\";\n", op.Name)
		subscribers += fmt.Sprintf(tabtab+"var prefix = \"%s\";\n", generatePrefixStringTemplate(scope))
		subscribers += tabtab + "var topic = \"${prefix}" + strings.Title(scope.Name) + "${delimiter}${op}\";\n"
		subscribers += tabtab + "var transport = provider.subscriberTransportFactory.getTransport();\n"
		subscribers += fmt.Sprintf(tabtab+"await transport.subscribe(topic, _recv%s(op, provider.protocolFactory, on%s));\n",
			op.Name, op.Type.ParamName())
		subscribers += tabtab + "return new frugal.FSubscription(topic, transport);\n"
		subscribers += tab + "}\n\n"

		subscribers += fmt.Sprintf(tab+"frugal.FAsyncCallback _recv%s(String op, frugal.FProtocolFactory protocolFactory, dynamic on%s(frugal.FContext ctx, %s req)) {\n",
			op.Name, op.Type.ParamName(), g.getDartTypeFromThriftType(op.Type))
		subscribers += fmt.Sprintf(tabtab+"frugal.FMethod method = new frugal.FMethod(on%s, '%s', 'subscribe%s', this._middleware);\n",
			op.Type.ParamName(), strings.Title(scope.Name), op.Type.ParamName())
		subscribers += fmt.Sprintf(tabtab+"callback%s(thrift.TTransport transport) {\n", op.Name)

		subscribers += tabtabtab + "var iprot = protocolFactory.getProtocol(transport);\n"
		subscribers += tabtabtab + "var ctx = iprot.readRequestHeader();\n"
		subscribers += tabtabtab + "var tMsg = iprot.readMessageBegin();\n"
		subscribers += tabtabtab + "if (tMsg.name != op) {\n"
		subscribers += tabtabtabtab + "thrift.TProtocolUtil.skip(iprot, thrift.TType.STRUCT);\n"
		subscribers += tabtabtabtab + "iprot.readMessageEnd();\n"
		subscribers += tabtabtabtab + "throw new thrift.TApplicationError(\n"
		subscribers += tabtabtabtab + "frugal.FrugalTApplicationErrorType.UNKNOWN_METHOD, tMsg.name);\n"
		subscribers += tabtabtab + "}\n"
		subscribers += g.generateReadFieldRec(parser.FieldFromType(op.Type, "req"), false, tabtabtab)
		subscribers += tabtabtab + "iprot.readMessageEnd();\n"
		subscribers += tabtabtab + "method([ctx, req]);\n"
		subscribers += tabtab + "}\n"
		subscribers += fmt.Sprintf(tabtab+"return callback%s;\n", op.Name)
		subscribers += tab + "}\n"
	}

	subscribers += "}\n"

	_, err := file.WriteString(subscribers)
	return err
}

// GenerateService generates the given service.
func (g *Generator) GenerateService(file *os.File, s *parser.Service) error {
	contents := ""
	contents += g.generateInterface(s)
	contents += g.generateClient(s)
	contents += g.generateServiceArgsResults(s)

	_, err := file.WriteString(contents)
	return err
}

func (g *Generator) generateInterface(service *parser.Service) string {
	contents := ""
	if service.Comment != nil {
		contents += g.GenerateInlineComment(service.Comment, "/")
	}
	if service.Extends != "" {
		contents += fmt.Sprintf("abstract class F%s extends %s {\n",
			strings.Title(service.Name), g.getServiceExtendsName(service))
	} else {
		contents += fmt.Sprintf("abstract class F%s {\n", strings.Title(service.Name))
	}
	for _, method := range service.Methods {
		contents += "\n"
		if method.Comment != nil {
			contents += g.GenerateInlineComment(method.Comment, tab+"/")
		}

		if _, ok := method.Annotations.Deprecated(); ok {
			contents += tab + "@deprecated\n"
		}

		contents += fmt.Sprintf(tab+"Future%s %s(frugal.FContext ctx%s);\n",
			g.generateReturnArg(method), parser.LowercaseFirstLetter(method.Name), g.generateInputArgs(method.Arguments))
	}
	contents += "}\n\n"
	return contents
}

func (g *Generator) getServiceExtendsName(service *parser.Service) string {
	serviceName := "F" + service.ExtendsService()
	prefix := ""
	include := service.ExtendsInclude()
	if include != "" {
		if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
			include = namespace.Value
		}
		prefix = "t_" + toLibraryName(include)
	} else {
		prefix = "t_" + toLibraryName(g.getNamespaceOrName())
	}
	return prefix + "." + serviceName
}

func (g *Generator) generateClient(service *parser.Service) string {
	servTitle := strings.Title(service.Name)
	contents := ""
	if service.Comment != nil {
		contents += g.GenerateInlineComment(service.Comment, "/")
	}
	if service.Extends != "" {
		contents += fmt.Sprintf("class F%sClient extends %sClient implements F%s {\n",
			servTitle, g.getServiceExtendsName(service), servTitle)
	} else {
		contents += fmt.Sprintf("class F%sClient implements F%s {\n",
			servTitle, servTitle)
	}
	contents += fmt.Sprintf(tab+"static final logging.Logger _log = new logging.Logger('%s');\n", servTitle)
	contents += tab + "Map<String, frugal.FMethod> _methods;\n\n"

	if service.Extends != "" {
		contents += tab + fmt.Sprintf("F%sClient(frugal.FServiceProvider provider, [List<frugal.Middleware> middleware])\n", servTitle)
		contents += tabtabtab + ": super(provider, middleware) {\n"
	} else {
		contents += tab + fmt.Sprintf("F%sClient(frugal.FServiceProvider provider, [List<frugal.Middleware> middleware]) {\n", servTitle)
	}
	contents += tabtab + "_transport = provider.transport;\n"
	contents += tabtab + "_protocolFactory = provider.protocolFactory;\n"
	contents += tabtab + "var combined = middleware ?? [];\n"
	contents += tabtab + "combined.addAll(provider.middleware);\n"
	contents += tabtab + "this._methods = {};\n"
	for _, method := range service.Methods {
		nameLower := parser.LowercaseFirstLetter(method.Name)
		contents += fmt.Sprintf(tabtab+"this._methods['%s'] = new frugal.FMethod(this._%s, '%s', '%s', combined);\n",
			nameLower, nameLower, servTitle, nameLower)
	}
	contents += tab + "}\n\n"

	contents += tab + "frugal.FTransport _transport;\n"
	contents += tab + "frugal.FProtocolFactory _protocolFactory;\n"
	contents += "\n"

	for _, method := range service.Methods {
		contents += g.generateClientMethod(service, method)
	}
	contents += "}\n\n"

	return contents
}

func (g *Generator) generateClientMethod(service *parser.Service, method *parser.Method) string {
	nameLower := parser.LowercaseFirstLetter(method.Name)
	contents := ""

	if method.Comment != nil {
		contents += g.GenerateInlineComment(method.Comment, tab+"/")
	}

	deprecationValue, deprecated := method.Annotations.Deprecated()
	if deprecated {
		if deprecationValue != "" {
			contents += fmt.Sprintf(tab+"@Deprecated(\"%s\")\n", deprecationValue)
		} else {
			contents += tab + "@deprecated\n"
		}
	}

	// Generate wrapper method
	contents += fmt.Sprintf(tab+"Future%s %s(frugal.FContext ctx%s) {\n",
		g.generateReturnArg(method), nameLower, g.generateInputArgs(method.Arguments))

	if deprecated {
		contents += fmt.Sprintf(tabtab+"_log.warning(\"Call to deprecated function '%s.%s'\");\n", service.Name, nameLower)
	}

	contents += fmt.Sprintf(tabtab+"return this._methods['%s']([ctx%s]) as Future%s;\n",
		nameLower, g.generateInputArgsWithoutTypes(method.Arguments), g.generateReturnArg(method))
	contents += fmt.Sprintf(tab + "}\n\n")

	// Generate the calling method
	contents += fmt.Sprintf(tab+"Future%s _%s(frugal.FContext ctx%s) async {\n",
		g.generateReturnArg(method), nameLower, g.generateInputArgs(method.Arguments))

	indent := tabtab

	contents += indent + "var memoryBuffer = new frugal.TMemoryOutputBuffer(_transport.requestSizeLimit);\n"
	contents += indent + "var oprot = _protocolFactory.getProtocol(memoryBuffer);\n"
	contents += indent + "oprot.writeRequestHeader(ctx);\n"
	msgType := "CALL"
	if method.Oneway {
		msgType = "ONEWAY"
	}
	contents += fmt.Sprintf(indent+"oprot.writeMessageBegin(new thrift.TMessage(\"%s\", thrift.TMessageType.%s, 0));\n",
		nameLower, msgType)
	contents += fmt.Sprintf(indent+"%s_args args = new %s_args();\n", method.Name, method.Name)
	for _, arg := range method.Arguments {
		argLower := parser.LowercaseFirstLetter(arg.Name)
		contents += fmt.Sprintf(indent+"args.%s = %s;\n", argLower, argLower)
	}
	contents += indent + "args.write(oprot);\n"
	contents += indent + "oprot.writeMessageEnd();\n"

	if method.Oneway {
		contents += indent + "await _transport.oneway(ctx, memoryBuffer.writeBytes);\n"
		contents += tab + "}\n\n"
		return contents
	}

	contents += indent + "var response = await _transport.request(ctx, memoryBuffer.writeBytes);\n"
	contents += "\n"

	contents += tabtab + "var iprot = _protocolFactory.getProtocol(response);\n"
	contents += tabtab + "iprot.readResponseHeader(ctx);\n"
	contents += tabtab + "thrift.TMessage msg = iprot.readMessageBegin();\n"
	contents += tabtab + "if (msg.type == thrift.TMessageType.EXCEPTION) {\n"
	contents += tabtabtab + "thrift.TApplicationError error = thrift.TApplicationError.read(iprot);\n"
	contents += tabtabtab + "iprot.readMessageEnd();\n"
	contents += tabtabtab + "if (error.type == frugal.FrugalTTransportErrorType.REQUEST_TOO_LARGE) {\n"
	contents += tabtabtabtab + "throw new thrift.TTransportError(frugal.FrugalTTransportErrorType.RESPONSE_TOO_LARGE, error.message);\n"
	contents += tabtabtab + "}\n"
	contents += tabtabtab + "throw error;\n"
	contents += tabtab + "}\n\n"

	contents += fmt.Sprintf(tabtab+"%s_result result = new %s_result();\n", method.Name, method.Name)
	contents += tabtab + "result.read(iprot);\n"
	contents += tabtab + "iprot.readMessageEnd();\n"
	if method.ReturnType == nil {
		contents += g.generateErrors(method)
	} else {
		contents += tabtab + "if (result.isSetSuccess()) {\n"
		contents += tabtabtab + "return result.success;\n"
		contents += tabtab + "}\n\n"
		contents += g.generateErrors(method)
		contents += tabtab + "throw new thrift.TApplicationError(\n"
		contents += fmt.Sprintf(tabtabtab+"frugal.FrugalTApplicationErrorType.MISSING_RESULT, "+
			"\"%s failed: unknown result\"\n",
			nameLower)
		contents += tabtab + ");\n"
	}
	contents += tab + "}\n"

	return contents
}

func (g *Generator) generateReturnArg(method *parser.Method) string {
	if method.ReturnType == nil {
		return ""
	}
	return fmt.Sprintf("<%s>", g.getDartTypeFromThriftType(method.ReturnType))
}

func (g *Generator) generateInputArgs(args []*parser.Field) string {
	argStr := ""
	for _, arg := range args {
		argStr += ", " + g.getDartTypeFromThriftType(arg.Type) + " " + parser.LowercaseFirstLetter(arg.Name)
	}
	return argStr
}

func (g *Generator) generateInputArgsWithoutTypes(args []*parser.Field) string {
	argStr := ""
	for _, arg := range args {
		argStr += ", " + parser.LowercaseFirstLetter(arg.Name)
	}
	return argStr
}

func (g *Generator) generateErrors(method *parser.Method) string {
	contents := ""
	for _, exp := range method.Exceptions {
		contents += fmt.Sprintf(tabtab+"if (result.%s != null) {\n", parser.LowercaseFirstLetter(exp.Name))
		contents += fmt.Sprintf(tabtabtab+"throw result.%s;\n", parser.LowercaseFirstLetter(exp.Name))
		contents += tabtab + "}\n"
	}
	return contents
}

func (g *Generator) getEnumFromThriftType(t *parser.Type) string {
	underlyingType := g.Frugal.UnderlyingType(t)
	switch underlyingType.Name {
	case "bool":
		return "thrift.TType.BOOL"
	case "byte", "i8":
		return "thrift.TType.BYTE"
	case "i16":
		return "thrift.TType.I16"
	case "i32":
		return "thrift.TType.I32"
	case "i64":
		return "thrift.TType.I64"
	case "double":
		return "thrift.TType.DOUBLE"
	case "string", "binary":
		return "thrift.TType.STRING"
	case "list":
		return "thrift.TType.LIST"
	case "set":
		return "thrift.TType.SET"
	case "map":
		return "thrift.TType.MAP"
	default:
		if g.Frugal.IsEnum(underlyingType) {
			return "thrift.TType.I32"
		} else if g.Frugal.IsStruct(underlyingType) {
			return "thrift.TType.STRUCT"
		}
		panic("not a valid thrift type: " + underlyingType.Name)
	}
}

func (g *Generator) isDartPrimitive(t *parser.Type) bool {
	underlyingType := g.Frugal.UnderlyingType(t)
	switch underlyingType.Name {
	case "bool", "byte", "i8", "i16", "i32", "i64", "double":
		return true
	default:
		if g.Frugal.IsEnum(underlyingType) {
			return true
		}
		return false
	}
}

func (g *Generator) getDartTypeFromThriftType(t *parser.Type) string {
	if t == nil {
		return "void"
	}
	underlyingType := g.Frugal.UnderlyingType(t)

	if g.Frugal.IsEnum(underlyingType) {
		return "int"
	}

	switch underlyingType.Name {
	case "bool":
		return "bool"
	case "byte", "i8":
		return "int"
	case "i16":
		return "int"
	case "i32":
		return "int"
	case "i64":
		return "int"
	case "double":
		return "double"
	case "string":
		return "String"
	case "binary":
		return "Uint8List"
	case "list":
		return fmt.Sprintf("List<%s>",
			g.getDartTypeFromThriftType(underlyingType.ValueType))
	case "set":
		return fmt.Sprintf("Set<%s>",
			g.getDartTypeFromThriftType(underlyingType.ValueType))
	case "map":
		return fmt.Sprintf("Map<%s, %s>",
			g.getDartTypeFromThriftType(underlyingType.KeyType),
			g.getDartTypeFromThriftType(underlyingType.ValueType))
	default:
		// This is a custom type
		return g.qualifiedTypeName(t)
	}
}

// get qualafied type names for custom types
func (g *Generator) qualifiedTypeName(t *parser.Type) string {
	param := t.ParamName()
	include := t.IncludeName()
	if include != "" {
		name := include
		if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
			name = namespace.Value
		}
		param = fmt.Sprintf("t_%s.%s", toLibraryName(name), param)
	} else {
		param = fmt.Sprintf("t_%s.%s", toLibraryName(g.getNamespaceOrName()), param)
	}
	return param
}

func (g *Generator) includeQualifier(t *parser.Type) string {
	include := t.IncludeName()
	if include != "" {
		if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
			include = namespace.Value
		}
		return fmt.Sprintf("t_%s", toLibraryName(include))
	}
	return fmt.Sprintf("t_%s", toLibraryName(g.getNamespaceOrName()))
}

func (g *Generator) getLibraryPrefix() string {
	prefix := ""
	if _, ok := g.Options[libraryPrefixOption]; ok {
		prefix += g.Options[libraryPrefixOption]
		if !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
	}
	return prefix
}

func (g *Generator) getPackagePrefix() string {
	prefix := ""
	if _, ok := g.Options[libraryPrefixOption]; ok {
		prefix += strings.Replace(g.getLibraryPrefix(), ".", "/", -1)
	}
	return prefix
}

func (g *Generator) generateIncludeImport(include *parser.Include) (string, error) {
	name := include.Name

	_, vendored := include.Annotations.Vendor()
	vendored = vendored && g.useVendor()
	vendorPath := ""

	if namespace := g.Frugal.NamespaceForInclude(name, lang); namespace != nil {
		name = namespace.Value
		if nsVendorPath, ok := namespace.Annotations.Vendor(); ok {
			vendorPath = nsVendorPath
		}
	}

	prefix := g.getPackagePrefix()
	if vendored {
		if vendorPath == "" {
			return "", fmt.Errorf("Vendored include %s does not specify vendor path for dart namespace", include.Name)
		}
		prefix = vendorPath
	}

	return g.getImportDeclaration(name, prefix), nil
}

func (g *Generator) getImportDeclaration(namespace, prefix string) string {
	namespace = toLibraryName(filepath.Base(namespace))
	if prefix == "" {
		prefix += namespace + "/"
	} else if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return fmt.Sprintf("import 'package:%s%s.dart' as t_%s;\n", prefix, namespace, namespace)
}

func (g *Generator) getNamespaceOrName() string {
	namespace := g.Frugal.Namespace(lang)
	if namespace != nil {
		return namespace.Value
	}
	return g.Frugal.Name
}

func (g *Generator) useEnums() bool {
	_, useEnums := g.Options["use_enums"]
	return useEnums
}

func (g *Generator) useVendor() bool {
	_, ok := g.Options[useVendorOption]
	return ok
}

func toLibraryName(name string) string {
	return strings.Replace(name, ".", "_", -1)
}

func snakeToCamel(name string) string {
	result := ""

	words := strings.Split(name, "_")
	for _, word := range words {
		result += strings.Title(word)
	}

	return result
}

// e.g. change APIForFileIO to api_for_file_io
func toFileName(name string) string {
	ret := ""
	tmp := []rune(name)
	isPrevLc := true
	isCurrentLc := tmp[0] == unicode.ToLower(tmp[0])
	isNextLc := false

	for i := range tmp {
		lc := unicode.ToLower(tmp[i])

		if i == len(name)-1 {
			isNextLc = false
		} else {
			isNextLc = (tmp[i+1] == unicode.ToLower(tmp[i+1]))
		}

		if i != 0 && !isCurrentLc && (isPrevLc || isNextLc) {
			ret += "_"
		}
		ret += string(lc)

		isPrevLc = isCurrentLc
		isCurrentLc = isNextLc
	}
	return ret
}

// toScreamingCapsConstant
func toScreamingCapsConstant(name string) string {
	return strings.ToUpper(toFileName(name))
}

func toFieldName(name string) string {
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
