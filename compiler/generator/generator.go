package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

const FilePrefix = "f_"

type FileType string

const (
	CombinedServiceFile FileType = "combined_service"
	CombinedScopeFile   FileType = "combined_scope"
	PublishFile         FileType = "publish"
	SubscribeFile       FileType = "subscribe"

	TypeFile               FileType = "types"
	ServiceArgsResultsFile FileType = "service_args_results"
)

// Languages is a map of supported language to a slice of the generator options
// it supports.
var Languages = map[string][]string{
	"go":   []string{"thrift_import", "frugal_import", "package_prefix", "gen_with_frugal"},
	"java": []string{"generated_annotations"},
	"dart": []string{"library_prefix"},
}

// ProgramGenerator generates source code in a specified language for a Frugal
// produced by the parser.
type ProgramGenerator interface {
	// Generate the Frugal in the given directory.
	Generate(frugal *parser.Frugal, outputDir string, genWithFrugal bool) error

	// GetOutputDir returns the full output directory for generated code.
	GetOutputDir(dir string, f *parser.Frugal) string

	// DefaultOutputDir returns the default directory for generated code.
	DefaultOutputDir() string
}

// Generator generates source code as implemented for specific languages.
type LanguageGenerator interface {
	// Generic methods
	SetFrugal(*parser.Frugal)
	InitializeGenerator(outputDir string) error
	CloseGenerator() error
	GenerateDependencies(dir string) error
	GenerateFile(name, outputDir string, fileType FileType) (*os.File, error)
	GenerateDocStringComment(*os.File) error
	GenerateConstants(f *os.File, name string) error
	GenerateNewline(*os.File, int) error
	GetOutputDir(dir string) string
	DefaultOutputDir() string
	PostProcess(*os.File) error

	// Thrift stuff
	GenerateConstantsContents([]*parser.Constant) error
	GenerateTypeDef(*parser.TypeDef) error
	GenerateEnum(*parser.Enum) error
	GenerateStruct(*parser.Struct) error
	GenerateUnion(*parser.Struct) error
	GenerateException(*parser.Struct) error
	GenerateServiceArgsResults(string, string, []*parser.Struct) error

	// Service-specific methods
	GenerateServicePackage(*os.File, *parser.Service) error
	GenerateServiceImports(*os.File, *parser.Service) error
	GenerateService(*os.File, *parser.Service) error

	// Scope-specific methods
	GenerateScopePackage(*os.File, *parser.Scope) error
	GenerateScopeImports(*os.File, *parser.Scope) error
	GeneratePublisher(*os.File, *parser.Scope) error
	GenerateSubscriber(*os.File, *parser.Scope) error
}

func GetPackageComponents(pkg string) []string {
	return strings.Split(pkg, ".")
}

// programGenerator is an implementation of the ProgramGenerator interface
type programGenerator struct {
	LanguageGenerator
	splitPublisherSubscriber bool
}

func NewProgramGenerator(generator LanguageGenerator, splitPublisherSubscriber bool) ProgramGenerator {
	return &programGenerator{generator, splitPublisherSubscriber}
}

// Generate the Frugal in the given directory.
func (o *programGenerator) Generate(frugal *parser.Frugal, outputDir string, genWithFrugal bool) error {
	o.SetFrugal(frugal)
	if genWithFrugal {
		o.InitializeGenerator(outputDir)
	}
	if err := o.GenerateDependencies(outputDir); err != nil {
		return err
	}

	if genWithFrugal {
		// generate thrift
		if err := o.generateThrift(frugal, outputDir); err != nil {
			return err
		}
	}

	// generate frugal
	if err := o.generateFrugal(frugal, outputDir); err != nil {
		return err
	}

	if genWithFrugal {
		return o.CloseGenerator()
	}
	return nil
}

func (o *programGenerator) generateThrift(frugal *parser.Frugal, outputDir string) error {
	if err := o.GenerateConstantsContents(frugal.Thrift.Constants); err != nil {
		return err
	}

	for _, typedef := range frugal.Thrift.Typedefs {
		if err := o.GenerateTypeDef(typedef); err != nil {
			return err
		}
	}

	for _, enum := range frugal.Thrift.Enums {
		if err := o.GenerateEnum(enum); err != nil {
			return err
		}
	}

	for _, s := range frugal.Thrift.Structs {
		if err := o.GenerateStruct(s); err != nil {
			return err
		}
	}

	for _, union := range frugal.Thrift.Unions {
		if err := o.GenerateUnion(union); err != nil {
			return err
		}
	}

	for _, exception := range frugal.Thrift.Exceptions {
		if err := o.GenerateException(exception); err != nil {
			return err
		}
	}

	for _, service := range frugal.Thrift.Services {
		structs := []*parser.Struct{}
		for _, method := range service.Methods {
			arg := &parser.Struct{
				Name:   fmt.Sprintf("%s_%s_args", service.Name, method.Name),
				Fields: method.Arguments,
				Type:   parser.StructTypeStruct,
			}
			structs = append(structs, arg)

			if !method.Oneway {
				numReturns := 0
				if method.ReturnType != nil {
					numReturns = 1
				}

				fields := make([]*parser.Field, len(method.Exceptions)+numReturns, len(method.Exceptions)+numReturns)
				if numReturns == 1 {
					fields[0] = frugal.FieldFromType(method.ReturnType, "success")
				}
				copy(fields[numReturns:], method.Exceptions)
				for _, field := range fields {
					field.Modifier = parser.Optional
				}

				result := &parser.Struct{
					Name:   fmt.Sprintf("%s_%s_result", service.Name, method.Name),
					Fields: fields,
					Type:   parser.StructTypeStruct,
				}
				structs = append(structs, result)
			}
		}
		if err := o.GenerateServiceArgsResults(service.Name, outputDir, structs); err != nil {
			return err
		}
	}

	return nil
}

func (o *programGenerator) generateFrugal(frugal *parser.Frugal, outputDir string) error {
	// If no frugal definitions, we can return.
	if !frugal.ContainsFrugalDefinitions() {
		return nil
	}

	// Generate services
	for _, service := range frugal.Thrift.Services {
		if err := o.generateServiceFile(service, outputDir); err != nil {
			return err
		}
	}
	// Generate scopes
	for _, scope := range frugal.Scopes {
		if o.splitPublisherSubscriber {
			if err := o.generateScopeFile(scope, outputDir, PublishFile); err != nil {
				return err
			}
			if err := o.generateScopeFile(scope, outputDir, SubscribeFile); err != nil {
				return err
			}
		} else {
			if err := o.generateScopeFile(scope, outputDir, CombinedScopeFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *programGenerator) generateServiceFile(service *parser.Service, outputDir string) error {
	file, err := o.GenerateFile(service.Name, outputDir, CombinedServiceFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := o.GenerateDocStringComment(file); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateServicePackage(file, service); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateServiceImports(file, service); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateService(file, service); err != nil {
		return err
	}

	return o.PostProcess(file)
}

func (o *programGenerator) generateScopeFile(scope *parser.Scope, outputDir string, fileType FileType) error {
	file, err := o.GenerateFile(scope.Name, outputDir, fileType)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := o.GenerateDocStringComment(file); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateScopePackage(file, scope); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateScopeImports(file, scope); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateConstants(file, scope.Name); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if fileType == CombinedScopeFile || fileType == PublishFile {
		if err := o.GeneratePublisher(file, scope); err != nil {
			return err
		}
	}

	if fileType == CombinedScopeFile {
		if err := o.GenerateNewline(file, 2); err != nil {
			return err
		}
	}

	if fileType == CombinedScopeFile || fileType == SubscribeFile {
		if err := o.GenerateSubscriber(file, scope); err != nil {
			return err
		}
	}

	if err := o.GenerateNewline(file, 1); err != nil {
		return err
	}

	return o.PostProcess(file)
}

// GetOutputDir returns the full output directory for generated code.
func (o *programGenerator) GetOutputDir(dir string, f *parser.Frugal) string {
	o.LanguageGenerator.SetFrugal(f)
	return o.LanguageGenerator.GetOutputDir(dir)
}

// DefaultOutputDir returns the default directory for generated code.
func (o *programGenerator) DefaultOutputDir() string {
	return o.LanguageGenerator.DefaultOutputDir()
}
