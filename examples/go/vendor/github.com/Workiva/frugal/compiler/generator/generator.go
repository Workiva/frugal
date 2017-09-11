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

package generator

import (
	"os"
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

// FilePrefix is the default prefix for generated files.
const FilePrefix = "f_"

// FileType represents a generated file type.
type FileType string

// Valid FileTypes.
const (
	CombinedServiceFile FileType = "combined_service"
	CombinedScopeFile   FileType = "combined_scope"
	PublishFile         FileType = "publish"
	SubscribeFile       FileType = "subscribe"

	TypeFile               FileType = "types"
	ServiceArgsResultsFile FileType = "service_args_results"
	ObjectFile             FileType = "object"
)

// Options contains language generator options. The map key is the option name,
// and the value is the option description.
type Options map[string]string

// LanguageOptions contains a map of language to generator options.
type LanguageOptions map[string]Options

// Languages is a map of supported language to a map of the generator options
// it supports.
var Languages = LanguageOptions{
	"go": Options{
		"thrift_import":  "Override Thrift package import path (default: git.apache.org/thrift.git/lib/go/thrift)",
		"frugal_import":  "Override Frugal package import path (default: github.com/Workiva/frugal/lib/go)",
		"package_prefix": "Package prefix for generated files",
		"async":          "Generate async client code using channels",
		"use_vendor":     "Use specified import references for vendored includes and do not generate code for them",
	},
	"java": Options{
		"generated_annotations": "[undated|suppress] " +
			"undated: suppress the date at @Generated annotations, " +
			"suppress: suppress @Generated annotations entirely",
		"async":            "Generate async client code using futures",
		"boxed_primitives": "Generate primitives as the boxed equivalents",
	},
	"dart": Options{
		"library_prefix": "Generate code that can be used within an existing library. " +
			"Use a dot-separated string, e.g. \"my_parent_lib.src.gen\"",
		"use_enums":  "Generate enums as enums rather than a class with numerical constants",
		"use_vendor": "Use specified import references for vendored includes and do not generate code for them",
	},
	"py": Options{
		"tornado":        "Generate code for use with Tornado (compatible with Python 2.7)",
		"asyncio":        "Generate code for use with asyncio (compatible with Python 3.5 or above)",
		"package_prefix": "Package prefix for generated files",
	},
	"html": Options{
		"standalone": "Self-contained mode, includes all CSS in the HTML files. Generates no style.css file, but HTML files will be larger",
	},
}

// ValidateOption indicates if the language option is supported for the given
// language.
func ValidateOption(lang, option string) bool {
	options, ok := Languages[lang]
	if !ok {
		return false
	}
	_, ok = options[option]
	return ok
}

// ProgramGenerator generates source code in a specified language for a Frugal
// produced by the parser.
type ProgramGenerator interface {
	// Generate the Frugal in the given directory.
	Generate(frugal *parser.Frugal, outputDir string) error

	// GetOutputDir returns the full output directory for generated code.
	GetOutputDir(dir string, f *parser.Frugal) string

	// DefaultOutputDir returns the default directory for generated code.
	DefaultOutputDir() string
}

// LanguageGenerator generates source code as implemented for specific
// languages.
type LanguageGenerator interface {
	// Generic methods
	SetFrugal(*parser.Frugal)
	SetupGenerator(outputDir string) error
	TeardownGenerator() error
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

// GetPackageComponents returns the package string split on dots.
func GetPackageComponents(pkg string) []string {
	return strings.Split(pkg, ".")
}

// programGenerator is an implementation of the ProgramGenerator interface
type programGenerator struct {
	LanguageGenerator
	splitPublisherSubscriber bool
}

// NewProgramGenerator creates a new ProgramGenerator using the given
// LanguageGenerator.
func NewProgramGenerator(generator LanguageGenerator, splitPublisherSubscriber bool) ProgramGenerator {
	return &programGenerator{generator, splitPublisherSubscriber}
}

// Generate the Frugal in the given directory.
func (o *programGenerator) Generate(frugal *parser.Frugal, outputDir string) error {
	o.SetFrugal(frugal)
	o.SetupGenerator(outputDir)

	if err := o.GenerateDependencies(outputDir); err != nil {
		return err
	}

	if err := o.GenerateConstantsContents(frugal.Constants); err != nil {
		return err
	}

	for _, typedef := range frugal.Typedefs {
		if err := o.GenerateTypeDef(typedef); err != nil {
			return err
		}
	}

	for _, enum := range frugal.Enums {
		if err := o.GenerateEnum(enum); err != nil {
			return err
		}
	}

	for _, s := range frugal.Structs {
		if err := o.GenerateStruct(s); err != nil {
			return err
		}
	}

	for _, union := range frugal.Unions {
		if err := o.GenerateUnion(union); err != nil {
			return err
		}
	}

	for _, exception := range frugal.Exceptions {
		if err := o.GenerateException(exception); err != nil {
			return err
		}
	}

	// Generate services
	for _, service := range frugal.Services {
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

	return o.TeardownGenerator()
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
