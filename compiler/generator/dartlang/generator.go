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
	minimumDartVersion    = "1.12.0"
	tab                   = "  "
	tabtab                = tab + tab
	tabtabtab             = tab + tab + tab
	tabtabtabtab          = tab + tab + tab + tab
	tabtabtabtabtab       = tab + tab + tab + tab + tab
	tabtabtabtabtabtab    = tab + tab + tab + tab + tab + tab
	tabtabtabtabtabtabtab = tab + tab + tab + tab + tab + tab + tab
)

// Generator implements the LanguageGenerator interface for Dart.
type Generator struct {
	*generator.BaseGenerator
	outputDir     string
	genWithFrugal bool
}

// NewGenerator creates a new Dart LanguageGenerator.
func NewGenerator(options map[string]string, genWithFrugal bool) generator.LanguageGenerator {
	return &Generator{&generator.BaseGenerator{Options: options}, "", genWithFrugal}
}

func (g *Generator) getLibraryName() string {
	if ns := g.Frugal.Thrift.Namespace(lang); ns != nil {
		return generator.LowercaseFirstLetter(toLibraryName(ns.Value))
	}
	return generator.LowercaseFirstLetter(g.Frugal.Name)
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

	if len(g.Frugal.Thrift.Constants) > 0 {
		constantsName := fmt.Sprintf("%sConstants", snakeToCamel(libraryName))
		contents += g.createExport(constantsName)
	}
	for _, s := range g.Frugal.Thrift.Structs {
		contents += g.createExport(s.Name)
	}
	for _, union := range g.Frugal.Thrift.Unions {
		contents += g.createExport(union.Name)
	}
	for _, exception := range g.Frugal.Thrift.Exceptions {
		contents += g.createExport(exception.Name)
	}
	for _, enum := range g.Frugal.Thrift.Enums {
		contents += g.createExport(enum.Name)
	}

	_, err = file.WriteString(contents)
	return err
}

func (g *Generator) createExport(structName string) string {
	srcDir := "src"
	if _, ok := g.Options["library_prefix"]; ok {
		srcDir = g.getLibraryName()
	}
	return fmt.Sprintf("export '%s/f_%s.dart' show %s;\n",
		srcDir, toFileName(structName), structName)
}

// TeardownGenerator is run after generation.
func (g *Generator) TeardownGenerator() error { return nil }

// GetOutputDir returns the output directory for generated files.
func (g *Generator) GetOutputDir(dir string) string {
	if namespace := g.Frugal.Thrift.Namespace(lang); namespace != nil {
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
	if _, ok := g.Options["library_prefix"]; !ok {
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
		"thrift": dep{
			Hosted:  hostedDep{Name: "thrift", URL: "https://pub.workiva.org"},
			Version: "^0.0.6",
		},
	}

	if g.Frugal.ContainsFrugalDefinitions() {
		deps["frugal"] = dep{
			Hosted:  hostedDep{Name: "frugal", URL: "https://pub.workiva.org"},
			Version: fmt.Sprintf("^%s", globals.Version),
		}
	}

	includesSet := make(map[string]bool)
	scopeIncludes, err := g.Frugal.ReferencedScopeIncludes()
	if err != nil {
		return err
	}
	for _, include := range scopeIncludes {
		includesSet[include.Name] = true
	}
	servIncludes, err := g.Frugal.ReferencedServiceIncludes()
	if err != nil {
		return err
	}
	for _, include := range servIncludes {
		includesSet[include.Name] = true
	}
	includes := make([]string, 0, len(includesSet))
	for include := range includesSet {
		includes = append(includes, include)
	}
	sort.Strings(includes)

	for _, include := range includes {
		name := include
		if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
			name = namespace.Value
		}
		deps[toLibraryName(name)] = dep{Path: "../" + toLibraryName(name)}
	}

	namespace := g.Frugal.Thrift.Namespace(lang)
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

	if _, ok := g.Options["library_prefix"]; ok {
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
	for _, service := range g.Frugal.Thrift.Services {
		servSrcDir := "src"
		if _, ok := g.Options["library_prefix"]; ok {
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
		if _, ok := g.Options["library_prefix"]; ok {
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
	if _, ok := g.Options["library_prefix"]; !ok {
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
		if err = g.GenerateObjectPackage(file, toFileName(name)); err != nil {
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

	if _, err = file.WriteString("\n\n"); err != nil {
		return err
	}
	if _, err = file.WriteString(g.GenerateThriftImports()); err != nil {
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
	// another constant. Need to recurse to get that value.
	identifier, ok := value.(parser.Identifier)
	// TODO this is essentially the same as the go generator, but I'm not
	// sure the best way to consolidate, especially with possibly different
	// generators to come. Maybe a post parse step to differentiate identifiers
	// somehow? Or extract a method to a super class?
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
		var s *parser.Struct
		for _, potential := range g.Frugal.Thrift.Structs {
			if underlyingType.Name == potential.Name {
				s = potential
				break
			}
		}

		contents := ""

		dartType := g.getDartTypeFromThriftType(parser.TypeFromStruct(s))
		contents += fmt.Sprintf("new %s()", dartType)
		for _, pair := range value.([]parser.KeyValue) {
			name := pair.Key.(string)
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
	file.WriteString(contents)

	return nil
}

// GenerateStruct generates the given struct.
func (g *Generator) GenerateStruct(s *parser.Struct) error {
	file, err := g.GenerateFile(s.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if _, err = file.WriteString("\n\n"); err != nil {
		return err
	}
	if _, err = file.WriteString(g.GenerateThriftImports()); err != nil {
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

// GenerateServiceArgsResults generates the args and results objects for the
// given service.
func (g *Generator) GenerateServiceArgsResults(serviceName string, outputDir string, structs []*parser.Struct) error {
	contents := ""
	contents += g.GenerateThriftImports()

	for _, s := range structs {
		contents += g.generateStruct(s)
	}

	file, err := g.GenerateFile(serviceName+"_structs", g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.WriteString(contents)
	return err
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
	contents += "implements TBase {\n"

	// Struct and field descriptors
	contents += fmt.Sprintf(tab+"static final TStruct _STRUCT_DESC = new TStruct(\"%s\");\n", s.Name)
	for _, field := range s.Fields {
		constantName := toScreamingCapsConstant(field.Name)
		contents += fmt.Sprintf(tab+"static final TField _%s_FIELD_DESC = new TField(\"%s\", %s, %d);\n",
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
	// TODO 2.0 don't use __isset helpers as primitives are nullable in dart
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
		contents += fmt.Sprintf(tabtabtabtabtab+"this.%s = value;\n", fName)
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
	contents := tab + "read(TProtocol iprot) {\n"
	contents += tabtab + "TField field;\n"
	contents += tabtab + "iprot.readStructBegin();\n"
	contents += tabtab + "while(true) {\n"
	contents += tabtabtab + "field = iprot.readFieldBegin();\n"
	contents += tabtabtab + "if(field.type == TType.STOP) {\n"
	contents += tabtabtabtab + "break;\n"
	contents += tabtabtab + "}\n"
	contents += tabtabtab + "switch(field.id) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtabtab+"case %s:\n", strings.ToUpper(field.Name))
		t := g.getEnumFromThriftType(g.Frugal.UnderlyingType(field.Type))
		contents += fmt.Sprintf(tabtabtabtabtab+"if(field.type == %s) {\n", t)
		contents += g.generateReadFieldRec(field, true, "")
		contents += tabtabtabtabtab + "} else {\n"
		contents += tabtabtabtabtabtab + "TProtocolUtil.skip(iprot, field.type);\n"
		contents += tabtabtabtabtab + "}\n"
		contents += tabtabtabtabtab + "break;\n"
	}
	contents += tabtabtabtab + "default:\n"
	contents += tabtabtabtabtab + "TProtocolUtil.skip(iprot, field.type);\n"
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
			contents += fmt.Sprintf(tabtabtab+"throw new TProtocolError(TProtocolErrorType.UNKNOWN, \"Required field '%s' was not present in struct %s\");\n", fName, s.Name)
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
	isEnum := g.Frugal.IsEnum(underlyingType)
	if underlyingType.IsPrimitive() || isEnum {
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
			if isEnum {
				thriftType = "I32"
			} else {
				panic("unkown thrift type: " + underlyingType.Name)
			}
		}

		contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"%s%s = iprot.read%s();\n", prefix, fName, thriftType)
		if primitive && first {
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"this.__isset_%s = true;\n", fName)
		}
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"%s%s = new %s();\n", prefix, fName, dartType)
		contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"%s.read(iprot);\n", fName)
	} else if underlyingType.IsContainer() {
		containerElem := getElem()
		valElem := getElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)
		valContents := g.generateReadFieldRec(valField, false, ind+tab)
		counterElem := getElem()
		dartType := g.getDartTypeFromThriftType(underlyingType)

		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"TList %s = iprot.readListBegin();\n", containerElem)
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"%s%s = new %s();\n", prefix, fName, dartType)
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"for(int %s = 0; %s < %s.length; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += fmt.Sprintf(tabtabtabtabtabtabtab+ind+"%s.add(%s);\n", fName, valElem)
			contents += tabtabtabtabtabtab + ind + "}\n"
			contents += tabtabtabtabtabtab + ind + "iprot.readListEnd();\n"
		case "set":
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"TSet %s = iprot.readSetBegin();\n", containerElem)
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"%s%s = new %s();\n", prefix, fName, dartType)
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"for(int %s = 0; %s < %s.length; ++%s) {\n",
				counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += fmt.Sprintf(tabtabtabtabtabtabtab+ind+"%s.add(%s);\n", fName, valElem)
			contents += tabtabtabtabtabtab + ind + "}\n"
			contents += tabtabtabtabtabtab + ind + "iprot.readSetEnd();\n"
		case "map":
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"TMap %s = iprot.readMapBegin();\n", containerElem)
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"%s%s = new %s();\n", prefix, fName, dartType)
			contents += fmt.Sprintf(tabtabtabtabtabtab+ind+"for(int %s = 0; %s < %s.length; ++%s) {\n",
				counterElem, counterElem, containerElem, counterElem)
			keyElem := getElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			contents += g.generateReadFieldRec(keyField, false, ind+tab)
			contents += valContents
			contents += fmt.Sprintf(tabtabtabtabtabtabtab+ind+"%s[%s] = %s;\n", fName, keyElem, valElem)
			contents += tabtabtabtabtabtab + ind + "}\n"
			contents += tabtabtabtabtabtab + ind + "iprot.readMapEnd();\n"
		default:
			panic("unrecognized thrift type: " + underlyingType.Name)
		}
	}

	return contents
}

func (g *Generator) generateWrite(s *parser.Struct) string {
	contents := tab + "write(TProtocol oprot) {\n"
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
	isEnum := g.Frugal.IsEnum(underlyingType)
	if underlyingType.IsPrimitive() || isEnum {
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
			if isEnum {
				write += "I32(%s);\n"
			} else {
				panic("unknown thrift type: " + underlyingType.Name)
			}
		}

		contents += fmt.Sprintf(write, fName)
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(tabtab+ind+"%s.write(oprot);\n", fName)
	} else if underlyingType.IsContainer() {
		valEnumType := g.getEnumFromThriftType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			valElem := getElem()
			valField := parser.FieldFromType(underlyingType.ValueType, valElem)
			contents += fmt.Sprintf(tabtab+ind+"oprot.writeListBegin(new TList(%s, %s.length));\n", valEnumType, fName)
			contents += fmt.Sprintf(tabtab+ind+"for(var %s in %s) {\n", valElem, fName)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += tabtab + ind + "}\n"
			contents += tabtab + ind + "oprot.writeListEnd();\n"
		case "set":
			valElem := getElem()
			valField := parser.FieldFromType(underlyingType.ValueType, valElem)
			contents += fmt.Sprintf(tabtab+ind+"oprot.writeSetBegin(new TSet(%s, %s.length));\n", valEnumType, fName)
			contents += fmt.Sprintf(tabtab+ind+"for(var %s in %s) {\n", valElem, fName)
			contents += g.generateWriteFieldRec(valField, false, ind+tab)
			contents += tabtab + ind + "}\n"
			contents += tabtab + ind + "oprot.writeSetEnd();\n"
		case "map":
			keyEnumType := g.getEnumFromThriftType(underlyingType.KeyType)
			keyElem := getElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			valField := parser.FieldFromType(underlyingType.ValueType, fmt.Sprintf("%s[%s]", fName, keyElem))
			contents += fmt.Sprintf(tabtab+ind+"oprot.writeMapBegin(new TMap(%s, %s, %s.length));\n", keyEnumType, valEnumType, fName)
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
	contents += tabtab + "// check for required fields\n"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required {
			fName := toFieldName(field.Name)
			if !g.isDartPrimitive(field.Type) {
				contents += fmt.Sprintf(tabtab+"if(%s == null) {\n", fName)
				contents += fmt.Sprintf(tabtabtab+"throw new TProtocolError(TProtocolErrorType.UNKNOWN, \"Required field '%s' was not present in struct %s\");\n", fName, s.Name)
				contents += tabtab + "}\n"
			}
		}
	}

	contents += tabtab + "// check that fields of type enum have valid values\n"
	for _, field := range s.Fields {
		if g.Frugal.IsEnum(field.Type) {
			fName := toFieldName(field.Name)
			isSetCheck := fmt.Sprintf("isSet%s()", strings.Title(field.Name))
			contents += fmt.Sprintf(tabtab+"if(%s && !%s.VALID_VALUES.contains(%s)) {\n",
				isSetCheck, g.qualifiedTypeName(field.Type), fName)
			contents += fmt.Sprintf(tabtabtab+"throw new TProtocolError(TProtocolErrorType.UNKNOWN, \"The field '%s' has been assigned the invalid value $%s\");\n", fName, fName)
			contents += tabtab + "}\n"
		}
	}

	contents += tab + "}\n"
	return contents
}

// GenerateServicePackage generates the package for the given service.
func (g *Generator) GenerateServicePackage(file *os.File, s *parser.Service) error {
	// TODO 2.0: change to serviceSuffix
	return g.generatePackage(file, s.Name, scopeSuffix)
}

// GenerateScopePackage generates the package for the given scope.
func (g *Generator) GenerateScopePackage(file *os.File, s *parser.Scope) error {
	return g.generatePackage(file, s.Name, scopeSuffix)
}

// GenerateObjectPackage generates the package for the given name.
func (g *Generator) GenerateObjectPackage(file *os.File, name string) error {
	return g.generatePackage(file, name, "")
}

func (g *Generator) generatePackage(file *os.File, name, suffix string) error {
	pkg := g.getLibraryName()

	libraryPrefix := g.getLibraryPrefix()
	libraryDeclaration := "library " + libraryPrefix + pkg
	if libraryPrefix == "" {
		libraryDeclaration += ".src"
	}

	_, err := file.WriteString(fmt.Sprintf("%s.%s%s%s;", libraryDeclaration,
		generator.FilePrefix, strings.ToLower(name), suffix))
	return err
}

// GenerateThriftImports generates necessary imports for Thrift.
func (g *Generator) GenerateThriftImports() string {
	imports := "import 'dart:typed_data' show Uint8List;\n"
	imports += "import 'package:thrift/thrift.dart';\n"
	// Import the current package
	imports += g.getImportDeclaration(g.getNamespaceOrName())

	// Import includes
	for _, include := range g.Frugal.Thrift.Includes {
		name := include.Name
		if namespace := g.Frugal.NamespaceForInclude(filepath.Base(include.Name), lang); namespace != nil {
			name = namespace.Value
		}

		imports += g.getImportDeclaration(name)
	}

	imports += "\n"
	return imports
}

// GenerateServiceImports generates necessary imports for the given service.
func (g *Generator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	imports := "import 'dart:async';\n\n"
	imports += "import 'dart:typed_data' show Uint8List;\n"
	imports += "import 'package:thrift/thrift.dart' as thrift;\n"
	imports += "import 'package:frugal/frugal.dart' as frugal;\n\n"
	// import included packages
	includes, err := s.ReferencedIncludes()
	if err != nil {
		return err
	}
	for _, include := range includes {
		name := include.Name
		if namespace := g.Frugal.NamespaceForInclude(include.Name, lang); namespace != nil {
			name = namespace.Value
		}

		imports += g.getImportDeclaration(name)
	}

	// Import same package.
	imports += g.getImportDeclaration(g.getNamespaceOrName())

	// Import thrift package for method args
	servSnake := toFileName(s.Name)
	if g.genWithFrugal {
		imports += fmt.Sprintf("import 'f_%s_structs.dart' as t_%s_file;\n", servSnake, servSnake)
	} else {
		imports += fmt.Sprintf("import '%s.dart' as t_%s_file;\n", servSnake, servSnake)
	}

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
		name := include.Name
		if namespace := s.Frugal.NamespaceForInclude(include.Name, lang); namespace != nil {
			name = namespace.Value
		}

		imports += g.getImportDeclaration(name)
	}

	// Import same package.
	imports += g.getImportDeclaration(g.getNamespaceOrName())

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
	publishers += tab + "frugal.FScopeTransport fTransport;\n"
	publishers += tab + "frugal.FProtocol fProtocol;\n"
	publishers += tab + "Map<String, frugal.FMethod> _methods;\n"
	publishers += tab + "frugal.Lock _writeLock;\n\n"

	publishers += fmt.Sprintf(tab+"%sPublisher(frugal.FScopeProvider provider, [List<frugal.Middleware> middleware]) {\n", strings.Title(scope.Name))
	publishers += tabtab + "fTransport = provider.fTransportFactory.getTransport();\n"
	publishers += tabtab + "fProtocol = provider.fProtocolFactory.getProtocol(fTransport);\n"
	publishers += tabtab + "_writeLock = new frugal.Lock();\n"
	publishers += tabtab + "this._methods = {};\n"
	for _, operation := range scope.Operations {
		publishers += fmt.Sprintf(tabtab+"this._methods['%s'] = new frugal.FMethod(this._publish%s, '%s', 'publish%s', middleware);\n",
			operation.Name, operation.Name, strings.Title(scope.Name), operation.Name)
	}
	publishers += tab + "}\n\n"

	publishers += tab + "Future open() {\n"
	publishers += tabtab + "return fTransport.open();\n"
	publishers += tab + "}\n\n"

	publishers += tab + "Future close() {\n"
	publishers += tabtab + "return fTransport.close();\n"
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
		publishers += fmt.Sprintf(tab+"Future publish%s(frugal.FContext ctx, %s%s req) {\n", op.Name, args, g.qualifiedTypeName(op.Type))
		publishers += fmt.Sprintf(tabtab+"return this._methods['%s']([ctx, %sreq]);\n", op.Name, argsWithoutTypes)
		publishers += tab + "}\n\n"

		publishers += fmt.Sprintf(tab+"Future _publish%s(frugal.FContext ctx, %s%s req) async {\n", op.Name, args, g.qualifiedTypeName(op.Type))

		publishers += tabtab + "await _writeLock.lock();\n"
		publishers += tabtab + "try {\n"
		publishers += fmt.Sprintf(tabtabtab+"var op = \"%s\";\n", op.Name)
		publishers += fmt.Sprintf(tabtabtab+"var prefix = \"%s\";\n", generatePrefixStringTemplate(scope))
		publishers += tabtabtab + "var topic = \"${prefix}" + strings.Title(scope.Name) + "${delimiter}${op}\";\n"
		publishers += tabtabtab + "fTransport.setTopic(topic);\n"
		publishers += tabtabtab + "var oprot = fProtocol;\n"
		publishers += tabtabtab + "var msg = new thrift.TMessage(op, thrift.TMessageType.CALL, 0);\n"
		publishers += tabtabtab + "oprot.writeRequestHeader(ctx);\n"
		publishers += tabtabtab + "oprot.writeMessageBegin(msg);\n"
		publishers += tabtabtab + "req.write(oprot);\n"
		publishers += tabtabtab + "oprot.writeMessageEnd();\n"
		publishers += tabtabtab + "await oprot.transport.flush();\n"
		publishers += tabtab + "} finally {\n"
		publishers += tabtabtab + "_writeLock.unlock();\n"
		publishers += tabtab + "}\n"
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

	subscribers += fmt.Sprintf(tab+"%sSubscriber(this.provider, [this._middleware]) {}\n\n", strings.Title(scope.Name))

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
			op.Name, args, op.Type.ParamName(), g.qualifiedTypeName(op.Type))
		subscribers += fmt.Sprintf(tabtab+"var op = \"%s\";\n", op.Name)
		subscribers += fmt.Sprintf(tabtab+"var prefix = \"%s\";\n", generatePrefixStringTemplate(scope))
		subscribers += tabtab + "var topic = \"${prefix}" + strings.Title(scope.Name) + "${delimiter}${op}\";\n"
		subscribers += tabtab + "var transport = provider.fTransportFactory.getTransport();\n"
		subscribers += fmt.Sprintf(tabtab+"await transport.subscribe(topic, _recv%s(op, provider.fProtocolFactory, on%s));\n",
			op.Name, op.Type.ParamName())
		subscribers += tabtab + "return new frugal.FSubscription(topic, transport);\n"
		subscribers += tab + "}\n\n"

		subscribers += fmt.Sprintf(tab+"_recv%s(String op, frugal.FProtocolFactory protocolFactory, dynamic on%s(frugal.FContext ctx, %s req)) {\n",
			op.Name, op.Type.ParamName(), g.qualifiedTypeName(op.Type))
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
		subscribers += tabtabtabtab + "thrift.TApplicationErrorType.UNKNOWN_METHOD, tMsg.name);\n"
		subscribers += tabtabtab + "}\n"
		subscribers += fmt.Sprintf(tabtabtab+"var req = new %s();\n", g.qualifiedTypeName(op.Type))
		subscribers += tabtabtab + "req.read(iprot);\n"
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
		contents += fmt.Sprintf(tab+"Future%s %s(frugal.FContext ctx%s);\n",
			g.generateReturnArg(method), generator.LowercaseFirstLetter(method.Name), g.generateInputArgs(method.Arguments))
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
	contents += tab + "Map<String, frugal.FMethod> _methods;\n\n"
	if service.Extends != "" {
		contents += fmt.Sprintf(tab+"F%sClient(frugal.FTransport transport, frugal.FProtocolFactory protocolFactory, [List<frugal.Middleware> middleware])\n", servTitle)
		contents += tabtabtab + ": super(transport, protocolFactory) {\n"
	} else {
		contents += fmt.Sprintf(tab+"F%sClient(frugal.FTransport transport, frugal.FProtocolFactory protocolFactory, [List<frugal.Middleware> middleware]) {\n", servTitle)
	}
	contents += tabtab + "_transport = transport;\n"
	contents += tabtab + "_transport.setRegistry(new frugal.FClientRegistry());\n"
	contents += tabtab + "_protocolFactory = protocolFactory;\n"
	contents += tabtab + "_oprot = _protocolFactory.getProtocol(_transport);\n\n"
	if service.Extends == "" {
		contents += tabtab + "writeLock = new frugal.Lock();\n"
	}
	contents += tabtab + "this._methods = {};\n"
	for _, method := range service.Methods {
		nameLower := generator.LowercaseFirstLetter(method.Name)
		contents += fmt.Sprintf(tabtab+"this._methods['%s'] = new frugal.FMethod(this._%s, '%s', '%s', middleware);\n",
			nameLower, nameLower, servTitle, nameLower)
	}
	contents += tab + "}\n\n"
	contents += tab + "frugal.FTransport _transport;\n"
	contents += tab + "frugal.FProtocolFactory _protocolFactory;\n"
	contents += tab + "frugal.FProtocol _oprot;\n"
	contents += tab + "frugal.FProtocol get oprot => _oprot;\n"
	if service.Extends == "" {
		contents += tab + "frugal.Lock writeLock;\n"
	}
	contents += "\n"

	for _, method := range service.Methods {
		contents += g.generateClientMethod(service, method)
	}
	contents += "}\n"
	return contents
}

func (g *Generator) generateClientMethod(service *parser.Service, method *parser.Method) string {
	servSnake := toFileName(service.Name)
	nameTitle := strings.Title(method.Name)
	nameLower := generator.LowercaseFirstLetter(method.Name)

	contents := ""
	if method.Comment != nil {
		contents += g.GenerateInlineComment(method.Comment, tab+"/")
	}
	// Generate wrapper method
	contents += fmt.Sprintf(tab+"Future%s %s(frugal.FContext ctx%s) {\n",
		g.generateReturnArg(method), nameLower, g.generateInputArgs(method.Arguments))
	contents += fmt.Sprintf(tabtab+"return this._methods['%s']([ctx%s]);\n", nameLower, g.generateInputArgsWithoutTypes(method.Arguments))
	contents += fmt.Sprintf(tab + "}\n\n")

	// Generate the calling method
	contents += fmt.Sprintf(tab+"Future%s _%s(frugal.FContext ctx%s) async {\n",
		g.generateReturnArg(method), nameLower, g.generateInputArgs(method.Arguments))

	// No need to register for oneway
	indent := tabtab
	if !method.Oneway {
		contents += tabtab + "var controller = new StreamController();\n"
		contents += tabtab + "var closeSubscription = _transport.onClose.listen((_) {\n"
		contents += tabtabtab + "controller.addError(new thrift.TTransportError(\n"
		contents += tabtabtabtab + "thrift.TTransportErrorType.NOT_OPEN,\n"
		contents += tabtabtabtab + "\"Transport closed before request completed.\"));\n"
		contents += tabtabtab + "});\n"
		contents += fmt.Sprintf(tabtab+"_transport.register(ctx, _recv%sHandler(ctx, controller));\n", nameTitle)
	}
	contents += indent + "await writeLock.lock();\n"

	if !method.Oneway {
		contents += tabtab + "try {\n"
		indent = tabtabtab
	}
	contents += indent + "try {\n"

	contents += indent + tab + "oprot.writeRequestHeader(ctx);\n"
	msgType := "CALL"
	if method.Oneway {
		msgType = "ONEWAY"
	}
	contents += fmt.Sprintf(indent+tab+"oprot.writeMessageBegin(new thrift.TMessage(\"%s\", thrift.TMessageType.%s, 0));\n",
		nameLower, msgType)
	contents += fmt.Sprintf(indent+tab+"t_%s_file.%s_args args = new t_%s_file.%s_args();\n",
		servSnake, nameLower, servSnake, nameLower)
	for _, arg := range method.Arguments {
		argLower := generator.LowercaseFirstLetter(arg.Name)
		contents += fmt.Sprintf(indent+tab+"args.%s = %s;\n", argLower, argLower)
	}
	contents += indent + tab + "args.write(oprot);\n"
	contents += indent + tab + "oprot.writeMessageEnd();\n"
	contents += indent + tab + "await oprot.transport.flush();\n"

	contents += indent + "} finally {\n"
	contents += indent + tab + "writeLock.unlock();\n"
	contents += indent + "}\n"
	// Nothing more to do for oneway
	if method.Oneway {
		contents += tab + "}\n\n"
		return contents
	}

	contents += "\n"

	// TODO 2.0.0: Dart TimeoutException should be wrapped in an FTimeoutException.
	// This should happen in a major release since it's an API change.
	contents += tabtabtab + "return await controller.stream.first.timeout(ctx.timeout);\n"
	contents += tabtab + "} finally {\n"
	contents += tabtabtab + "closeSubscription.cancel();\n"
	contents += tabtabtab + "_transport.unregister(ctx);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"
	if method.Oneway {
		return contents
	}

	// Generate the callback
	contents += fmt.Sprintf(tab+"_recv%sHandler(frugal.FContext ctx, StreamController controller) {\n", nameTitle)
	contents += fmt.Sprintf(tabtab+"%sCallback(thrift.TTransport transport) {\n", nameLower)
	contents += tabtabtab + "try {\n"
	contents += tabtabtabtab + "var iprot = _protocolFactory.getProtocol(transport);\n"
	contents += tabtabtabtab + "iprot.readResponseHeader(ctx);\n"
	contents += tabtabtabtab + "thrift.TMessage msg = iprot.readMessageBegin();\n"
	contents += tabtabtabtab + "if (msg.type == thrift.TMessageType.EXCEPTION) {\n"
	contents += tabtabtabtabtab + "thrift.TApplicationError error = thrift.TApplicationError.read(iprot);\n"
	contents += tabtabtabtabtab + "iprot.readMessageEnd();\n"
	contents += tabtabtabtabtab + "if (error.type == frugal.FTransport.RESPONSE_TOO_LARGE) {\n"
	contents += tabtabtabtabtabtab + "controller.addError(new frugal.FMessageSizeError.response());\n"
	contents += tabtabtabtabtabtab + "return;\n"
	contents += tabtabtabtabtab + "}\n"
	contents += tabtabtabtabtab + "if (error.type == frugal.FRateLimitError.RATE_LIMIT_EXCEEDED) {\n"
	contents += tabtabtabtabtabtab + "controller.addError(new frugal.FRateLimitError());\n"
	contents += tabtabtabtabtabtab + "return;\n"
	contents += tabtabtabtabtab + "}\n"
	contents += tabtabtabtabtab + "throw error;\n"
	contents += tabtabtabtab + "}\n\n"

	contents += fmt.Sprintf(tabtabtabtab+"t_%s_file.%s_result result = new t_%s_file.%s_result();\n",
		servSnake, nameLower, servSnake, nameLower)
	contents += tabtabtabtab + "result.read(iprot);\n"
	contents += tabtabtabtab + "iprot.readMessageEnd();\n"
	if method.ReturnType == nil {
		contents += g.generateErrors(method)
		contents += tabtabtabtab + "controller.add(null);\n"
	} else {
		contents += tabtabtabtab + "if (result.isSetSuccess()) {\n"
		contents += tabtabtabtabtab + "controller.add(result.success);\n"
		contents += tabtabtabtabtab + "return;\n"
		contents += tabtabtabtab + "}\n\n"
		contents += g.generateErrors(method)
		contents += tabtabtabtab + "throw new thrift.TApplicationError(\n"
		contents += fmt.Sprintf(tabtabtabtabtab+"thrift.TApplicationErrorType.MISSING_RESULT, "+
			"\"%s failed: unknown result\"\n",
			nameLower)
		contents += tabtabtabtab + ");\n"
	}
	contents += tabtabtab + "} catch(e) {\n"
	contents += tabtabtabtab + "controller.addError(e);\n"
	contents += tabtabtabtab + "rethrow;\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "}\n"
	contents += fmt.Sprintf(tabtab+"return %sCallback;\n", nameLower)
	contents += tab + "}\n\n"

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
		argStr += ", " + g.getDartTypeFromThriftType(arg.Type) + " " + generator.LowercaseFirstLetter(arg.Name)
	}
	return argStr
}

func (g *Generator) generateInputArgsWithoutTypes(args []*parser.Field) string {
	argStr := ""
	for _, arg := range args {
		argStr += ", " + generator.LowercaseFirstLetter(arg.Name)
	}
	return argStr
}

func (g *Generator) generateErrors(method *parser.Method) string {
	contents := ""
	for _, exp := range method.Exceptions {
		contents += fmt.Sprintf(tabtabtabtab+"if (result.%s != null) {\n", generator.LowercaseFirstLetter(exp.Name))
		contents += fmt.Sprintf(tabtabtabtabtab+"controller.addError(result.%s);\n", generator.LowercaseFirstLetter(exp.Name))
		contents += tabtabtabtabtab + "return;\n"
		contents += tabtabtabtab + "}\n"
	}
	return contents
}

func (g *Generator) getEnumFromThriftType(t *parser.Type) string {
	underlyingType := g.Frugal.UnderlyingType(t)
	switch underlyingType.Name {
	case "bool":
		return "TType.BOOL"
	case "byte", "i8":
		return "TType.BYTE"
	case "i16":
		return "TType.I16"
	case "i32":
		return "TType.I32"
	case "i64":
		return "TType.I64"
	case "double":
		return "TType.DOUBLE"
	case "string", "binary":
		return "TType.STRING"
	case "list":
		return "TType.LIST"
	case "set":
		return "TType.SET"
	case "map":
		return "TType.MAP"
	default:
		if g.Frugal.IsEnum(underlyingType) {
			return "TType.I32"
		} else if g.Frugal.IsStruct(underlyingType) {
			return "TType.STRUCT"
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
		name = toLibraryName(name)
		param = fmt.Sprintf("t_%s.%s", name, param)
	} else {
		param = fmt.Sprintf("t_%s.%s", toLibraryName(g.getNamespaceOrName()), param)
	}
	return param
}

func (g *Generator) getLibraryPrefix() string {
	prefix := ""
	if _, ok := g.Options["library_prefix"]; ok {
		prefix += g.Options["library_prefix"]
		if !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
	}
	return prefix
}

func (g *Generator) getPackagePrefix() string {
	prefix := ""
	if _, ok := g.Options["library_prefix"]; ok {
		prefix += strings.Replace(g.getLibraryPrefix(), ".", "/", -1)
	}
	return prefix
}

func (g *Generator) getImportDeclaration(namespace string) string {
	namespace = toLibraryName(filepath.Base(namespace))
	prefix := g.getPackagePrefix()
	if prefix == "" {
		prefix += namespace + "/"
	}
	return fmt.Sprintf("import 'package:%s%s.dart' as t_%s;\n", prefix, namespace, namespace)
}

func (g *Generator) getNamespaceOrName() string {
	namespace := g.Frugal.Thrift.Namespace(lang)
	if namespace != nil {
		return namespace.Value
	}
	return g.Frugal.Name
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

var elemNum int

// getElem returns a unique identifier name
func getElem() string {
	s := fmt.Sprintf("elem%d", elemNum)
	elemNum++
	return s
}
